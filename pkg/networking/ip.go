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

package networking

import (
	"net"
)

// BroadcastFromIPNet returns the broadcast address from an IPNet.
func BroadcastFromIPNet(ipNet *net.IPNet) net.IP {
	broadcast := make([]byte, len(ipNet.IP))
	copy(broadcast, ipNet.IP)

	for i := 0; i < len(ipNet.IP); i++ {
		broadcast[i] = ipNet.IP[i] | ^ipNet.Mask[i]
	}

	return broadcast
}

// NextIP returns the next ip.
func NextIP(ip net.IP) net.IP {
	next := make([]byte, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}

	return next
}
