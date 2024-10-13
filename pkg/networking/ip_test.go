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

package networking_test

import (
	"net"
	"reflect"
	"testing"

	"github.com/lioneljouin/meridio-experiment/pkg/networking"
	"go.uber.org/goleak"
)

func TestBroadcastFromIPNet(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	type args struct {
		ipNet *net.IPNet
	}

	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{
			name: "ipv4: 172.16.1.0/24",
			args: args{
				ipNet: &net.IPNet{
					IP:   net.IPv4(172, 16, 1, 0).To4(),
					Mask: net.IPv4Mask(255, 255, 255, 0),
				},
			},
			want: net.IPv4(172, 16, 1, 255).To4(),
		},
		{
			name: "ipv4: 0.0.0.0/0",
			args: args{
				ipNet: &net.IPNet{
					IP:   net.IPv4(0, 0, 0, 0).To4(),
					Mask: net.IPv4Mask(0, 0, 0, 0),
				},
			},
			want: net.IPv4(255, 255, 255, 255).To4(),
		},
		{
			name: "ipv4: 255.255.255.255/32",
			args: args{
				ipNet: &net.IPNet{
					IP:   net.IPv4(255, 255, 255, 255).To4(),
					Mask: net.IPv4Mask(255, 255, 255, 255),
				},
			},
			want: net.IPv4(255, 255, 255, 255).To4(),
		},
		{
			name: "ipv6: 2000::/64",
			args: args{
				ipNet: &net.IPNet{
					IP:   net.IP{0x20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					Mask: net.IPMask(net.ParseIP("ffff:ffff:ffff:ffff::")),
				},
			},
			want: net.IP{0x20, 0, 0, 0, 0, 0, 0, 0, 255, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			name: "ipv6: ::/0",
			args: args{
				ipNet: &net.IPNet{
					IP:   net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					Mask: net.IPMask(net.ParseIP("::")),
				},
			},
			want: net.IP{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
		{
			name: "ipv6: ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128",
			args: args{
				ipNet: &net.IPNet{
					IP:   net.IP{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
					Mask: net.IPMask(net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")),
				},
			},
			want: net.IP{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := networking.BroadcastFromIPNet(tt.args.ipNet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BroadcastFromIPNet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextIP(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	type args struct {
		ip net.IP
	}

	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{
			name: "ipv4: 172.16.1.0",
			args: args{
				ip: net.IPv4(172, 16, 1, 0).To4(),
			},
			want: net.IPv4(172, 16, 1, 1).To4(),
		},
		{
			name: "ipv4: 0.0.0.0",
			args: args{
				ip: net.IPv4(0, 0, 0, 0).To4(),
			},
			want: net.IPv4(0, 0, 0, 1).To4(),
		},
		{
			name: "ipv4: 172.255.255.255",
			args: args{
				ip: net.IPv4(172, 255, 255, 255).To4(),
			},
			want: net.IPv4(173, 0, 0, 0).To4(),
		},
		{
			name: "ipv4: 255.255.255.255",
			args: args{
				ip: net.IPv4(255, 255, 255, 255).To4(),
			},
			want: net.IPv4(0, 0, 0, 0).To4(),
		},
		{
			name: "ipv6: 2000::",
			args: args{
				ip: net.IP{0x20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			want: net.IP{0x20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name: "ipv6: ::",
			args: args{
				ip: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			},
			want: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
		},
		{
			name: "ipv6: 2000:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			args: args{
				ip: net.IP{0x20, 0x1, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			},
			want: net.IP{0x20, 0x2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
		{
			name: "ipv6: ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
			args: args{
				ip: net.IP{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255},
			},
			want: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := networking.NextIP(tt.args.ip); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NextIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
