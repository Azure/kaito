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
	"sync"

	"github.com/azure/kaito/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2" //nolint:revive,stylecheck
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/karpenter-core/pkg/apis"
	"github.com/aws/karpenter-core/pkg/apis/v1alpha5"
	"github.com/aws/karpenter-core/pkg/operator/injection"
)

var (
	CleanableObjects = []client.Object{
		&v1.Pod{},
		&appsv1.Deployment{},
		&v1.Service{},
		&v1alpha1.Workspace{},
		&v1alpha5.Machine{},
		&v1.Node{},
		&appsv1.DaemonSet{},
		&appsv1.ReplicaSet{},
	}
)

// nolint:gocyclo
func (env *Environment) BeforeEach() {
	env.Context = injection.WithSettingsOrDie(env.Context, env.KubeClient, apis.Settings...)
}

func (env *Environment) Cleanup() {
	env.CleanupObjects(CleanableObjects...)
}

func (env *Environment) AfterEach() {
	env.printControllerLogs(&v1.PodLogOptions{Container: "kaito"})
}

func (env *Environment) CleanupObjects(cleanableObjects ...client.Object) {
	wg := sync.WaitGroup{}
	for _, obj := range cleanableObjects {
		wg.Add(1)
		go func(obj client.Object) {
			defer wg.Done()
			defer GinkgoRecover()

		}(obj)
	}
	wg.Wait()
}
