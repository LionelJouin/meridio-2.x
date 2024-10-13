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

	v1discovery "k8s.io/api/discovery/v1"
)

// MergeEndpointSlices merges two endpoint slices address and keep the data from endpointSliceA
// if endpointSliceA and endpointSliceB data are not corresponding.
func MergeEndpointSlices(
	endpointSliceA *v1discovery.EndpointSlice,
	endpointSliceB *v1discovery.EndpointSlice,
) *v1discovery.EndpointSlice {
	if endpointSliceB == nil {
		return endpointSliceA
	}

	if endpointSliceA == nil {
		return endpointSliceB
	}

	endpointSliceRes := &v1discovery.EndpointSlice{
		ObjectMeta: endpointSliceA.ObjectMeta,
		Endpoints:  []v1discovery.Endpoint{},
	}

	endpointsMap := map[string]*v1discovery.Endpoint{}

	for _, endpnt := range endpointSliceA.Endpoints {
		currentEndpoint := endpnt

		endpointsMap[string(currentEndpoint.TargetRef.UID)] = &currentEndpoint
	}

	for _, endpnt := range endpointSliceB.Endpoints {
		currentEndpoint := endpnt

		existingEndpoint, exists := endpointsMap[string(currentEndpoint.TargetRef.UID)]
		if !exists {
			endpointsMap[string(currentEndpoint.TargetRef.UID)] = &currentEndpoint

			continue
		}

		existingEndpoint.Addresses = append(existingEndpoint.Addresses, currentEndpoint.Addresses...)

		if existingEndpoint.Zone == nil {
			existingEndpoint.Zone = currentEndpoint.Zone
		}
	}

	for _, endpnt := range endpointsMap {
		endpointSliceRes.Endpoints = append(endpointSliceRes.Endpoints, *endpnt)
	}

	return endpointSliceRes
}

// SplitEndpointSlices splits an endpoint slice into 2 (ipv4 and ipv6)
// Only endpoints are kept.
func SplitEndpointSlices(
	endpointSlice *v1discovery.EndpointSlice,
) (*v1discovery.EndpointSlice, *v1discovery.EndpointSlice) {
	ipv4EndpointSlice := &v1discovery.EndpointSlice{
		AddressType: v1discovery.AddressTypeIPv4,
	}
	ipv6EndpointSlice := &v1discovery.EndpointSlice{
		AddressType: v1discovery.AddressTypeIPv6,
	}

	if endpointSlice == nil {
		return ipv4EndpointSlice, ipv6EndpointSlice
	}

	for _, endpnt := range endpointSlice.Endpoints {
		ipv4Endpoint := endpnt
		ipv4Endpoint.Addresses = []string{}
		ipv6Endpoint := endpnt
		ipv6Endpoint.Addresses = []string{}

		for _, address := range endpnt.Addresses {
			ipAddr := net.ParseIP(address)
			if ipAddr == nil {
				continue
			}

			if ipAddr.To4() != nil {
				ipv4Endpoint.Addresses = append(ipv4Endpoint.Addresses, ipAddr.String())

				continue
			}

			ipv6Endpoint.Addresses = append(ipv6Endpoint.Addresses, ipAddr.String())
		}

		if len(ipv4Endpoint.Addresses) > 0 {
			ipv4EndpointSlice.Endpoints = append(ipv4EndpointSlice.Endpoints, ipv4Endpoint)
		}

		if len(ipv6Endpoint.Addresses) > 0 {
			ipv6EndpointSlice.Endpoints = append(ipv6EndpointSlice.Endpoints, ipv6Endpoint)
		}
	}

	return ipv4EndpointSlice, ipv6EndpointSlice
}
