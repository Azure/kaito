package controllers

import (
	"reflect"
	"sort"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSelectWorkspaceNodes(t *testing.T) {

	testcases := map[string]struct {
		qualified []*corev1.Node
		preferred []string
		previous  []string
		count     int
		expected  []string
	}{
		"two qualified nodes, need one": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
			},
			preferred: []string{},
			previous:  []string{},
			count:     1,
			expected:  []string{"node1"},
		},

		"three qualified nodes, prefer two of them": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{"node3", "node2"},
			previous:  []string{},
			count:     2,
			expected:  []string{"node2", "node3"},
		},

		"three qualified nodes, two of them are selected previously, need two": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{},
			previous:  []string{"node3", "node2"},
			count:     2,
			expected:  []string{"node2", "node3"},
		},

		"three qualified nodes, one preferred, one previous, need two": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{"node3"},
			previous:  []string{"node2"},
			count:     2,
			expected:  []string{"node2", "node3"},
		},

		"three qualified nodes, one preferred, one previous, need one": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
					},
				},
			},
			preferred: []string{"node3"},
			previous:  []string{"node2"},
			count:     1,
			expected:  []string{"node3"},
		},

		"three qualified nodes, one is created by kaito, need one": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							"kaito.sh/machine-type": "gpu",
						},
					},
				},
			},
			preferred: []string{},
			previous:  []string{},
			count:     1,
			expected:  []string{"node3"},
		},
		"three qualified nodes, one is created by kaito, one is preferred, one is previous, need two": {
			qualified: []*corev1.Node{
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node1",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node2",
					},
				},
				&corev1.Node{
					ObjectMeta: v1.ObjectMeta{
						Name: "node3",
						Labels: map[string]string{
							"kaito.sh/machine-type": "gpu",
						},
					},
				},
			},
			preferred: []string{"node2"},
			previous:  []string{"node1"},
			count:     2,
			expected:  []string{"node1", "node2"},
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {

			selectedNodes := selectWorkspaceNodes(tc.qualified, tc.preferred, tc.previous, tc.count)

			selectedNodesArray := []string{}

			for _, each := range selectedNodes {
				selectedNodesArray = append(selectedNodesArray, each.Name)
			}

			sort.Strings(selectedNodesArray)
			sort.Strings(tc.expected)

			if !reflect.DeepEqual(selectedNodesArray, tc.expected) {
				t.Errorf("%s: selected Nodes %+v are different from the expected %+v", k, selectedNodesArray, tc.expected)
			}
		})
	}
}
