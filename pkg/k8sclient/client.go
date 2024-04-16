// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package k8sclient

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Client client.Client

func SetGlobalClient(c client.Client) {
	Client = c
}

func GetGlobalClient() client.Client {
	return Client
}
