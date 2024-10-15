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

package reconciler

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/networkattachment"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type Service struct {
	client.Client
	EndpointSliceReconciler *EndpointSlice
}

// Reconcile reconciles all services managed by the gateway.
func (s *Service) Reconcile(ctx context.Context, gateway *gatewayapiv1.Gateway) error {
	networks := networkattachment.GetNetworksFromGateway(gateway)

	services := &v1.ServiceList{}

	// Get pods for this service so the endpointslices can be reconciled.
	matchingLabels := client.MatchingLabels{
		apis.LabelServiceProxyName: gateway.Name,
	}

	err := s.List(ctx, services, matchingLabels) // todo: filter namespace
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	for _, service := range services.Items {
		svc := service

		err = s.reconcileService(ctx, &svc, networks)
		if err != nil {
			return err
		}
	}

	// todo: cleanup old endpointslices

	return nil
}

// reconcileService reconciles a specific service.
func (s *Service) reconcileService(ctx context.Context, service *v1.Service, networks []*v1alpha1.Network) error {
	// Get pods for this service so the endpointslices can be reconciled.
	var matchingLabels client.MatchingLabels = service.Spec.Selector

	delete(matchingLabels, v1alpha1.LabelDummmySericeSelector)

	pods := &v1.PodList{}

	err := s.List(ctx,
		pods,
		matchingLabels) // todo: filter namespace
	if err != nil {
		return fmt.Errorf("failed to list the pods: %w", err)
	}

	return s.EndpointSliceReconciler.Reconcile(ctx, service, pods, networks)
}
