/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package common

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/azure/kaito/api/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	loggingtesting "knative.dev/pkg/logging/testing"
	"knative.dev/pkg/system"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Environment struct {
	context.Context

	Client     client.Client
	Config     *rest.Config
	KubeClient kubernetes.Interface
}

func NewEnvironment(t *testing.T) *Environment {
	ctx := loggingtesting.TestContextWithLogger(t)
	config := NewConfig()
	client := lo.Must(NewClient(config))

	lo.Must0(os.Setenv(system.NamespaceEnvKey, "kaito"))
	kubernetesInterface := kubernetes.NewForConfigOrDie(config)

	gomega.SetDefaultEventuallyTimeout(5 * time.Minute)
	gomega.SetDefaultEventuallyPollingInterval(1 * time.Second)
	return &Environment{
		Context:    ctx,
		Config:     config,
		Client:     client,
		KubeClient: kubernetesInterface,
	}
}

func NewConfig() *rest.Config {
	config := controllerruntime.GetConfigOrDie()
	config.UserAgent = fmt.Sprintf("%s-%s", v1alpha1.GroupVersion.Group, v1alpha1.GroupVersion.Version)
	config.QPS = 1e6
	config.Burst = 1e6
	return config
}

func NewClient(config *rest.Config) (client.Client, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return client.New(config, client.Options{Scheme: scheme})
}
