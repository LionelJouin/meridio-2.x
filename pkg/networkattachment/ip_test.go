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

package networkattachment_test

import (
	"reflect"
	"testing"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/networkattachment"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetIPs(t *testing.T) {
	type args struct {
		pod      v1.Pod
		networks []*v1alpha1.Network
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "no-ip-no-networks",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
					},
				},
				networks: []*v1alpha1.Network{},
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "no-ip-networks",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
					},
				},
				networks: []*v1alpha1.Network{
					{
						Name: "macvlan-nad",
						NetworkAttachementAnnotation: &v1alpha1.NetworkAttachementAnnotation{
							Key:       "k8s.v1.cni.cncf.io/networks",
							StatusKey: "k8s.v1.cni.cncf.io/network-status",
							Value:     `[{"name": "macvlan-nad","interface": "vlan-100"}]`,
						},
					},
				},
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "ip-no-networks",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Annotations: map[string]string{
							"k8s.v1.cni.cncf.io/network-status": `[{"name":"default/macvlan-nad","interface":"net1","ips":["fd00::1"],"mac":"0a:45:14:32:ec:f2","dns":{}}]`,
						},
					},
				},
				networks: []*v1alpha1.Network{},
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "single-ipv4-network-attachement-annotation",
			args: args{
				pod: v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Annotations: map[string]string{
							"k8s.v1.cni.cncf.io/network-status": `[{"name":"default/macvlan-nad","interface":"net1","ips":["192.168.1.100"],"mac":"0a:45:14:32:ec:f2","dns":{}}]`,
						},
					},
				},
				networks: []*v1alpha1.Network{
					{
						Name: "macvlan-nad",
						NetworkAttachementAnnotation: &v1alpha1.NetworkAttachementAnnotation{
							Key:       "k8s.v1.cni.cncf.io/networks",
							StatusKey: "k8s.v1.cni.cncf.io/network-status",
							Value:     `[{"name": "macvlan-nad","interface": "vlan-100"}]`,
						},
					},
				},
			},
			want:    []string{"192.168.1.100"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := networkattachment.GetIPs(tt.args.pod, tt.args.networks)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}
