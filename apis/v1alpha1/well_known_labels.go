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

package v1alpha1

const (
	// LabelDummmySericeSelector is used as a dummy service selector for Kubernetes to
	// not find any application pod and then keep the kubernetes endpointslice empty. This selector
	// will be ignored by the endpointslice controllers of this PoC.
	LabelDummmySericeSelector = "meridio-experiment/dummy-service-selector"

	// LabelServiceMaxEndpoints defines the maximum number of endpoints that a
	// service can handle.
	LabelServiceMaxEndpoints = "meridio-experiment/service-max-endpoints"

	// PodSelectedNetworks represents the networks that must be in the pods selected by the services.
	PodSelectedNetworks = "meridio-experiment/networks"

	PodSelectedNetworkSubnets = "meridio-experiment/network-subnets"
)
