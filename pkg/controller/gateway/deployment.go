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

package gateway

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/pkg/template"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/proxy/apis"
	ctrl "sigs.k8s.io/controller-runtime"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	label                              = "app"
	routerContainerName                = "router"
	statelessLoadBalancerContainerName = "stateless-load-balancer"
)

func (c *Controller) reconcileStatelessLoadBalancerDeployment(
	ctx context.Context,
	gateway *gatewayapiv1.Gateway,
) error {
	statelessLoadBalancerDeployment := &appsv1.Deployment{}

	statelessLoadBalancerDeploymentLatestState, err := c.getStatelessLoadBalancerDeployment(gateway)
	if err != nil {
		return err
	}

	err = c.Get(ctx, types.NamespacedName{
		Name:      getStatelessLoadBalancerDeploymentName(gateway),
		Namespace: gateway.Namespace,
	}, statelessLoadBalancerDeployment)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Create
			err = c.Create(ctx, statelessLoadBalancerDeploymentLatestState)
			if err != nil {
				return fmt.Errorf("failed to create the stateless-load-balancer deployment: %w", err)
			}

			return nil
		}

		return fmt.Errorf("failed to get the stateless-load-balancer deployment: %w", err)
	}

	// TODO: Update
	// // Update
	// err = c.Patch(ctx, statelessLoadBalancerDeployment, client.StrategicMergeFrom(statelessLoadBalancerDeploymentLatestState))
	// if err != nil {
	// 	return fmt.Errorf("failed to update the stateless-load-balancer deployment: %w", err)
	// }

	return nil
}

func (c *Controller) getStatelessLoadBalancerDeployment(
	gateway *gatewayapiv1.Gateway,
) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}

	err := template.Read("/templates/stateless-load-balancer.yaml", deployment)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment stateless-load-balancer template: %w", err)
	}

	name := getStatelessLoadBalancerDeploymentName(gateway)

	deployment.ObjectMeta.Name = name
	deployment.ObjectMeta.Namespace = gateway.Namespace

	if deployment.Labels == nil {
		deployment.Labels = map[string]string{}
	}

	deployment.Labels[label] = name

	deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			label: name,
		},
	}

	if deployment.Spec.Template.Labels == nil {
		deployment.Spec.Template.Labels = map[string]string{}
	}

	deployment.Spec.Template.Labels[label] = name

	for key, value := range gateway.Spec.Infrastructure.Labels {
		deployment.Labels[string(key)] = string(value)
		deployment.Spec.Template.Labels[string(key)] = string(value)
	}

	if deployment.Annotations == nil {
		deployment.Annotations = map[string]string{}
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{}
	}

	for key, value := range gateway.Spec.Infrastructure.Annotations {
		deployment.Annotations[string(key)] = string(value)
		deployment.Spec.Template.Annotations[string(key)] = string(value)
	}

	deployment.Labels[apis.LabelServiceProxyName] = gateway.GetName()
	deployment.Spec.Template.Labels[apis.LabelServiceProxyName] = gateway.GetName()

	for index, container := range deployment.Spec.Template.Spec.Containers {
		switch container.Name {
		case statelessLoadBalancerContainerName:
			deployment.Spec.Template.Spec.Containers[index].Args = append(deployment.Spec.Template.Spec.Containers[index].Args,
				fmt.Sprintf("--gateway-class-name=%s", gateway.Spec.GatewayClassName),
				fmt.Sprintf("--name=%s", gateway.Name),
				fmt.Sprintf("--namespace=%s", gateway.Namespace),
			)
		case routerContainerName:
			deployment.Spec.Template.Spec.Containers[index].Args = append(deployment.Spec.Template.Spec.Containers[index].Args,
				fmt.Sprintf("--name=%s", gateway.Name),
				fmt.Sprintf("--namespace=%s", gateway.Namespace),
			)
		}
	}

	err = ctrl.SetControllerReference(gateway, deployment, c.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to SetControllerReference on stateless-load-balancer deployment: %w", err)
	}

	return deployment, nil
}

func getStatelessLoadBalancerDeploymentName(gateway *gatewayapiv1.Gateway) string {
	return fmt.Sprintf("stateless-load-balancer-%s", gateway.Name)
}
