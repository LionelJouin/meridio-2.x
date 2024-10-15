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
	"fmt"

	netdefv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

// GetIPsFromNetworkAttachmentAnnotation returns the IPs of a specifc network
// represented in the pod network attachment annotation.
func GetIPsFromNetworkAttachmentAnnotation(namespace string, networks string, status string) ([]string, error) {
	var networkSelectionElements []netdefv1.NetworkSelectionElement

	var networkStatuses []netdefv1.NetworkStatus

	if networks != "" {
		err := json.Unmarshal([]byte(networks), &networkSelectionElements)
		if err != nil {
			return nil, fmt.Errorf("failed to json.Unmarshal Network Attachment elements: %w", err)
		}
	}

	if status != "" {
		err := json.Unmarshal([]byte(status), &networkStatuses)
		if err != nil {
			return nil, fmt.Errorf("failed to json.Unmarshal Network Attachment status: %w", err)
		}
	}

	networkSelectionElementsMap := map[string]struct{}{}

	for _, networkSelectionElement := range networkSelectionElements {
		name := fmt.Sprintf("%s/%s", namespace, networkSelectionElement.Name)

		if networkSelectionElement.Namespace != "" {
			name = fmt.Sprintf("%s/%s", networkSelectionElement.Namespace, networkSelectionElement.Name)
		}

		networkSelectionElementsMap[name] = struct{}{}
	}

	ips := []string{}

	for _, netStatus := range networkStatuses {
		currentStatus := netStatus

		_, exists := networkSelectionElementsMap[currentStatus.Name]
		if exists {
			ips = append(ips, currentStatus.IPs...)
		}
	}

	return ips, nil
}
