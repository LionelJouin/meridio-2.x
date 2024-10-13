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

package statelessloadbalancer

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/pkg/endpointslice"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Controller) reconcileEndpointSlices(ctx context.Context, services []*v1.Service) error {
	for _, service := range services {
		err := c.reconcileEndpointSlice(ctx, service)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) reconcileEndpointSlice(ctx context.Context, service *v1.Service) error {
	// Get and set endpoints corresponding to this service. The endpoints are splitted
	// into 2 endpointslices, 1 for ipv4 and the other for ipv6. They are then merged together.
	ipv4EndpointSlice := &v1discovery.EndpointSlice{}
	ipv6EndpointSlice := &v1discovery.EndpointSlice{}

	err := c.Get(ctx, types.NamespacedName{
		Name:      endpointslice.GetEndpointSliceName(service, v1discovery.AddressTypeIPv4),
		Namespace: service.GetNamespace(),
	}, ipv4EndpointSlice)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get IPv4 EndpointSlice: %w", err)
		}
	}

	err = c.Get(ctx, types.NamespacedName{
		Name:      endpointslice.GetEndpointSliceName(service, v1discovery.AddressTypeIPv6),
		Namespace: service.GetNamespace(),
	}, ipv6EndpointSlice)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get IPv6 EndpointSlice: %w", err)
		}
	}

	mergedEndpointSlices := endpointslice.MergeEndpointSlices(ipv4EndpointSlice, ipv6EndpointSlice)
	mergedEndpoints := []v1discovery.Endpoint{}

	if mergedEndpointSlices != nil {
		mergedEndpoints = mergedEndpointSlices.Endpoints
	}

	err = c.ServiceManager.SetEndpoints(ctx, service, mergedEndpoints)
	if err != nil {
		return fmt.Errorf("failed to set endpoints with service manager: %w", err)
	}

	return nil
}
