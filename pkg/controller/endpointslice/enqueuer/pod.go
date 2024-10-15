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

package enqueuer

import (
	"context"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (e *enqueuer) PodEnqueue(
	ctx context.Context,
	object client.Object,
) []reconcile.Request {
	reconcileRequests := []reconcile.Request{}

	services := e.getServicesForGateways(ctx, e.getGatewaysForGatewayClass(ctx))

items:
	for _, service := range services {
		for labelSelectorKey, labelSelectorValue := range service.Spec.Selector {
			if labelSelectorKey == v1alpha1.LabelDummmySericeSelector {
				continue
			}

			value, exists := object.GetLabels()[labelSelectorKey]
			if !exists || value != labelSelectorValue {
				continue items
			}
		}

		gatewayName, exists := service.Labels[apis.LabelServiceProxyName]
		if !exists {
			continue
		}

		reconcileRequests = append(reconcileRequests,
			reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      gatewayName,
					Namespace: service.GetNamespace(),
				},
			})
	}

	return reconcileRequests
}

func (e *enqueuer) getServicesForGateways(ctx context.Context, gateways []gatewayapiv1.Gateway) []v1.Service {
	services := []v1.Service{}

	for _, gateway := range gateways {
		serviceList := &v1.ServiceList{}

		err := e.List(ctx,
			serviceList,
			client.MatchingLabels{
				apis.LabelServiceProxyName: gateway.Name,
			}) // todo: filter namespace
		if err != nil {
			log.FromContextOrGlobal(ctx).Error(err, "failed listing the services during the pod enqueue")
		}

		services = append(services, serviceList.Items...)
	}

	return services
}

func (e *enqueuer) getGatewaysForGatewayClass(ctx context.Context) []gatewayapiv1.Gateway {
	gatewayList := &gatewayapiv1.GatewayList{}

	// err := c.List(ctx,
	// 	gatewayList,
	// 	client.MatchingFields{
	// 		"spec.gatewayClassName": c.GatewayClassName,
	// 	})
	err := e.List(ctx,
		gatewayList) // todo: filter namespace
	if err != nil {
		log.FromContextOrGlobal(ctx).Error(err, "failed listing the gateways during the pod enqueue")

		return nil
	}

	gateways := []gatewayapiv1.Gateway{}

	for _, gateway := range gatewayList.Items {
		if string(gateway.Spec.GatewayClassName) != e.GatewayClassName {
			continue
		}

		gateways = append(gateways, gateway)
	}

	return gateways
}
