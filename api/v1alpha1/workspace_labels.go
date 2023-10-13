package v1alpha1

const (

	// Non-prefixed labels/annotations are reserved for end-use.

	// KAITOPrefix Kubernetes Data Mining prefix.
	KAITOPrefix = "kubernetes-kaito.sh/"

	// AnnotationServiceType determines whether kaito creates ClusterIP or LoadBalancer type service.
	AnnotationServiceType = KAITOPrefix + "service-type"

	// LabelWorkspaceName is the label for workspace name.
	LabelWorkspaceName = KAITOPrefix + "workspace-name"

	ServiceTypeClusterIP    = "cluster-ip"
	ServiceTypeLoadBalancer = "load-balancer"
)
