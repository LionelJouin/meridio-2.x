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

import "github.com/lioneljouin/meridio-experiment/apis/v1alpha1"

// GatewaySpec defines the desired state of Gateway.
type Gateway interface {
	GetName() string

	// Address of the Gateway Router
	GetAddress() string

	// Interface used to access the Gateway Router
	GetInterface() string

	GetProtocol() v1alpha1.RoutingProtocol

	// Parameters to set up the BGP session to specified Address.
	// If the Protocol is static, this property must be empty.
	// If the Protocol is bgp, the minimal parameters to be defined in bgp properties
	// are RemoteASN and LocalASN
	GetBgpSpec() BgpSpec

	// Parameters to work with the static routing configured on the Gateway Router with specified Address.
	// If the Protocol is bgp, this property must be empty.
	GetStatic() StaticSpec
}

// BgpSpec defines the parameters to set up a BGP session.
type BgpSpec interface {
	// The ASN number of the Gateway Router
	GetRemoteASN() *uint32

	// The ASN number of the system where the Attractor FrontEnds locates
	GetLocalASN() *uint32

	// BFD monitoring of BGP session.
	GetBfdSpec() BfdSpec

	// Hold timer of the BGP session. Please refere to BGP material to understand what this implies.
	// The value must be a valid duration format. For example, 90s, 1m, 1h.
	// The duration will be rounded by second
	// Minimum duration is 3s.
	GetHoldTime() string

	// BGP listening port of the Gateway Router.
	GetRemotePort() *uint16

	// BGP listening port of the Attractor FrontEnds.
	GetLocalPort() *uint16
}

// StaticSpec defines the parameters to set up static routes.
type StaticSpec interface {
	// BFD monitoring of Static session.
	GetBfdSpec() BfdSpec
}

// Bfd defines the parameters to configure the BFD session.
// The static gateways shares the same interface shall define the same bfd configuration.
type BfdSpec interface {
	// BFD monitoring.
	// Valid values are:
	// - false: no BFD monitoring;
	// - true: turns on the BFD monitoring.
	// When left empty, there is no BFD monitoring.
	GetSwitch() *bool

	// Min-tx timer of bfd session. Please refere to BFD material to understand what this implies.
	// The value must be a valid duration format. For example, 300ms, 90s, 1m, 1h.
	// The duration will be rounded by millisecond.
	GetMinTx() string

	// Min-rx timer of bfd session. Please refere to BFD material to understand what this implies.
	// The value must be a valid duration format. For example, 300ms, 90s, 1m, 1h.
	// The duration will be rounded by millisecond.
	GetMinRx() string

	// Multiplier of bfd session.
	// When this number of bfd packets failed to receive, bfd session will go down.
	GetMultiplier() *uint16
}
