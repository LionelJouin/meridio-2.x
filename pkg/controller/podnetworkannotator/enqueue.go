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

package podnetworkannotator

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (c *Controller) serviceEnqueue(
	ctx context.Context,
	object client.Object,
) []reconcile.Request {
	service, ok := object.(*v1.Service)
	if !ok {
		return []reconcile.Request{}
	}

	// todo: check if parent is the right class

	pods, err := c.getPodsForService(ctx, service)
	if err != nil {
		return []reconcile.Request{}
	}

	reconcileRequests := []reconcile.Request{}

	for _, pod := range pods.Items {
		reconcileRequests = append(reconcileRequests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
			},
		})
	}

	return reconcileRequests
}

func (c *Controller) gatewayEnqueue(
	ctx context.Context,
	object client.Object,
) []reconcile.Request {
	serviceList := &v1.ServiceList{}

	err := c.List(ctx,
		serviceList,
		client.MatchingLabels{
			apis.LabelServiceProxyName: object.GetName(),
		},
		client.InNamespace(object.GetNamespace()),
	)
	if err != nil {
		log.FromContextOrGlobal(ctx).Error(err, "failed listing the services during the pod enqueue")

		return []reconcile.Request{}
	}

	pods, err := c.getPodsForServices(ctx, serviceList)
	if err != nil {
		return []reconcile.Request{}
	}

	reconcileRequests := []reconcile.Request{}

	for _, pod := range pods {
		reconcileRequests = append(reconcileRequests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
			},
		})
	}

	return reconcileRequests
}

func (c *Controller) podEnqueue(
	ctx context.Context,
	object client.Object,
) []reconcile.Request {
	gatewayName, exists := object.GetLabels()[apis.LabelServiceProxyName]
	if !exists {
		return []reconcile.Request{}
	}

	serviceList := &v1.ServiceList{}

	err := c.List(ctx,
		serviceList,
		client.MatchingLabels{
			apis.LabelServiceProxyName: gatewayName,
		},
		client.InNamespace(object.GetNamespace()),
	)
	if err != nil {
		log.FromContextOrGlobal(ctx).Error(err, "failed listing the services during the pod enqueue")

		return []reconcile.Request{}
	}

	pods, err := c.getPodsForServices(ctx, serviceList)
	if err != nil {
		return []reconcile.Request{}
	}

	reconcileRequests := []reconcile.Request{}

	for _, pod := range pods {
		reconcileRequests = append(reconcileRequests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
			},
		})
	}

	return reconcileRequests
}

func (c *Controller) getPodsForServices(ctx context.Context, services *v1.ServiceList) ([]*v1.Pod, error) {
	pods := []*v1.Pod{}
	podsMap := map[types.NamespacedName]struct{}{}

	for _, service := range services.Items {
		podList, err := c.getPodsForService(ctx, &service)
		if err != nil {
			return nil, err
		}

		for _, pod := range podList.Items {
			namespacedName := types.NamespacedName{
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
			}

			_, exists := podsMap[namespacedName]
			if exists {
				continue
			}

			podsMap[namespacedName] = struct{}{}
			pods = append(pods, &pod)
		}
	}

	return pods, nil
}

func (c *Controller) getPodsForService(ctx context.Context, service *v1.Service) (*v1.PodList, error) {
	var matchingLabels client.MatchingLabels = service.Spec.Selector

	delete(matchingLabels, v1alpha1.LabelDummmySericeSelector)

	pods := &v1.PodList{}

	err := c.List(ctx,
		pods,
		matchingLabels,
		client.InNamespace(service.GetNamespace()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed listing the pods during the pod enqueue")
	}

	return pods, nil
}
