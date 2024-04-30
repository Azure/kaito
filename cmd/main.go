// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package main

import (
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/azure/kaito/pkg/k8sclient"
	"github.com/azure/kaito/pkg/utils/consts"
	cliflag "k8s.io/component-base/cli/flag"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"

	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/azure/kaito/pkg/controllers"
	"github.com/azure/kaito/pkg/webhooks"
	"k8s.io/klog/v2"
	"knative.dev/pkg/injection/sharedmain"
	"knative.dev/pkg/signals"
	"knative.dev/pkg/webhook"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	kaitov1alpha1 "github.com/azure/kaito/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

const (
	WebhookServiceName = "WEBHOOK_SERVICE"
	WebhookServicePort = "WEBHOOK_PORT"
)

var (
	scheme = runtime.NewScheme()

	exitWithErrorFunc = func() {
		klog.Flush()
		os.Exit(1)
	}
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(kaitov1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1alpha5.SchemeBuilder.AddToScheme(scheme))
	utilruntime.Must(v1beta1.SchemeBuilder.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
	klog.InitFlags(nil)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var enableWebhook bool
	var probeAddr string
	var featureGates string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&enableWebhook, "webhook", true,
		"Enable webhook for controller manager. Default is true.")
	flag.StringVar(&featureGates, "feature-gates", "karpenter=false", "Enable Kaito feature gates. Default,	karpenter=false.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "ef60f9b0.io",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		klog.ErrorS(err, "unable to start manager")
		exitWithErrorFunc()
	}

	k8sclient.SetGlobalClient(mgr.GetClient())

	if err = (&controllers.WorkspaceReconciler{
		Client:   k8sclient.GetGlobalClient(),
		Log:      log.Log.WithName("controllers").WithName("Workspace"),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("KAITO-Workspace-controller"),
	}).SetupWithManager(mgr); err != nil {
		klog.ErrorS(err, "unable to create controller", "controller", "Workspace")
		exitWithErrorFunc()
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		klog.ErrorS(err, "unable to set up health check")
		exitWithErrorFunc()
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		klog.ErrorS(err, "unable to set up ready check")
		exitWithErrorFunc()
	}

	if enableWebhook {
		klog.InfoS("starting webhook reconcilers")
		p, err := strconv.Atoi(os.Getenv(WebhookServicePort))
		if err != nil {
			klog.ErrorS(err, "unable to parse the webhook port number")
			exitWithErrorFunc()
		}
		ctx := webhook.WithOptions(signals.NewContext(), webhook.Options{
			ServiceName: os.Getenv(WebhookServiceName),
			Port:        p,
			SecretName:  "workspace-webhook-cert",
		})
		ctx = sharedmain.WithHealthProbesDisabled(ctx)
		ctx = sharedmain.WithHADisabled(ctx)
		go sharedmain.MainWithConfig(ctx, "webhook", ctrl.GetConfigOrDie(), webhooks.NewWebhooks()...)

		// wait 2 seconds to allow reconciling webhookconfiguration and service endpoint.
		time.Sleep(2 * time.Second)
		if err = ParseFeatureGates(featureGates); err != nil {
			klog.ErrorS(err, "unable to parse `feature-gates` flag")
			exitWithErrorFunc()
		}
	}

	klog.InfoS("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		klog.ErrorS(err, "problem running manager")
		exitWithErrorFunc()
	}
}

// ParseFeatureGates parses the feature gates flag and sets the environment variables for each feature.
func ParseFeatureGates(featureGates string) error {
	gateMap := map[string]bool{}

	if err := cliflag.NewMapStringBool(&gateMap).Set(featureGates); err != nil {
		return err
	}
	if val, ok := gateMap["karpenter"]; ok {
		// set the environment variable to enable karpenter feature
		return os.Setenv(consts.FeatureFlagEnableKarpenter, strconv.FormatBool(val))
	}
	// add more feature gates here
	return nil
}
