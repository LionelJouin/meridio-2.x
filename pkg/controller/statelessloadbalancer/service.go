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

	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (c *Controller) reconcileServices(ctx context.Context, gateway *gatewayapiv1.Gateway) error {
	serviceList := &v1.ServiceList{}

	// Get pods for this service so the endpointslices can be reconciled.
	matchingLabels := client.MatchingLabels{
		apis.LabelServiceProxyName: gateway.Name,
	}

	err := c.List(ctx, serviceList, matchingLabels) // todo: filter namespace
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	services := []*v1.Service{}

	for _, service := range serviceList.Items {
		s := service
		services = append(services, &s)
	}

	err = c.ServiceManager.SetServices(ctx, services)
	if err != nil {
		return fmt.Errorf("failed to set services: %w", err)
	}

	err = c.reconcileEndpointSlices(ctx, services)
	if err != nil {
		return err
	}

	return nil
}
