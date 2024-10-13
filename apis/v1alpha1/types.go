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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayRouter is a specification for a GatewayRouter resource.
type GatewayRouter struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the GatewayRouter.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec GatewayRouterSpec `json:"spec"`

	// Most recently observed status of the GatewayRouter.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status GatewayRouterStatus `json:"status"`
}

// GatewayRouterSpec defines the desired state of GatewayRouter.
type GatewayRouterSpec struct {
	// Address of the Gateway Router
	Address string `json:"address"`

	// Interface used to access the Gateway Router
	Interface string `json:"interface"`

	// The routing choice between the Gateway Router and Attractor FrontEnds.
	// +optional
	Protocol RoutingProtocol `json:"protocol,omitempty"`

	// Parameters to set up the BGP session to specified Address.
	// If the Protocol is static, this property must be empty.
	// If the Protocol is bgp, the minimal parameters to be defined in bgp properties
	// are RemoteASN and LocalASN
	// +optional
	Bgp BgpSpec `json:"bgp,omitempty"`

	// Parameters to work with the static routing configured on the Gateway Router with specified Address.
	// If the Protocol is bgp, this property must be empty.
	// +optional
	Static StaticSpec `json:"static,omitempty"`
}

// RoutingProtocol represents the routing protocol used in a gateway router.
// +enum
type RoutingProtocol string

const (
	// BGP, Border Gateway Protocol.
	BGP RoutingProtocol = "BGP"
	// Static Routing.
	Static RoutingProtocol = "Static"
)

// BgpSpec defines the parameters to set up a BGP session.
type BgpSpec struct {
	// The ASN number of the Gateway Router
	//nolint:tagliatelle
	RemoteASN *uint32 `json:"remoteASN,omitempty"`

	// The ASN number of the system where the Attractor FrontEnds locates
	//nolint:tagliatelle
	LocalASN *uint32 `json:"localASN,omitempty"`

	// BFD monitoring of BGP session.
	// +optional
	BFD BfdSpec `json:"bfd,omitempty"`

	// Hold timer of the BGP session. Please refere to BGP material to understand what this implies.
	// The value must be a valid duration format. For example, 90s, 1m, 1h.
	// The duration will be rounded by second
	// Minimum duration is 3s.
	// +optional
	HoldTime string `json:"holdTime,omitempty"`

	// BGP listening port of the Gateway Router.
	// +optional
	RemotePort *uint16 `json:"remotePort,omitempty"`

	// BGP listening port of the Attractor FrontEnds.
	// +optional
	LocalPort *uint16 `json:"localPort,omitempty"`

	// BGP authentication (RFC2385).
	// +optional
	Auth *BgpAuth `json:"auth,omitempty"`
}

// StaticSpec defines the parameters to set up static routes.
type StaticSpec struct {
	// BFD monitoring of Static session.
	// +optional
	BFD BfdSpec `json:"bfd,omitempty"`
}

// Bfd defines the parameters to configure the BFD session.
// The static gateway routers shares the same interface shall define the same bfd configuration.
type BfdSpec struct {
	// BFD monitoring.
	// Valid values are:
	// - false: no BFD monitoring;
	// - true: turns on the BFD monitoring.
	// When left empty, there is no BFD monitoring.
	// +optional
	Switch *bool `json:"switch,omitempty"`

	// Min-tx timer of bfd session. Please refere to BFD material to understand what this implies.
	// The value must be a valid duration format. For example, 300ms, 90s, 1m, 1h.
	// The duration will be rounded by millisecond.
	// +optional
	MinTx string `json:"minTx,omitempty"`

	// Min-rx timer of bfd session. Please refere to BFD material to understand what this implies.
	// The value must be a valid duration format. For example, 300ms, 90s, 1m, 1h.
	// The duration will be rounded by millisecond.
	// +optional
	MinRx string `json:"minRx,omitempty"`

	// Multiplier of bfd session.
	// When this number of bfd packets failed to receive, bfd session will go down.
	// +optional
	Multiplier *uint16 `json:"multiplier,omitempty"`
}

// BgpAuth defines the parameters to configure BGP authentication.
type BgpAuth struct {
	// Name of the BGP authentication key, used internally as a reference.
	// KeyName is a key in the data section of a Secret. The associated value in
	// the Secret is the password (pre-shared key) to be used for authentication.
	// Must consist of alphanumeric characters, ".", "-" or "_".
	KeyName string `json:"keyName,omitempty"`

	// Name of the kubernetes Secret containing the password (pre-shared key)
	// that can be looked up based on KeyName.
	// Must be a valid lowercase RFC 1123 subdomain. (Must consist of lower case alphanumeric
	// characters, '-' or '.', and must start and end with an alphanumeric character.)
	KeySource string `json:"keySource,omitempty"`
}

// GatewayRouterStatus is the status for a GatewayRouter resource.
type GatewayRouterStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayRouterList is a list of GatewayRouter resources.
type GatewayRouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GatewayRouter `json:"items"`
}

// Network represents a single network, its way to attach it, and the way it should be mounted to
// the endpoints and proxy pods.
type Network struct {
	// Name of the network.
	Name string `json:"name,omitempty"`

	// NetworkAttachementAnnotation represents a network attached via an annotation.
	// +optional
	NetworkAttachementAnnotation *NetworkAttachementAnnotation `json:"networkAttachementAnnotation,omitempty"`
}

// NetworkAttachementAnnotation represents a network attached via an annotation.
type NetworkAttachementAnnotation struct {
	// Key of the network attachement (e.g.: k8s.v1.cni.cncf.io/networks).
	Key string `json:"key,omitempty"`

	// StatusKey of the network attachement status (e.g.: k8s.v1.cni.cncf.io/network-status).
	StatusKey string `json:"statusKey,omitempty"`

	// Value added for the "Key" (e.g.: [{"name":"macvlan-vlan-100","interface":"macvlan-100"}]).
	Value string `json:"value,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L34Route is a specification for a L34Route resource.
type L34Route struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the L34Route.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec L34RouteSpec `json:"spec"`

	// Most recently observed status of the L34Route.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status L34RouteStatus `json:"status"`
}

// +enum
type TransportProtocol string

const (
	// TCP represents the layer 4 protocol.
	TCP TransportProtocol = "TCP"
	// UDP represents the layer 4 protocol.
	UDP TransportProtocol = "UDP"
	// SCTP represents the layer 4 protocol.
	SCTP TransportProtocol = "SCTP"
)

// L34RouteSpec is the spec for a L34Route resource.
type L34RouteSpec struct {
	gatewayapiv1.CommonRouteSpec `json:",inline"`

	// BackendRefs defines the backend(s) where matching requests should be
	// sent. If unspecified or invalid (refers to a non-existent resource or a
	// Service with no endpoints), the underlying implementation MUST actively
	// reject connection attempts to this backend. Connection rejections must
	// respect weight; if an invalid backend is requested to have 80% of
	// connections, then 80% of connections must be rejected instead.
	//
	// Support: Core for Kubernetes Service
	//
	// Support: Extended for Kubernetes ServiceImport
	//
	// Support: Implementation-specific for any other resource
	//
	// Support for weight: Extended
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []gatewayapiv1.BackendRef `json:"backendRefs,omitempty"`

	// Destination CIDRs that this L34Route will send traffic to.
	// The destination CIDRs should not have overlaps.
	//nolint:tagliatelle
	DestinationCIDRs []string `json:"destinationCIDRs"`

	// Source CIDRs allowed in the L34Route.
	// The source CIDRs should not have overlaps.
	//nolint:tagliatelle
	SourceCIDRs []string `json:"sourceCIDRs,omitempty"`

	// Source port ranges allowed in the L34Route.
	// The ports should not have overlaps.
	// Ports can be defined by:
	// - a single port, such as 3000;
	// - a port range, such as 3000-4000;
	// - "any", which is equivalent to port range 0-65535.
	SourcePorts []string `json:"sourcePorts,omitempty"`

	// Destination port ranges allowed in the L34Route.
	// The ports should not have overlaps.
	// Ports can be defined by:
	// - a single port, such as 3000;
	// - a port range, such as 3000-4000;
	// - "any", which is equivalent to port range 0-65535.
	DestinationPorts []string `json:"destinationPorts,omitempty"`

	// Protocols allowed in this L34Route.
	// The protocols should not have overlaps.
	Protocols []TransportProtocol `json:"protocols"`

	// Priority of the L34Route
	Priority int32 `json:"priority"`

	// ByteMatches matches bytes in the L4 header in the L34Route.
	// +optional
	ByteMatches []string `json:"byteMatches,omitempty"`
}

// L34RouteStatus is the status for a L34Route resource.
type L34RouteStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// L34RouteList is a list of L34Route resources.
type L34RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []L34Route `json:"items"`
}
