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

package networkannotation

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
)

const NetworkAnnotationKey = "meridio-experiment/network-configuration"

// key is the gateway name
type NetworkConfiguration map[string]*NetworkRoute

type NetworkRoute struct {
	VIPsV4     []string `json:"vipsV4"`
	VIPsV6     []string `json:"vipsV6"`
	GatewaysV4 []string `json:"gatewaysV4"`
	GatewaysV6 []string `json:"gatewaysV6"`
	TableID    int      `json:"tableID"`
}

// GetNetworkConfiguration ...
func GetNetworkConfiguration(
	pod *v1.Pod,
) (NetworkConfiguration, error) {
	var currentNetworkConfiguration NetworkConfiguration

	networks, exists := pod.GetAnnotations()[NetworkAnnotationKey]
	if exists {
		err := json.Unmarshal([]byte(networks), &currentNetworkConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to json.Unmarshal Network Annotation: %w", err)
		}
	}

	return currentNetworkConfiguration, nil
}
