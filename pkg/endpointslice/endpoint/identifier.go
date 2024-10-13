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

package endpoint

import (
	"strconv"

	v1discovery "k8s.io/api/discovery/v1"
)

// GetIdentifier returns the identifier in an endpont.
func GetIdentifier(endpoint v1discovery.Endpoint) *int {
	if endpoint.Zone == nil {
		return nil
	}

	id, err := strconv.Atoi(*endpoint.Zone)
	if err != nil {
		return nil
	}

	return &id
}

// SetIdentifier sets the identifier in an endpont.
func SetIdentifier(endpoint v1discovery.Endpoint, identifier int) *v1discovery.Endpoint {
	id := strconv.Itoa(identifier)
	endpoint.Zone = &id

	return &endpoint
}
