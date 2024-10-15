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

package network

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/pkg/networkannotation"
	v1 "k8s.io/api/core/v1"
)

type Annotater struct {
	MinTableID int
	MaxTableID int
}

// SetNetworkAnnotation modifies the network annotation of the pod to correspond to
// the stateless load balancer requirement (VIP + policy routes) with the parameters (VIPs, Gateways).
func (a *Annotater) SetNetworkAnnotation(
	pod *v1.Pod,
	newNetworkConfiguration networkannotation.NetworkConfiguration,
) (*v1.Pod, error) {
	newPod := pod.DeepCopy()

	if len(newNetworkConfiguration) == 0 {
		delete(newPod.Annotations, networkannotation.NetworkAnnotationKey)

		return newPod, nil
	}

	var currentNetworkConfiguration networkannotation.NetworkConfiguration

	if newPod.GetAnnotations() == nil {
		newPod.Annotations = map[string]string{}
	} else {
		networks, exists := newPod.GetAnnotations()[networkannotation.NetworkAnnotationKey]
		if exists {
			err := json.Unmarshal([]byte(networks), &currentNetworkConfiguration)
			if err != nil {
				return nil, fmt.Errorf("failed to json.Unmarshal Network Annotation: %w", err)
			}
		}
	}

	mergedNetworkConfiguration, err := a.getNetworkAnnotation(
		currentNetworkConfiguration,
		newNetworkConfiguration,
	)
	if err != nil {
		return nil, err
	}

	mergedNetworkConfigurationJSON, err := json.Marshal(mergedNetworkConfiguration)
	if err != nil {
		return nil, fmt.Errorf("failed to json.Marshal mergedNetworkConfiguration: %w", err)
	}

	newPod.GetAnnotations()[networkannotation.NetworkAnnotationKey] = string(mergedNetworkConfigurationJSON)

	return newPod, nil
}

func (a *Annotater) getNetworkAnnotation(
	currentNetworkConfiguration networkannotation.NetworkConfiguration,
	newNetworkConfiguration networkannotation.NetworkConfiguration,
) (networkannotation.NetworkConfiguration, error) {
	reservedTableIDs := map[int]struct{}{}

	// Check still existing network configuration
	for gatewayName := range currentNetworkConfiguration {
		newNC, exists := newNetworkConfiguration[gatewayName]
		if exists {
			reservedTableIDs[newNC.TableID] = struct{}{}
			continue
		}
	}

	// Build new network configuration
	for gatewayName, newNC := range newNetworkConfiguration {
		currentNC, exists := currentNetworkConfiguration[gatewayName]
		if exists {
			newNC.TableID = currentNC.TableID
			continue
		}

		newNC.TableID = a.getFreeTableID(reservedTableIDs)
		if newNC.TableID == -1 {
			return nil, errors.New("no more table ID available")
		}

		reservedTableIDs[newNC.TableID] = struct{}{}
	}

	return newNetworkConfiguration, nil
}

func (a *Annotater) getFreeTableID(
	reservedTableIDs map[int]struct{},
) int {
	for i := a.MinTableID; i <= a.MaxTableID; i++ {
		_, exists := reservedTableIDs[i]
		if !exists {
			return i
		}
	}

	return -1
}
