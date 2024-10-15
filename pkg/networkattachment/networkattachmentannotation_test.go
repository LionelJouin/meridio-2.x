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

	"github.com/lioneljouin/meridio-experiment/pkg/networkattachment"
)

func TestGetIPsFromNetworkAttachmentAnnotation(t *testing.T) {
	type args struct {
		namespace string
		networks  string
		status    string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "no networks",
			args: args{
				namespace: "default",
				networks:  "",
				status:    "",
			},
			want:    []string{},
			wantErr: false,
		},
		{
			name: "ipv4",
			args: args{
				namespace: "default",
				networks:  `[{"name":"macvlan-nad","interface":"net1"}]`,
				status:    `[{"name":"kindnet","interface":"eth0","ips":["10.244.1.5","fd00:10:244:1::5"],"mac":"46:db:9f:b0:46:fe","default":true,"dns":{},"gateway":["u003cnilu003e","u003cnilu003e"]},{"name":"default/macvlan-nad","interface":"net1","ips":["169.255.100.2"],"mac":"0a:45:14:32:ec:f2","dns":{}}]`,
			},
			want:    []string{"169.255.100.2"},
			wantErr: false,
		},
		{
			name: "ipv6",
			args: args{
				namespace: "default",
				networks:  `[{"name":"macvlan-nad","interface":"net1"}]`,
				status:    `[{"name":"kindnet","interface":"eth0","ips":["10.244.1.5","fd00:10:244:1::5"],"mac":"46:db:9f:b0:46:fe","default":true,"dns":{},"gateway":["u003cnilu003e","u003cnilu003e"]},{"name":"default/macvlan-nad","interface":"net1","ips":["fd00::1"],"mac":"0a:45:14:32:ec:f2","dns":{}}]`,
			},
			want:    []string{"fd00::1"},
			wantErr: false,
		},
		{
			name: "ipv4 and ipv6",
			args: args{
				namespace: "default",
				networks:  `[{"name":"macvlan-nad","interface":"net1"}]`,
				status:    `[{"name":"kindnet","interface":"eth0","ips":["10.244.1.5","fd00:10:244:1::5"],"mac":"46:db:9f:b0:46:fe","default":true,"dns":{},"gateway":["u003cnilu003e","u003cnilu003e"]},{"name":"default/macvlan-nad","interface":"net1","ips":["169.255.100.2","fd00::1"],"mac":"0a:45:14:32:ec:f2","dns":{}}]`,
			},
			want:    []string{"169.255.100.2", "fd00::1"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := networkattachment.GetIPsFromNetworkAttachmentAnnotation(tt.args.namespace, tt.args.networks, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIPsFromNetworkAttachmentAnnotation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIPsFromNetworkAttachmentAnnotation() = %v, want %v", got, tt.want)
			}
		})
	}
}
