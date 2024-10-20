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

package networkattachment

import (
	"encoding/json"
	"net"

	netdefv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func GetNetworksFromGateway(gateway *gatewayapiv1.Gateway) []*v1alpha1.Network {
	networkAttachmentAnnotation, exists := gateway.Spec.Infrastructure.Annotations[v1alpha1.PodSelectedNetworks]
	if !exists {
		return []*v1alpha1.Network{}
	}

	podSelectedNetworkSubnets := []*net.IPNet{}
	podSelectedNetworkSubnetsStr, exists := gateway.Spec.Infrastructure.Annotations[v1alpha1.PodSelectedNetworkSubnets]
	if exists {
		podSelectedNetworkSubnetsSlice := []string{}
		err := json.Unmarshal([]byte(podSelectedNetworkSubnetsStr), &podSelectedNetworkSubnetsSlice)
		if err != nil {
			return []*v1alpha1.Network{}
		}

		for _, podSelectedNetworkSubnet := range podSelectedNetworkSubnetsSlice {
			_, ipNet, err := net.ParseCIDR(podSelectedNetworkSubnet)
			if err != nil {
				continue
			}

			podSelectedNetworkSubnets = append(podSelectedNetworkSubnets, ipNet)
		}
	}

	return []*v1alpha1.Network{
		{
			Name: netdefv1.NetworkAttachmentAnnot,
			NetworkAttachementAnnotation: &v1alpha1.NetworkAttachementAnnotation{
				Key:       netdefv1.NetworkAttachmentAnnot,
				StatusKey: netdefv1.NetworkStatusAnnot,
				Value:     string(networkAttachmentAnnotation),
			},
			NetwokSubnets: podSelectedNetworkSubnets,
		},
	}
}
