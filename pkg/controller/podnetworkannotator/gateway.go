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
	"net"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	"github.com/lioneljouin/meridio-experiment/pkg/networkannotation"
	"github.com/lioneljouin/meridio-experiment/pkg/networkattachment"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/proxy/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (c *Controller) reconcileGateways(ctx context.Context, pod *v1.Pod) error {
	gateways, err := c.getGatewaysForGatewayClass(ctx)
	if err != nil {
		return err
	}

	newNetworkConfiguration := networkannotation.NetworkConfiguration{}

	for _, gateway := range gateways {
		networkRoute, err := c.reconcileGateway(ctx, &gateway, pod) // check if gateway selects the pod
		if err != nil {
			return err
		}

		if networkRoute == nil {
			continue
		}

		newNetworkConfiguration[types.NamespacedName{
			Name:      gateway.GetName(),
			Namespace: gateway.GetNamespace(),
		}.String()] = networkRoute
	}

	pod, err = c.NetworkAnnotator.SetNetworkAnnotation(pod, newNetworkConfiguration)
	if err != nil {
		return fmt.Errorf("failed to set network to the pod: %w", err)
	}

	// cleanup the pod

	err = c.Update(ctx, pod)
	if err != nil {
		return fmt.Errorf("failed to udpdate the pod: %w", err)
	}

	return nil
}

func (c *Controller) reconcileGateway(ctx context.Context, gateway *gatewayapiv1.Gateway, pod *v1.Pod) (*networkannotation.NetworkRoute, error) {
	serviceList := &v1.ServiceList{}

	err := c.List(ctx,
		serviceList,
		client.MatchingLabels{
			apis.LabelServiceProxyName: gateway.Name,
		},
		client.InNamespace(pod.GetNamespace()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list the services: %w", err)
	}

	if len(filterServices(pod, serviceList.Items)) == 0 {
		return nil, nil
	}

	vipV4, vipV6 := getVIPs(gateway)

	networks := networkattachment.GetNetworksFromGateway(gateway)
	gatewaysV4, gatewaysV6, err := c.getGatewayIPs(ctx, gateway, networks)
	if err != nil {
		return nil, err
	}

	if (len(vipV4) == 0 || len(gatewaysV4) == 0) && (len(vipV6) == 0 || len(gatewaysV6) == 0) {
		return nil, nil
	}

	return &networkannotation.NetworkRoute{
		VIPsV4:     vipV4,
		VIPsV6:     vipV6,
		GatewaysV4: gatewaysV4,
		GatewaysV6: gatewaysV6,
	}, nil
}

func getVIPs(
	gateway *gatewayapiv1.Gateway,
) ([]string, []string) {
	vipsV4 := []string{}
	vipsV6 := []string{}

	// remove duplicates and invalid VIPs
	vipMap := map[string]struct{}{}

	for _, address := range gateway.Status.Addresses {
		ipAddr := net.ParseIP(address.Value)

		if ipAddr == nil {
			continue
		}

		v4 := true
		vip := fmt.Sprintf("%s/32", ipAddr)

		if ipAddr.To4() == nil {
			vip = fmt.Sprintf("%s/128", ipAddr)
			v4 = false
		}

		_, exists := vipMap[vip]
		if exists {
			continue
		}

		if v4 {
			vipsV4 = append(vipsV4, vip)
		} else {
			vipsV6 = append(vipsV6, vip)
		}

		vipMap[vip] = struct{}{}
	}

	return vipsV4, vipsV6
}

func (c *Controller) getGatewayIPs(
	ctx context.Context,
	gateway *gatewayapiv1.Gateway,
	networks []*v1alpha1.Network,
) ([]string, []string, error) {
	pods := &v1.PodList{}

	err := c.List(ctx,
		pods,
		client.MatchingLabels{
			apis.LabelServiceProxyName: gateway.GetName(),
		}) // todo: filter namespace
	if err != nil {
		return nil, nil, fmt.Errorf("failed listing the service proxy data-plane pods: %w", err)
	}

	ipv4 := []string{}
	ipv6 := []string{}

	for _, pod := range pods.Items {
		ips, _ := c.GetIPsFunc(pod, networks) // todo: error

		// Get only IPs of correct IP family
		for _, ip := range ips {
			ipAddr := net.ParseIP(ip)

			if ipAddr == nil {
				continue
			}

			if ipAddr.To4() != nil {
				ipv4 = append(ipv4, ip)

				continue
			}

			ipv6 = append(ipv6, ip)
		}
	}

	return ipv4, ipv6, nil
}

func filterServices(
	pod *v1.Pod,
	services []v1.Service,
) []*v1.Service {
	res := []*v1.Service{}

items:
	for _, service := range services {
		for labelSelectorKey, labelSelectorValue := range service.Spec.Selector {
			if labelSelectorKey == v1alpha1.LabelDummmySericeSelector {
				continue
			}

			value, exists := pod.GetLabels()[labelSelectorKey]
			if !exists || value != labelSelectorValue {
				continue items
			}
		}
		ns := service
		res = append(res, &ns)
	}

	return res
}

func (c *Controller) getGatewaysForGatewayClass(ctx context.Context) ([]gatewayapiv1.Gateway, error) {
	gatewayList := &gatewayapiv1.GatewayList{}

	// err := c.List(ctx,
	// 	gatewayList,
	// 	client.MatchingFields{
	// 		"spec.gatewayClassName": c.GatewayClassName,
	// 	})
	err := c.List(ctx,
		gatewayList) // todo: filter namespace
	if err != nil {
		log.FromContextOrGlobal(ctx).Error(err, "failed listing the gateways during the pod enqueue")

		return nil, err
	}

	gateways := []gatewayapiv1.Gateway{}

	for _, gateway := range gatewayList.Items {
		if string(gateway.Spec.GatewayClassName) != c.GatewayClassName {
			continue
		}

		gateways = append(gateways, gateway)
	}

	return gateways, nil
}
