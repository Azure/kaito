package node

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

func GetNode(cts context.Context, nodeName string) (*v1.Node, error) {
	return nil, nil
}
