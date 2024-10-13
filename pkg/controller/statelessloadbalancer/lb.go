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

	"github.com/lioneljouin/meridio-experiment/pkg/nfqlb"
)

// NFQLBInstance is a wrapper of nfqlb.NFQueueLoadBalancer to match
// the LoadBalancerInstance interface.
type NFQLBInstance struct {
	*nfqlb.NFQueueLoadBalancer
}

// NewNFQLB is the constructor of NFQLBInstance.
func NewNFQLB(nfqLoadBalancer *nfqlb.NFQueueLoadBalancer) *NFQLBInstance {
	return &NFQLBInstance{
		nfqLoadBalancer,
	}
}

// AddService implements AddService of LoadBalancerInstance for nfqlb.
//
//nolint:ireturn
func (nfqlbi *NFQLBInstance) AddService(ctx context.Context, name string) (ServiceInstance, error) {
	service, err := nfqlbi.NFQueueLoadBalancer.AddService(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &nfqlbServiceInstance{
		serviceName: name,
		Service:     service,
	}, nil
}

type nfqlbServiceInstance struct {
	serviceName string
	*nfqlb.Service
}

func (nfqlbsi *nfqlbServiceInstance) GetName() string {
	return nfqlbsi.serviceName
}

func (nfqlbsi *nfqlbServiceInstance) AddFlow(ctx context.Context, flowToAdd Flow) error {
	err := nfqlbsi.Service.AddFlow(ctx, flowToAdd)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (nfqlbsi *nfqlbServiceInstance) DeleteFlow(ctx context.Context, flowToDelete Flow) error {
	err := nfqlbsi.Service.DeleteFlow(ctx, flowToDelete)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
