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

package endpointslice_test

import (
	"testing"

	"github.com/lioneljouin/meridio-experiment/pkg/endpointslice"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetEndpointSliceName(t *testing.T) {
	type args struct {
		service     *v1.Service
		addressType v1discovery.AddressType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1",
			args: args{
				service:     getService("test"),
				addressType: v1discovery.AddressTypeIPv4,
			},
			want: "test-ipv4",
		},
		{
			name: "1",
			args: args{
				service:     getService("test"),
				addressType: v1discovery.AddressTypeIPv6,
			},
			want: "test-ipv6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointslice.GetEndpointSliceName(tt.args.service, tt.args.addressType); got != tt.want {
				t.Errorf("GetEndpointSliceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getService(name string) *v1.Service {
	return &v1.Service{
		ObjectMeta: v1meta.ObjectMeta{
			Name: name,
		},
	}
}
