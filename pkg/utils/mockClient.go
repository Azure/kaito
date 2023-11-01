// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license

package utils

import (
	"context"
	"reflect"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a mock for the controller-runtime dynamic client interface.
type Client struct {
	mock.Mock

	ObjectMap  map[reflect.Type]map[k8sClient.ObjectKey]k8sClient.Object
	StatusMock *StatusClient
}

var _ k8sClient.Client = &Client{}

func NewClient() *Client {
	return &Client{
		StatusMock: &StatusClient{},
		ObjectMap:  map[reflect.Type]map[k8sClient.ObjectKey]k8sClient.Object{},
	}
}

// Retrieves or creates a map associated with the type of obj
func (c *Client) ensureMapFor(obj k8sClient.Object) map[k8sClient.ObjectKey]k8sClient.Object {
	t := reflect.TypeOf(obj)
	if _, ok := c.ObjectMap[t]; !ok {
		//create a new map with the object key if it doesn't exist
		c.ObjectMap[t] = map[k8sClient.ObjectKey]k8sClient.Object{}
	}
	return c.ObjectMap[t]
}

func (c *Client) InsertObjectInMap(obj k8sClient.Object) {
	relevantMap := c.ensureMapFor(obj)
	objKey := k8sClient.ObjectKeyFromObject(obj)

	relevantMap[objKey] = obj
}

// StatusClient interface

func (c *Client) Status() k8sClient.StatusWriter {
	return c.StatusMock
}

// Reader interface

func (c *Client) Get(ctx context.Context, key types.NamespacedName, obj k8sClient.Object, opts ...k8sClient.GetOption) error {
	relevantMap := c.ensureMapFor(obj)

	for _, val := range relevantMap {
		v := reflect.ValueOf(obj).Elem()
		v.Set(reflect.ValueOf(val).Elem())
		break
	}

	args := c.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func (c *Client) List(ctx context.Context, list k8sClient.ObjectList, opts ...k8sClient.ListOption) error {
	args := c.Called(ctx, list, opts)
	return args.Error(0)
}

// Writer interface

func (c *Client) Create(ctx context.Context, obj k8sClient.Object, opts ...k8sClient.CreateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Delete(ctx context.Context, obj k8sClient.Object, opts ...k8sClient.DeleteOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Update(ctx context.Context, obj k8sClient.Object, opts ...k8sClient.UpdateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *Client) Patch(ctx context.Context, obj k8sClient.Object, patch k8sClient.Patch, opts ...k8sClient.PatchOption) error {
	args := c.Called(ctx, obj, patch, opts)
	return args.Error(0)
}

func (c *Client) DeleteAllOf(ctx context.Context, obj k8sClient.Object, opts ...k8sClient.DeleteAllOfOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

// SubResource implements client.Client
func (*Client) SubResource(subResource string) k8sClient.SubResourceClient {
	panic("unimplemented")
}

// GroupVersionKindFor implements client.Client
func (*Client) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	panic("unimplemented")
}

// IsObjectNamespaced implements client.Client
func (*Client) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	panic("unimplemented")
}

func (c *Client) Scheme() *runtime.Scheme {
	args := c.Called()
	return args.Get(0).(*runtime.Scheme)
}

func (c *Client) RESTMapper() meta.RESTMapper {
	args := c.Called()
	return args.Get(0).(meta.RESTMapper)
}

type StatusClient struct {
	mock.Mock
}

func (c *StatusClient) Create(ctx context.Context, obj k8sClient.Object, subResource k8sClient.Object, opts ...k8sClient.SubResourceCreateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *StatusClient) Patch(ctx context.Context, obj k8sClient.Object, patch k8sClient.Patch, opts ...k8sClient.SubResourcePatchOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

func (c *StatusClient) Update(ctx context.Context, obj k8sClient.Object, opts ...k8sClient.SubResourceUpdateOption) error {
	args := c.Called(ctx, obj, opts)
	return args.Error(0)
}

var _ k8sClient.StatusWriter = &StatusClient{}
