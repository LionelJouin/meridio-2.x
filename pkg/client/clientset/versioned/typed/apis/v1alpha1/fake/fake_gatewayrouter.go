/*
Copyright (c) 2024 OpenInfra Foundation Europe

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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeGatewayRouters implements GatewayRouterInterface
type FakeGatewayRouters struct {
	Fake *FakeMeridioV1alpha1
	ns   string
}

var gatewayroutersResource = v1alpha1.SchemeGroupVersion.WithResource("gatewayrouters")

var gatewayroutersKind = v1alpha1.SchemeGroupVersion.WithKind("GatewayRouter")

// Get takes name of the gatewayRouter, and returns the corresponding gatewayRouter object, and an error if there is any.
func (c *FakeGatewayRouters) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.GatewayRouter, err error) {
	emptyResult := &v1alpha1.GatewayRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(gatewayroutersResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.GatewayRouter), err
}

// List takes label and field selectors, and returns the list of GatewayRouters that match those selectors.
func (c *FakeGatewayRouters) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.GatewayRouterList, err error) {
	emptyResult := &v1alpha1.GatewayRouterList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(gatewayroutersResource, gatewayroutersKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.GatewayRouterList{ListMeta: obj.(*v1alpha1.GatewayRouterList).ListMeta}
	for _, item := range obj.(*v1alpha1.GatewayRouterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested gatewayRouters.
func (c *FakeGatewayRouters) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(gatewayroutersResource, c.ns, opts))

}

// Create takes the representation of a gatewayRouter and creates it.  Returns the server's representation of the gatewayRouter, and an error, if there is any.
func (c *FakeGatewayRouters) Create(ctx context.Context, gatewayRouter *v1alpha1.GatewayRouter, opts v1.CreateOptions) (result *v1alpha1.GatewayRouter, err error) {
	emptyResult := &v1alpha1.GatewayRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(gatewayroutersResource, c.ns, gatewayRouter, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.GatewayRouter), err
}

// Update takes the representation of a gatewayRouter and updates it. Returns the server's representation of the gatewayRouter, and an error, if there is any.
func (c *FakeGatewayRouters) Update(ctx context.Context, gatewayRouter *v1alpha1.GatewayRouter, opts v1.UpdateOptions) (result *v1alpha1.GatewayRouter, err error) {
	emptyResult := &v1alpha1.GatewayRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(gatewayroutersResource, c.ns, gatewayRouter, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.GatewayRouter), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeGatewayRouters) UpdateStatus(ctx context.Context, gatewayRouter *v1alpha1.GatewayRouter, opts v1.UpdateOptions) (result *v1alpha1.GatewayRouter, err error) {
	emptyResult := &v1alpha1.GatewayRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(gatewayroutersResource, "status", c.ns, gatewayRouter, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.GatewayRouter), err
}

// Delete takes name of the gatewayRouter and deletes it. Returns an error if one occurs.
func (c *FakeGatewayRouters) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(gatewayroutersResource, c.ns, name, opts), &v1alpha1.GatewayRouter{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeGatewayRouters) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(gatewayroutersResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.GatewayRouterList{})
	return err
}

// Patch applies the patch and returns the patched gatewayRouter.
func (c *FakeGatewayRouters) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.GatewayRouter, err error) {
	emptyResult := &v1alpha1.GatewayRouter{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(gatewayroutersResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.GatewayRouter), err
}
