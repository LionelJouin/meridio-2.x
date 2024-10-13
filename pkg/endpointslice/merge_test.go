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
	"reflect"
	"testing"

	"github.com/lioneljouin/meridio-experiment/pkg/endpointslice"
	v1core "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestMergeEndpointSlices(t *testing.T) {
	type args struct {
		endpointSliceA *v1discovery.EndpointSlice
		endpointSliceB *v1discovery.EndpointSlice
	}
	tests := []struct {
		name string
		args args
		want *v1discovery.EndpointSlice
	}{
		{
			name: "nil",
			args: args{
				endpointSliceA: nil,
				endpointSliceB: nil,
			},
			want: nil,
		},
		{
			name: "endpointSliceA: not nil, endpointSliceB: nil",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{}, stringPoint("10"))),
				endpointSliceB: nil,
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{}, stringPoint("10"))),
		},
		{
			name: "endpointSliceA: nil, endpointSliceB: not nil",
			args: args{
				endpointSliceA: nil,
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{}, stringPoint("10"))),
		},
		{
			name: "merge 1",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1", "2000::1"}, stringPoint("10"))),
		},
		{
			name: "merge 2",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("456", []string{"20.0.0.2"}, stringPoint("15")), newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1", "2000::1"}, stringPoint("10")), newEndpoint("456", []string{"20.0.0.2"}, stringPoint("15"))),
		},
		{
			name: "merge 3- nil zone ipv4",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, nil)),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1", "2000::1"}, stringPoint("10"))),
		},
		{
			name: "merge 4 - nil zone ipv6",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, nil)),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1", "2000::1"}, stringPoint("10"))),
		},
		{
			name: "merge 6 - empty ipv4 addresses",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{}, stringPoint("10"))),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
		},
		{
			name: "merge 7 - empty ipv6 addresses",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
		},
		{
			name: "merge 6 - empty ipv4",
			args: args{
				endpointSliceA: newEndpointSlice("abc", ""),
				endpointSliceB: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
		},
		{
			name: "merge 7 - empty ipv6 addresses",
			args: args{
				endpointSliceA: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
				endpointSliceB: newEndpointSlice("abc", ""),
			},
			want: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := endpointslice.MergeEndpointSlices(tt.args.endpointSliceA, tt.args.endpointSliceB); !equals(got, tt.want) {
				t.Errorf("MergeEndpointSlices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func equals(endpointSliceA *v1discovery.EndpointSlice, endpointSliceB *v1discovery.EndpointSlice) bool {
	if endpointSliceA == endpointSliceB {
		return true
	}

	if endpointSliceA == nil || endpointSliceB == nil {
		return false
	}

	if !reflect.DeepEqual(endpointSliceA.ObjectMeta, endpointSliceB.ObjectMeta) {
		return false
	}

	if endpointSliceA.AddressType != endpointSliceB.AddressType {
		return false
	}

	endpointsAMap := map[types.UID]v1discovery.Endpoint{}

	for _, endpnt := range endpointSliceA.Endpoints {
		endpointsAMap[endpnt.TargetRef.UID] = endpnt
	}

	for _, endpnt := range endpointSliceB.Endpoints {
		endpntA, exists := endpointsAMap[endpnt.TargetRef.UID]
		if !exists {
			return false
		}
		if !reflect.DeepEqual(endpntA, endpnt) {
			return false
		}
	}

	return true
}

func newEndpointSlice(name string, addressType v1discovery.AddressType, endpoints ...v1discovery.Endpoint) *v1discovery.EndpointSlice {
	return &v1discovery.EndpointSlice{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		AddressType: addressType,
		Endpoints:   endpoints,
	}
}

func newEndpoint(uid string, addresses []string, zone *string) v1discovery.Endpoint {
	return v1discovery.Endpoint{
		Addresses: addresses,
		Zone:      zone,
		TargetRef: &v1core.ObjectReference{
			UID: types.UID(uid),
		},
	}
}

func stringPoint(val string) *string {
	return &val
}

func TestSplitEndpointSlices(t *testing.T) {
	type args struct {
		endpointSlice *v1discovery.EndpointSlice
	}
	tests := []struct {
		name  string
		args  args
		want  *v1discovery.EndpointSlice
		want1 *v1discovery.EndpointSlice
	}{
		{
			name: "nil",
			args: args{
				endpointSlice: nil,
			},
			want:  newEndpointSlice("", v1discovery.AddressTypeIPv4),
			want1: newEndpointSlice("", v1discovery.AddressTypeIPv6),
		},
		{
			name: "duastack",
			args: args{
				endpointSlice: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1", "2000::1"}, stringPoint("10"))),
			},
			want:  newEndpointSlice("", v1discovery.AddressTypeIPv4, newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
			want1: newEndpointSlice("", v1discovery.AddressTypeIPv6, newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
		},
		{
			name: "ipv4 only",
			args: args{
				endpointSlice: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
			},
			want:  newEndpointSlice("", v1discovery.AddressTypeIPv4, newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10"))),
			want1: newEndpointSlice("", v1discovery.AddressTypeIPv6),
		},
		{
			name: "ipv6 only",
			args: args{
				endpointSlice: newEndpointSlice("abc", "", newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
			},
			want:  newEndpointSlice("", v1discovery.AddressTypeIPv4),
			want1: newEndpointSlice("", v1discovery.AddressTypeIPv6, newEndpoint("123", []string{"2000::1"}, stringPoint("10"))),
		},
		{
			name: "duastack 2",
			args: args{
				endpointSlice: newEndpointSlice("abc", "", newEndpoint("123", []string{"20.0.0.1", "2000::1"}, stringPoint("10")), newEndpoint("456", []string{"20.0.0.2", "2000::2"}, stringPoint("15"))),
			},
			want:  newEndpointSlice("", v1discovery.AddressTypeIPv4, newEndpoint("123", []string{"20.0.0.1"}, stringPoint("10")), newEndpoint("456", []string{"20.0.0.2"}, stringPoint("15"))),
			want1: newEndpointSlice("", v1discovery.AddressTypeIPv6, newEndpoint("123", []string{"2000::1"}, stringPoint("10")), newEndpoint("456", []string{"2000::2"}, stringPoint("15"))),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := endpointslice.SplitEndpointSlices(tt.args.endpointSlice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitEndpointSlices() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("SplitEndpointSlices() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
