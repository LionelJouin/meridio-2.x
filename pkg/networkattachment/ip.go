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
	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// GetIPs returns the IPs of a pod from a list of networks.
func GetIPs(pod v1.Pod, networks []*v1alpha1.Network) ([]string, error) {
	ips := []string{}

	for _, network := range networks {
		if network.NetworkAttachementAnnotation != nil {
			networkAttachmentElements := network.NetworkAttachementAnnotation.Value

			status, exists := pod.GetAnnotations()[network.NetworkAttachementAnnotation.StatusKey]
			if !exists {
				continue
			}

			newIPs, err := GetIPsFromNetworkAttachmentAnnotation(pod.Namespace, networkAttachmentElements, status)
			if err != nil {
				continue // todo: print error?
			}

			ips = append(ips, newIPs...)
		}
	}

	return ips, nil
}
