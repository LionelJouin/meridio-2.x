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

package endpointslice

import (
	"net"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetIPs represents the function to get the IPs assigned to a pod
// for a specific network.
type GetIPs func(pod v1.Pod, networks []*v1alpha1.Network) ([]string, error)

func GetEndpointSlice(
	service *v1.Service,
	pods *v1.PodList,
	addressType v1discovery.AddressType,
	networks []*v1alpha1.Network,
	getIPsFunc GetIPs,
) (*v1discovery.EndpointSlice, error) {
	endpoints := getEndpoints(pods, addressType, networks, getIPsFunc)

	endpointSlice := &v1discovery.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetEndpointSliceName(service, addressType),
			Namespace: service.GetNamespace(),
			Labels: map[string]string{
				v1discovery.LabelServiceName: service.GetName(),
			},
		},
		Endpoints:   endpoints,
		AddressType: addressType,
	}

	return endpointSlice, nil
}

func getEndpoints(
	pods *v1.PodList,
	addressType v1discovery.AddressType,
	networks []*v1alpha1.Network,
	getIPsFunc GetIPs,
) []v1discovery.Endpoint {
	endpoints := []v1discovery.Endpoint{}

	// get the IPs and readiness of the pods
	for _, pod := range pods.Items {
		ready := podReady(pod)

		ips, _ := getIPsFunc(pod, networks) // todo: error

		endpointIPs := []string{}

		// Get only IPs of correct IP family
		for _, ip := range ips {
			ipAddr := net.ParseIP(ip)

			if ipAddr == nil {
				continue
			}

			ipFamily := v1discovery.AddressTypeIPv6

			if ipAddr.To4() != nil {
				ipFamily = v1discovery.AddressTypeIPv4
			}

			if ipFamily == addressType {
				endpointIPs = append(endpointIPs, ip)
			}
		}

		// an endpoint does not access empty address list
		if len(endpointIPs) == 0 {
			continue
		}

		endpnt := v1discovery.Endpoint{
			TargetRef: &v1.ObjectReference{
				Kind:      pod.Kind,
				Name:      pod.GetName(),
				Namespace: pod.GetNamespace(),
				UID:       pod.GetUID(),
			},
			Conditions: v1discovery.EndpointConditions{
				Ready: &ready,
			},
			Addresses: endpointIPs,
		}
		endpoints = append(endpoints, endpnt)
	}

	return endpoints
}

// podReady checks if a pod is ready: All containers in ready state and pod status running.
func podReady(pod v1.Pod) bool {
	for _, container := range pod.Status.ContainerStatuses {
		if !container.Ready {
			return false
		}
	}

	return pod.Status.Phase == v1.PodRunning
}
