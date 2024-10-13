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

package gateway

import (
	"context"
	"fmt"
	"net"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	"k8s.io/utils/ptr"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// reconcileL34Routes reconciles all l34routes managed by the gateway.
func (c *Controller) reconcileL34Routes(ctx context.Context, gateway *gatewayapiv1.Gateway) error {
	l34routeList := &v1alpha1.L34RouteList{}

	err := c.List(ctx,
		l34routeList) // todo: filter namespace
	if err != nil {
		log.FromContextOrGlobal(ctx).Error(err, "failed listing the L34routes while reconciling the L34Routes")

		return nil
	}

	l34Routes := []*v1alpha1.L34Route{}

	for _, l34Route := range l34routeList.Items {
		if len(l34Route.Spec.ParentRefs) == 0 ||
			string(l34Route.Spec.ParentRefs[0].Name) != gateway.GetName() ||
			l34Route.GetNamespace() != gateway.GetNamespace() {
			continue
		}

		l34r := l34Route

		l34Routes = append(l34Routes, &l34r)
	}

	// update gateway status IPs with service IPs
	gateway.Status.Addresses = []gatewayapiv1.GatewayStatusAddress{}
	vips := map[string]struct{}{}

	for _, l34Route := range l34Routes {
		for _, destinationCIDR := range l34Route.Spec.DestinationCIDRs {
			_, ipNet, err := net.ParseCIDR(destinationCIDR)
			if err != nil {
				continue
			}

			_, exists := vips[ipNet.IP.String()]
			if exists {
				continue
			}

			gateway.Status.Addresses = append(gateway.Status.Addresses, gatewayapiv1.GatewayStatusAddress{
				Type:  ptr.To(gatewayapiv1.IPAddressType),
				Value: ipNet.IP.String(),
			})

			vips[ipNet.IP.String()] = struct{}{}
		}
	}

	err = c.Status().Update(ctx, gateway)
	if err != nil {
		return fmt.Errorf("failed to update gateway status: %w", err)
	}

	return nil
}
