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

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
)

func isIPv4CIDR(cidr string) bool {
	ipAddr, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return ipAddr.To4() != nil
}

func isIPv6CIDR(cidr string) bool {
	ipAddr, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}

	return ipAddr.To4() == nil
}

func isIPv4(ip string) bool {
	ipAddr := net.ParseIP(ip)

	if ipAddr == nil {
		return false
	}

	return ipAddr.To4() != nil
}

func isIPv6(ip string) bool {
	ipAddr := net.ParseIP(ip)

	if ipAddr == nil {
		return false
	}

	return ipAddr.To4() == nil
}

// setPolicyRoutes finds all policy routes with the table ID used for Bird.
// It deleted all policy routes for no longer existing vips and creates the
// ones for the newly added.
func setPolicyRoutes(vips []string) error {
	rules, err := netlink.RuleListFiltered(netlink.FAMILY_ALL, &netlink.Rule{
		Table: defaultKernelTableID,
	}, netlink.RT_FILTER_TABLE)
	if err != nil {
		return fmt.Errorf("failed to list rules: %w", err)
	}

	vipMap := map[string]*net.IPNet{}

	for _, vip := range vips {
		_, vipIPNet, err := net.ParseCIDR(vip)
		if err != nil {
			continue
		}

		vipMap[vipIPNet.String()] = vipIPNet
	}

	var errFinal error

	for _, rule := range rules {
		currentRule := rule

		_, exists := vipMap[rule.Src.String()]
		if !exists {
			err := netlink.RuleDel(&currentRule)
			if err != nil {
				errFinal = fmt.Errorf("failed to RuleDel ; %w; %w", err, errFinal)
			}

			continue
		}

		delete(vipMap, rule.Src.String())
	}

	for _, vipIPNet := range vipMap {
		rule := netlink.NewRule()
		rule.Priority = 100
		rule.Table = defaultKernelTableID
		rule.Src = vipIPNet

		err := netlink.RuleAdd(rule)
		if err != nil {
			errFinal = fmt.Errorf("failed to RuleAdd ; %w; %w", err, errFinal)
		}
	}

	return errFinal
}
