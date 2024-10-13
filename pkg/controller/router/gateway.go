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

package router

import (
	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/bird"
)

func newGateway(gw *v1alpha1.GatewayRouter) *gateway {
	protocol := gw.Spec.Protocol
	if protocol == "" {
		protocol = v1alpha1.BGP
	}

	newGw := &gateway{
		name:     gw.GetName(),
		intf:     gw.Spec.Interface,
		address:  gw.Spec.Address,
		protocol: protocol,
	}

	switch newGw.GetProtocol() {
	case v1alpha1.BGP:
		newGw.bgp = &bgpSpec{
			remoteASN:  gw.Spec.Bgp.RemoteASN,
			localASN:   gw.Spec.Bgp.LocalASN,
			holdTime:   gw.Spec.Bgp.HoldTime,
			remotePort: gw.Spec.Bgp.RemotePort,
			localPort:  gw.Spec.Bgp.LocalPort,
			bfd: &bfdSpec{
				sw:         gw.Spec.Bgp.BFD.Switch,
				minTx:      gw.Spec.Bgp.BFD.MinTx,
				minRx:      gw.Spec.Bgp.BFD.MinRx,
				multiplier: gw.Spec.Bgp.BFD.Multiplier,
			},
		}
	case v1alpha1.Static:
		newGw.static = &staticSpec{
			bfd: &bfdSpec{
				sw:         gw.Spec.Bgp.BFD.Switch,
				minTx:      gw.Spec.Bgp.BFD.MinTx,
				minRx:      gw.Spec.Bgp.BFD.MinRx,
				multiplier: gw.Spec.Bgp.BFD.Multiplier,
			},
		}
	}

	return newGw
}

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
