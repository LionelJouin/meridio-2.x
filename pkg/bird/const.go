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

package bird

const (
	defaultBGPHoldTime          = 3
	defaultLocalASN      uint32 = 8103
	defaultLocalPort     uint16 = 10179
	defaultRemoteASN     uint32 = 4248829953
	defaultRemotePort    uint16 = 10179
	defaultKernelTableID        = 4096
	defaultLogFileSize          = 20000
)

// 0: kernel table ID
// 1: kernel table ID
const baseConfig = `protocol device {
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
	kernel table %d;
	merge paths on;
}

protocol kernel {
	ipv6 {
		import none;
		export filter gateway_routes;
	};
	kernel table %d;
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
}`

// 0: name of the static protocol
// 1: IP Family
// 2: routes
const vipsTemplate = `protocol static %s {
	%s { preference 110; };
%s
}`

// 0: CIDR of the route
const vipRouteTemplate = "\troute %s via \"lo\";\n"

// Represents the BGP protocol
// 0: Name of the gateway
// 1: Interface used for the gateway
// 2: Local Port
// 3: Local ASN
// 4: Remote IP (Gateway IP)
// 5: Remote Port
// 6: Remote ASN
// 7: BFD
// 8: Hold Time
// 9: IP Family
const bgpTemplate = `protocol bgp '%s' from BGP_TEMPLATE {
	interface "%s";
	local port %d as %d;
	neighbor %s port %d as %d;
	%s
	hold time %d;
	%s {
		import filter gateway_routes;
		export filter announced_routes;
	};
}`

const bfdTemplate = `protocol bfd {
	interface "*" {
	};
}`

// 0: bfd properties
const bgpBfdTemplate = `bfd {%s};`
