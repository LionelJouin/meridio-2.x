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

package statelessloadbalancer

import (
	"context"
	"fmt"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/log"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func (c *Controller) reconcileL34Routes(ctx context.Context, gateway *gatewayapiv1.Gateway) error {
	l34routeList := &v1alpha1.L34RouteList{}

	// err := c.List(ctx,
	// 	gatewayList,
	// 	client.MatchingFields{
	// 		"spec.ParentRefs[0].Name": c.GatewayClassName,
	// 	})
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

	err = c.ServiceManager.SetFlows(ctx, l34Routes)
	if err != nil {
		return fmt.Errorf("failed set flows while reconciling the L34Routes: %w", err)
	}

	return nil
}
