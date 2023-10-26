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
	"bytes"
	"fmt"
	"io"

	//"github.com/aws/karpenter-core/pkg/test"
	"github.com/azure/kaito/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2" //nolint:revive,stylecheck
	. "github.com/onsi/gomega"    //nolint:revive,stylecheck
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (env *Environment) ExpectCreatedWithOffset(offset int, objects ...client.Object) {
	for _, object := range objects {
		object.SetLabels(lo.Assign(object.GetLabels(), map[string]string{
			"test": "unspecified",
		}))
		ExpectWithOffset(offset+1, env.Client.Create(env, object)).To(Succeed())
	}
}

func (env *Environment) ExpectCreated(objects ...client.Object) {
	env.ExpectCreatedWithOffset(1, objects...)
}

func (env *Environment) EventuallyExpectCreatedWorkspaceCount(comparator string, count int) []*v1alpha1.Workspace {
	By(fmt.Sprintf("waiting for created workspace to be %s to %d", comparator, count))
	workspaceList := &v1alpha1.WorkspaceList{}
	EventuallyWithOffset(1, func(g Gomega) {
		g.Expect(env.Client.List(env.Context, workspaceList)).To(Succeed())
		g.Expect(len(workspaceList.Items)).To(BeNumerically(comparator, count))
	}).Should(Succeed())
	return lo.Map(workspaceList.Items, func(w v1alpha1.Workspace, _ int) *v1alpha1.Workspace {
		return &w
	})
}

func (env *Environment) EventuallyExpectWorkspaceReady(workspaces ...*v1alpha1.Workspace) {
	Eventually(func(g Gomega) {
		for _, workspace := range workspaces {
			temp := &v1alpha1.Workspace{}
			g.Expect(env.Client.Get(env.Context, client.ObjectKeyFromObject(workspace), temp)).Should(Succeed())
		}
	}).Should(Succeed())
}

func (env *Environment) ExpectKaitoDeployment() {
	GinkgoHelper()
	deploymentList := &appsv1.DeploymentList{}
	Expect(env.Client.List(env.Context, deploymentList, client.MatchingLabels{
		"app.kubernetes.io/instance": "kaito",
	})).To(Succeed())
}

func (env *Environment) ExpectKaitoReplicaSet() {
	GinkgoHelper()
	replicaSetList := &appsv1.ReplicaSetList{}
	Expect(env.Client.List(env.Context, replicaSetList, client.MatchingLabels{
		"app.kubernetes.io/instance": "kaito",
	})).To(Succeed())
}

func (env *Environment) ExpectKaitoService() {
	GinkgoHelper()
	serviceList := &v1.ServiceList{}
	Expect(env.Client.List(env.Context, serviceList, client.MatchingLabels{
		"app.kubernetes.io/instance": "kaito",
	})).To(Succeed())
}

func (env *Environment) ExpectKaitoPods() []*v1.Pod {
	GinkgoHelper()
	podList := &v1.PodList{}
	Expect(env.Client.List(env.Context, podList, client.MatchingLabels{
		"app.kubernetes.io/instance": "kaito",
	})).To(Succeed())
	return lo.Map(podList.Items, func(p v1.Pod, _ int) *v1.Pod { return &p })
}

var (
	lastLogged = metav1.Now()
)

func (env *Environment) printControllerLogs(options *v1.PodLogOptions) {
	fmt.Println("------- START CONTROLLER LOGS -------")
	defer fmt.Println("------- END CONTROLLER LOGS -------")

	if options.SinceTime == nil {
		options.SinceTime = lastLogged.DeepCopy()
		lastLogged = metav1.Now()
	}
	pods := env.ExpectKaitoPods()
	for _, pod := range pods {
		temp := options.DeepCopy() // local version of the log options

		fmt.Printf("------- pod/%s -------\n", pod.Name)
		if pod.Status.ContainerStatuses[0].RestartCount > 0 {
			fmt.Printf("[PREVIOUS CONTAINER LOGS]\n")
			temp.Previous = true
		}
		stream, err := env.KubeClient.CoreV1().Pods("kaito").GetLogs(pod.Name, temp).Stream(env.Context)
		if err != nil {
			logging.FromContext(env.Context).Errorf("fetching controller logs: %s", err)
			return
		}
		log := &bytes.Buffer{}
		_, err = io.Copy(log, stream)
		Expect(err).ToNot(HaveOccurred())
		logging.FromContext(env.Context).Info(log)
	}
}
