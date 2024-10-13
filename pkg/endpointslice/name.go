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
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
)

// GetEndpointSliceName concatenates the service name and the address type (ip family).
func GetEndpointSliceName(service *v1.Service, addressType v1discovery.AddressType) string {
	return fmt.Sprintf("%s-%s", service.GetName(), strings.ToLower(string(addressType)))
}
