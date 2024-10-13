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

package bird_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/bird"
	"github.com/vishvananda/netlink"
)

var bfd = &bfdSpec{
	sw:         newBool(true),
	minTx:      "300ms",
	minRx:      "300ms",
	multiplier: newUint16(5),
}

var bgp = &bgpSpec{
	remotePort: newUint16(10179),
	localPort:  newUint16(10179),
	remoteASN:  newUint32(4248829953),
	localASN:   newUint32(8103),
	holdTime:   "24s",
	bfd:        bfd,
}

var gatewayIPv4BGPBFD = &gateway{
	name:     "gateway-v4-a-1",
	address:  "169.254.100.150",
	protocol: v1alpha1.BGP,
	bgp:      bgp,
	intf:     "eth0",
}

var gatewayIPv6BGPBFD = &gateway{
	name:     "gateway-v6-a-1",
	address:  "100:100::150",
	protocol: v1alpha1.BGP,
	bgp:      bgp,
	intf:     "eth0",
}

func TestBird_Configure(t *testing.T) {
	type fields struct{}
	type args struct {
		vips     []string
		gateways []bird.Gateway
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantConfig string
	}{
		{
			name: "empty",
			args: args{
				vips:     []string{},
				gateways: []bird.Gateway{},
			},
			wantErr:    false,
			wantConfig: emptyConfig,
		},
		{
			name: "empty",
			args: args{
				vips:     []string{"20.0.0.1/32", "40.0.0.150/32", "2000::1/128", "4000::150/128"},
				gateways: []bird.Gateway{},
			},
			wantErr:    false,
			wantConfig: ipv4AndIPv6VIPsConfig,
		},
		{
			name: "empty",
			args: args{
				vips:     []string{},
				gateways: []bird.Gateway{gatewayIPv4BGPBFD, gatewayIPv6BGPBFD},
			},
			wantErr:    false,
			wantConfig: ipv4AndIPv6BGPBFD,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := bird.New()
			b.ConfigFile = "./bird.conf"

			if err := b.Configure(context.TODO(), tt.args.vips, tt.args.gateways); (err != nil) != tt.wantErr {
				t.Errorf("Bird.Configure() error = %v, wantErr %v", err, tt.wantErr)
			}

			config, err := os.ReadFile(b.ConfigFile)
			if err != nil {
				t.Errorf("error reading bird config file = %v", err)
			}

			if string(config) != tt.wantConfig {
				t.Errorf("Bird.Configure() config = %v, want %v", string(config), tt.wantConfig)
			}

			err = os.Remove(b.ConfigFile)
			if err != nil {
				t.Errorf("error deleting bird config file = %v", err)
			}

			vips, err := getPolicyRoutes()
			if err != nil {
				t.Errorf("error getting policy routes = %v", err)
			}

			sort.Strings(vips)
			sort.Strings(tt.args.vips)

			if !reflect.DeepEqual(vips, tt.args.vips) {
				t.Errorf("Bird.Configure() policy routes = %v, want %v", vips, tt.args.vips)
			}
		})
	}
}

func getPolicyRoutes() ([]string, error) {
	rules, err := netlink.RuleListFiltered(netlink.FAMILY_ALL, &netlink.Rule{
		Table: 4096,
	}, netlink.RT_FILTER_TABLE)
	if err != nil {
		return nil, fmt.Errorf("failed to list rules: %w", err)
	}

	vips := []string{}
	for _, rule := range rules {
		vips = append(vips, rule.Src.String())
	}

	return vips, nil
}

var emptyConfig = `log "/var/log/bird.log" 20000 "/var/log/bird.log.backup" { debug, trace, info, remote, warning, error, auth, fatal, bug };
log stderr all;

protocol device {
}

filter gateway_routes {
	if ( net ~ [ 0.0.0.0/0 ] ) then accept;
	if ( net ~ [ 0::/0 ] ) then accept;
	if source = RTS_BGP then accept;
	else reject;
}

filter announced_routes {
	if ( net ~ [ 0.0.0.0/0 ] ) then reject;
	if ( net ~ [ 0::/0 ] ) then reject;
	if source = RTS_STATIC && dest != RTD_BLACKHOLE then accept;
	else reject;
}

template bgp BGP_TEMPLATE {
	debug {events, states, interfaces};
 	direct;
 	hold time 3;
	bfd on;
	graceful restart off;
	setkey off;
	ipv4 {
		import none;
		export none;
		next hop self; # advertise this router as next hop
	};
	ipv6 {
		import none;
		export none;
		next hop self; # advertise this router as next hop
	};
}

protocol kernel {
	ipv4 {
		import none;
		export filter gateway_routes;
	};
	kernel table 4096;
	merge paths on;
}

protocol kernel {
	ipv6 {
		import none;
		export filter gateway_routes;
	};
	kernel table 4096;
	merge paths on;
}

ipv4 table drop4;

ipv6 table drop6;

protocol kernel {
	ipv4 {
		table drop4;
		import none;
		export all;
	};
	kernel table 4097;
}

protocol kernel {
	ipv6 {
		table drop6;
		import none;
		export all;
	};
	kernel table 4097;
}

protocol static DROP4 {
	ipv4 { table drop4; preference 0; };
	route 0.0.0.0/0 blackhole {
		krt_metric=4294967295;
		igp_metric=4294967295;
	};
}

protocol static DROP6 {
	ipv6 { table drop6; preference 0; };
	route 0::/0 blackhole {
		krt_metric=4294967295;
		igp_metric=4294967295;
	};
}

protocol bfd {
	interface "*" {
	};
}`

var ipv4AndIPv6VIPsConfig = `log "/var/log/bird.log" 20000 "/var/log/bird.log.backup" { debug, trace, info, remote, warning, error, auth, fatal, bug };
log stderr all;

protocol device {
}

filter gateway_routes {
	if ( net ~ [ 0.0.0.0/0 ] ) then accept;
	if ( net ~ [ 0::/0 ] ) then accept;
	if source = RTS_BGP then accept;
	else reject;
}

filter announced_routes {
	if ( net ~ [ 0.0.0.0/0 ] ) then reject;
	if ( net ~ [ 0::/0 ] ) then reject;
	if source = RTS_STATIC && dest != RTD_BLACKHOLE then accept;
	else reject;
}

template bgp BGP_TEMPLATE {
	debug {events, states, interfaces};
 	direct;
 	hold time 3;
	bfd on;
	graceful restart off;
	setkey off;
	ipv4 {
		import none;
		export none;
		next hop self; # advertise this router as next hop
	};
	ipv6 {
		import none;
		export none;
		next hop self; # advertise this router as next hop
	};
}

protocol kernel {
	ipv4 {
		import none;
		export filter gateway_routes;
	};
	kernel table 4096;
	merge paths on;
}

protocol kernel {
	ipv6 {
		import none;
		export filter gateway_routes;
	};
	kernel table 4096;
	merge paths on;
}

ipv4 table drop4;

ipv6 table drop6;

protocol kernel {
	ipv4 {
		table drop4;
		import none;
		export all;
	};
	kernel table 4097;
}

protocol kernel {
	ipv6 {
		table drop6;
		import none;
		export all;
	};
	kernel table 4097;
}

protocol static DROP4 {
	ipv4 { table drop4; preference 0; };
	route 0.0.0.0/0 blackhole {
		krt_metric=4294967295;
		igp_metric=4294967295;
	};
}

protocol static DROP6 {
	ipv6 { table drop6; preference 0; };
	route 0::/0 blackhole {
		krt_metric=4294967295;
		igp_metric=4294967295;
	};
}

protocol static VIP4 {
	ipv4 { preference 110; };
	route 20.0.0.1/32 via "lo";
	route 40.0.0.150/32 via "lo";

}protocol static VIP6 {
	ipv6 { preference 110; };
	route 2000::1/128 via "lo";
	route 4000::150/128 via "lo";

}

protocol bfd {
	interface "*" {
	};
}`

var ipv4AndIPv6BGPBFD = `log "/var/log/bird.log" 20000 "/var/log/bird.log.backup" { debug, trace, info, remote, warning, error, auth, fatal, bug };
log stderr all;

protocol device {
}

filter gateway_routes {
	if ( net ~ [ 0.0.0.0/0 ] ) then accept;
	if ( net ~ [ 0::/0 ] ) then accept;
	if source = RTS_BGP then accept;
	else reject;
}

filter announced_routes {
	if ( net ~ [ 0.0.0.0/0 ] ) then reject;
	if ( net ~ [ 0::/0 ] ) then reject;
	if source = RTS_STATIC && dest != RTD_BLACKHOLE then accept;
	else reject;
}

template bgp BGP_TEMPLATE {
	debug {events, states, interfaces};
 	direct;
 	hold time 3;
	bfd on;
	graceful restart off;
	setkey off;
	ipv4 {
		import none;
		export none;
		next hop self; # advertise this router as next hop
	};
	ipv6 {
		import none;
		export none;
		next hop self; # advertise this router as next hop
	};
}

protocol kernel {
	ipv4 {
		import none;
		export filter gateway_routes;
	};
	kernel table 4096;
	merge paths on;
}

protocol kernel {
	ipv6 {
		import none;
		export filter gateway_routes;
	};
	kernel table 4096;
	merge paths on;
}

ipv4 table drop4;

ipv6 table drop6;

protocol kernel {
	ipv4 {
		table drop4;
		import none;
		export all;
	};
	kernel table 4097;
}

protocol kernel {
	ipv6 {
		table drop6;
		import none;
		export all;
	};
	kernel table 4097;
}

protocol static DROP4 {
	ipv4 { table drop4; preference 0; };
	route 0.0.0.0/0 blackhole {
		krt_metric=4294967295;
		igp_metric=4294967295;
	};
}

protocol static DROP6 {
	ipv6 { table drop6; preference 0; };
	route 0::/0 blackhole {
		krt_metric=4294967295;
		igp_metric=4294967295;
	};
}

protocol bgp 'gateway-v4-a-1' from BGP_TEMPLATE {
	interface "eth0";
	local port 10179 as 8103;
	neighbor 169.254.100.150 port 10179 as 4248829953;
	bfd {
		min rx interval 300ms;
		min tx interval 300ms;
		multiplier 5;
	};
	hold time 3;
	ipv4 {
		import filter gateway_routes;
		export filter announced_routes;
	};
}

protocol bgp 'gateway-v6-a-1' from BGP_TEMPLATE {
	interface "eth0";
	local port 10179 as 8103;
	neighbor 100:100::150 port 10179 as 4248829953;
	bfd {
		min rx interval 300ms;
		min tx interval 300ms;
		multiplier 5;
	};
	hold time 3;
	ipv6 {
		import filter gateway_routes;
		export filter announced_routes;
	};
}

protocol bfd {
	interface "*" {
	};
}`

type gateway struct {
	name     string
	address  string
	intf     string
	protocol v1alpha1.RoutingProtocol
	bgp      *bgpSpec
	static   *staticSpec
}

func (gw *gateway) GetName() string {
	return gw.name
}

func (gw *gateway) GetAddress() string {
	return gw.address
}

func (gw *gateway) GetInterface() string {
	return gw.intf
}

func (gw *gateway) GetProtocol() v1alpha1.RoutingProtocol {
	return gw.protocol
}

func (gw *gateway) GetBgpSpec() bird.BgpSpec {
	return gw.bgp
}

func (gw *gateway) GetStatic() bird.StaticSpec {
	return gw.static
}

type bgpSpec struct {
	remoteASN  *uint32
	localASN   *uint32
	bfd        *bfdSpec
	holdTime   string
	remotePort *uint16
	localPort  *uint16
}

func (bgps *bgpSpec) GetRemoteASN() *uint32 {
	return bgps.remoteASN
}

func (bgps *bgpSpec) GetLocalASN() *uint32 {
	return bgps.localASN
}

func (bgps *bgpSpec) GetBfdSpec() bird.BfdSpec {
	return bgps.bfd
}

func (bgps *bgpSpec) GetHoldTime() string {
	return bgps.holdTime
}

func (bgps *bgpSpec) GetRemotePort() *uint16 {
	return bgps.remotePort
}

func (bgps *bgpSpec) GetLocalPort() *uint16 {
	return bgps.localPort
}

type staticSpec struct {
	bfd *bfdSpec
}

func (ss *staticSpec) GetBfdSpec() bird.BfdSpec {
	return ss.bfd
}

type bfdSpec struct {
	sw         *bool
	minTx      string
	minRx      string
	multiplier *uint16
}

func (bfds *bfdSpec) GetSwitch() *bool {
	return bfds.sw
}

func (bfds *bfdSpec) GetMinTx() string {
	return bfds.minTx
}

func (bfds *bfdSpec) GetMinRx() string {
	return bfds.minRx
}

func (bfds *bfdSpec) GetMultiplier() *uint16 {
	return bfds.multiplier
}

func newBool(val bool) *bool {
	return &val
}

func newUint16(val uint16) *uint16 {
	return &val
}

func newUint32(val uint32) *uint32 {
	return &val
}
