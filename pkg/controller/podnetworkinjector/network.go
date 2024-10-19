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

package podnetworkinjector

import (
	"context"
	"fmt"
	"net"

	"github.com/lioneljouin/meridio-experiment/pkg/log"
	"github.com/lioneljouin/meridio-experiment/pkg/networkannotation"
	"github.com/vishvananda/netlink"
)

type route struct {
	VIPsV4     map[string]*net.IPNet
	VIPsV6     map[string]*net.IPNet
	GatewaysV4 map[string]net.IP
	GatewaysV6 map[string]net.IP
}

func NetworkConfigurationByTableID(networkConfiguration networkannotation.NetworkConfiguration) map[int]*route {
	networkConfigurationByTableID := map[int]*route{}

	for _, networkConfig := range networkConfiguration {
		networkConfigurationByTableID[networkConfig.TableID] = &route{
			VIPsV4:     vipsToIPNet(networkConfig.VIPsV4, true),
			VIPsV6:     vipsToIPNet(networkConfig.VIPsV6, false),
			GatewaysV4: gatewaysToIP(networkConfig.GatewaysV4, true),
			GatewaysV6: gatewaysToIP(networkConfig.GatewaysV6, false),
		}
	}

	return networkConfigurationByTableID
}

func vipsToIPNet(vips []string, isIPv4 bool) map[string]*net.IPNet {
	vipIPNets := map[string]*net.IPNet{}

	for _, vip := range vips {
		ip, ipNet, err := net.ParseCIDR(vip)
		if err != nil || (isIPv4 && ip.To4() == nil) || (!isIPv4 && ip.To4() != nil) {
			continue
		}

		ipNet.IP = ip
		vipIPNets[ipNet.String()] = ipNet
	}

	return vipIPNets
}

func gatewaysToIP(gateways []string, isIPv4 bool) map[string]net.IP {
	gatewayIPs := map[string]net.IP{}

	for _, gateway := range gateways {
		ip := net.ParseIP(gateway)
		if ip == nil || (isIPv4 && ip.To4() == nil) || (!isIPv4 && ip.To4() != nil) {
			return nil
		}

		gatewayIPs[ip.String()] = ip
	}

	return gatewayIPs
}

func (c *Controller) configureRouting(ctx context.Context, networkConfig map[int]*route) error {
	rules, err := netlink.RuleList(netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to RuleList: %w", err)
	}

	for _, rule := range rules {
		if rule.Table < c.MinTableID || rule.Table > c.MaxTableID {
			continue
		}

		if rule.Src == nil {
			continue
		}

		route, exists := networkConfig[rule.Table]
		if !exists {
			err := deleteTable(rule.Table)
			if err != nil {
				log.FromContextOrGlobal(ctx).Error(err, "failed to deleteTable")
			}

			continue
		}

		_, existsV4 := route.VIPsV4[rule.Src.String()]
		if existsV4 {
			delete(route.VIPsV4, rule.Src.String())
			continue
		}
		_, existsV6 := route.VIPsV6[rule.Src.String()]
		if existsV6 {
			delete(route.VIPsV6, rule.Src.String())
			continue
		}

		err := netlink.RuleDel(&rule)
		if err != nil {
			log.FromContextOrGlobal(ctx).Error(err, "failed to RuleDel")
		}
	}

	for tableID, route := range networkConfig {
		err := setRules(route.VIPsV4, tableID, netlink.FAMILY_V4)
		if err != nil {
			log.FromContextOrGlobal(ctx).Error(err, "failed to setRules V4")
		}

		err = setRules(route.VIPsV6, tableID, netlink.FAMILY_V6)
		if err != nil {
			log.FromContextOrGlobal(ctx).Error(err, "failed to setRules V6")
		}

		err = setRoutes(route.GatewaysV4, tableID, true)
		if err != nil {
			log.FromContextOrGlobal(ctx).Error(err, "failed to setRoutes V4")
		}

		err = setRoutes(route.GatewaysV6, tableID, false)
		if err != nil {
			log.FromContextOrGlobal(ctx).Error(err, "failed to setRoutes V6")
		}
	}

	return nil
}

func deleteTable(tableID int) error {
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, &netlink.Route{Table: tableID}, netlink.RT_FILTER_TABLE)
	if err != nil {
		return fmt.Errorf("failed to RouteListFiltered: %w", err)
	}

	for _, route := range routes {
		err := netlink.RouteDel(&route)
		if err != nil {
			return fmt.Errorf("failed to RouteDel: %w", err)
		}
	}

	return nil
}

func setRules(vips map[string]*net.IPNet, tableID int, family int) error {
	for _, vip := range vips {
		rule := netlink.NewRule()
		rule.Src = vip
		rule.Family = family
		rule.Table = tableID

		err := netlink.RuleAdd(rule)
		if err != nil {
			return fmt.Errorf("failed to RuleAdd (rule: %v): %w", rule, err)
		}
	}

	return nil
}

func setRoutes(gateways map[string]net.IP, tableID int, isIPv4 bool) error {
	nexthops := []*netlink.NexthopInfo{}

	family := netlink.FAMILY_V4
	dst := net.IPv4(0, 0, 0, 0)
	if !isIPv4 {
		family = netlink.FAMILY_V6
		dst = net.ParseIP("::")
	}

	for _, gateway := range gateways {
		if isIPv4 && gateway.To4() != nil {
			nexthops = append(nexthops, &netlink.NexthopInfo{
				Gw: gateway,
			})
		} else if !isIPv4 && gateway.To4() == nil {
			nexthops = append(nexthops, &netlink.NexthopInfo{
				Gw: gateway,
			})
		}
	}

	if len(nexthops) <= 0 {
		// todo: delete previous routes
		return nil
	}

	route := &netlink.Route{
		Dst: &net.IPNet{
			IP:   dst,
			Mask: net.IPMask{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		MultiPath: nexthops,
		Family:    family,
		Table:     tableID,
	}

	err := netlink.RouteReplace(route)
	if err != nil {
		return fmt.Errorf("failed to RouteReplace (route: %v): %w", route, err)
	}

	return nil
}
