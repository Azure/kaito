package v1alpha1

const (

	// Non-prefixed labels/annotations are reserved for end-use.

	// KDMPrefix Kubernetes Data Mining prefix.
	KDMPrefix = "kubernetes-kdm.io/"

	// AnnotationServiceType determines whether kdm creates ClusterIP or LoadBalancer type service.
	AnnotationServiceType = KDMPrefix + "service-type"

	// LabelWorkspaceName is the label for workspace name.
	LabelWorkspaceName = KDMPrefix + "workspace-name"

	ServiceTypeClusterIP    = "cluster-ip"
	ServiceTypeLoadBalancer = "load-balancer"
)
