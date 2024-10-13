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
	"errors"
	"fmt"
	"sync"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
	"github.com/lioneljouin/meridio-experiment/pkg/endpointslice/endpoint"
	v1 "k8s.io/api/core/v1"
	v1discovery "k8s.io/api/discovery/v1"
)

var errServiceNotExisting = errors.New("the service does not exist")

// LoadBalancerInstance defines an interface to add/delete load-balancer services
// within a load balancer instance (e.g. nfqlb).
type LoadBalancerInstance interface {
	// AddService adds a load-balancer service.
	AddService(ctx context.Context, name string) (ServiceInstance, error)
	// DeleteService deletes a load-balancer service and all related configuration (targets and flows).
	DeleteService(ctx context.Context, name string) error
}

// ServiceInstance represents a service instantiated by the load balancer instance.
type ServiceInstance interface {
	// Name of the Service
	GetName() string
	// AddFlow adds/updates a Flow selecting the associated load-balancer service.
	AddFlow(ctx context.Context, flowToAdd Flow) error
	// DeleteFlow adds a Flow selecting the associated load-balancer service.
	DeleteFlow(ctx context.Context, flowToDelete Flow) error
	// AddTarget adds a target identifier to the load-balancer service
	// and configures the policy route associated.
	AddTarget(ctx context.Context, ips []string, identifier int) error
	// DeleteTarget deletes a target identifier to the load-balancer service
	// and deletes the policy route associated.
	DeleteTarget(ctx context.Context, ips []string, identifier int) error
}

// Flow is the interface that wraps the basic Flow method.
type Flow interface {
	// Name of the flow
	GetName() string
	// Source CIDRs allowed in the flow
	// e.g.: ["124.0.0.0/24", "2001::/32"
	GetSourceCIDRs() []string
	// Destination CIDRs allowed in the flow
	// e.g.: ["124.0.0.0/24", "2001::/32"
	GetDestinationCIDRs() []string
	// Source port ranges allowed in the flow
	// e.g.: ["35000-35500", "40000"]
	GetSourcePortRanges() []string
	// Destination port ranges allowed in the flow
	// e.g.: ["35000-35500", "40000"]
	GetDestinationPortRanges() []string
	// Protocols allowed
	// e.g.: ["tcp", "udp"]
	GetProtocols() []string
	// Priority of the flow
	GetPriority() int32
	// Bytes in L4 header
	GetByteMatches() []string
}

// Manager is an helper structure to control a load balancer instance.
type Manager struct {
	LoadBalancer LoadBalancerInstance
	services     map[string]ServiceInstance         // key: service name
	flows        map[string]*flowImpl               // key: <l34Route-name>.<service.name>
	endpoints    map[string][]*v1discovery.Endpoint // key: service name
	mu           sync.Mutex
}

// NewManager is the constructor of Manager.
func NewManager(loadBalancer LoadBalancerInstance) *Manager {
	mngr := &Manager{
		LoadBalancer: loadBalancer,
		services:     map[string]ServiceInstance{},
		flows:        map[string]*flowImpl{},
		endpoints:    map[string][]*v1discovery.Endpoint{},
	}

	return mngr
}

// SetService adds the service to the load balancer instance if not already existing.
func (m *Manager) SetServices(
	ctx context.Context,
	services []*v1.Service,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errFinal error

	newServices := map[string]*v1.Service{} // key: <l34Route-name>.<service.name>

	for _, service := range services {
		newServices[service.GetName()] = service
	}

	// To delete
	for _, service := range m.services {
		_, exists := newServices[service.GetName()]
		if exists {
			continue
		}

		err := m.LoadBalancer.DeleteService(ctx, service.GetName())
		if err != nil {
			errFinal = fmt.Errorf("failed to DeleteService ; %w; %w", err, errFinal)

			continue
		}

		delete(m.services, service.GetName())
		delete(m.endpoints, service.GetName())
	}

	// To add
	for _, service := range newServices {
		_, exists := m.services[service.GetName()]
		if exists {
			continue
		}

		lbService, err := m.LoadBalancer.AddService(ctx, service.GetName())
		if err != nil {
			return fmt.Errorf("failed to AddService: %w", err)
		}

		m.services[service.GetName()] = lbService
		m.endpoints[service.GetName()] = []*v1discovery.Endpoint{}
	}

	// cleanup flows
	for name, flow := range m.flows {
		_, exists := m.services[flow.service.GetName()]
		if exists {
			continue
		}

		delete(m.flows, name)
	}

	return errFinal
}

// SetFlows adds the non existing l34Routes, updates the existing ones and removes the ones that are not passed
// as parameter.
func (m *Manager) SetFlows(ctx context.Context,
	l34Routes []*v1alpha1.L34Route,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errFinal error

	newFlows := map[string]*flowImpl{} // key: <l34Route-name>.<service.name>

	for _, l34Route := range l34Routes {
		flowI := &flowImpl{
			L34Route: l34Route,
			service:  m.getServiceForL34Route(l34Route),
		}

		if flowI.service == nil {
			continue
		}

		newFlows[flowI.GetName()] = flowI
	}

	// To delete
	for _, flow := range m.flows {
		_, exists := newFlows[flow.GetName()]
		if exists {
			continue
		}

		err := flow.service.DeleteFlow(ctx, flow)
		if err != nil {
			errFinal = fmt.Errorf("failed to DeleteFlow ; %w; %w", err, errFinal)
		}

		delete(m.flows, flow.GetName())
	}

	// To add/update
	for _, flow := range newFlows {
		m.flows[flow.GetName()] = flow

		err := flow.service.AddFlow(ctx, flow)
		if err != nil {
			errFinal = fmt.Errorf("failed to AddFlow ; %w; %w", err, errFinal)
		}
	}

	return errFinal
}

func (m *Manager) getServiceForL34Route(l34Route *v1alpha1.L34Route) ServiceInstance {
	if len(l34Route.Spec.BackendRefs) == 0 {
		return nil
	}

	serviceInstance, exists := m.services[string(l34Route.Spec.BackendRefs[0].Name)]
	if !exists {
		return nil
	}

	return serviceInstance
}

// SetEndpoints adds the non existing endpoint, updates the existing ones and removes the ones that are not passed
// as parameter.
func (m *Manager) SetEndpoints(
	ctx context.Context,
	service *v1.Service,
	endpoints []v1discovery.Endpoint,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	serviceInstance, exists := m.services[service.GetName()]
	if !exists {
		return errServiceNotExisting
	}

	previousEndpoints, exists := m.endpoints[service.GetName()]
	if !exists {
		return errServiceNotExisting
	}

	endpointsMap := map[string]*v1discovery.Endpoint{}

	for _, endpnt := range endpoints {
		currentEndpoint := endpnt

		endpointsMap[string(currentEndpoint.TargetRef.UID)] = &currentEndpoint
	}

	var errFinal error

	finalEndpoints := []*v1discovery.Endpoint{}

	endpts, err := removeEndpoints(ctx, serviceInstance, previousEndpoints, endpointsMap)
	if err != nil {
		errFinal = fmt.Errorf("%w; %w", err, errFinal)
	}

	finalEndpoints = append(finalEndpoints, endpts...)

	endpts, err = addEndpoints(ctx, serviceInstance, endpointsMap)
	if err != nil {
		errFinal = fmt.Errorf("%w; %w", err, errFinal)
	}

	finalEndpoints = append(finalEndpoints, endpts...)

	m.endpoints[service.GetName()] = finalEndpoints

	return errFinal
}

func addEndpoints(
	ctx context.Context,
	serviceInstance ServiceInstance,
	endpointsMap map[string]*v1discovery.Endpoint,
) ([]*v1discovery.Endpoint, error) {
	var errFinal error

	finalEndpoints := []*v1discovery.Endpoint{}

	for _, endpnt := range endpointsMap {
		id := endpoint.GetIdentifier(*endpnt)
		if id == nil || endpnt.Conditions.Ready == nil || !*endpnt.Conditions.Ready {
			continue
		}

		err := serviceInstance.AddTarget(ctx, endpnt.Addresses, *id)
		if err != nil {
			errFinal = fmt.Errorf("failed to AddTarget ; %w; %w", err, errFinal)
		}

		finalEndpoints = append(finalEndpoints, endpnt)
	}

	return finalEndpoints, errFinal
}

func removeEndpoints(
	ctx context.Context,
	serviceInstance ServiceInstance,
	previousEndpoints []*v1discovery.Endpoint,
	endpointsMap map[string]*v1discovery.Endpoint,
) ([]*v1discovery.Endpoint, error) {
	var errFinal error

	finalEndpoints := []*v1discovery.Endpoint{}

	for _, endpnt := range previousEndpoints {
		id := endpoint.GetIdentifier(*endpnt)
		if id == nil {
			continue
		}

		newEndpoint, exists := endpointsMap[string(endpnt.TargetRef.UID)]
		if !exists { // no longer exist
			err := serviceInstance.DeleteTarget(ctx, endpnt.Addresses, *id)
			if err != nil {
				errFinal = fmt.Errorf("failed to DeleteTarget ; %w; %w", err, errFinal)
			}

			continue
		}

		// verify if the endpoint has changed
		newID := endpoint.GetIdentifier(*newEndpoint)

		if newID != nil &&
			*newID == *id &&
			sameStringSlice(endpnt.Addresses, newEndpoint.Addresses) &&
			newEndpoint.Conditions.Ready != nil &&
			*newEndpoint.Conditions.Ready { // has not changed and is still ready
			delete(endpointsMap, string(endpnt.TargetRef.UID))
			finalEndpoints = append(finalEndpoints, endpnt)

			continue
		}

		err := serviceInstance.DeleteTarget(ctx, endpnt.Addresses, *id)
		if err != nil {
			errFinal = fmt.Errorf("failed to DeleteTarget ; %w; %w", err, errFinal)
		}
	}

	return finalEndpoints, errFinal
}

// https://stackoverflow.com/questions/36000487/check-for-equality-on-slices-without-order
func sameStringSlice(sliceA []string, sliceB []string) bool {
	if len(sliceA) != len(sliceB) {
		return false
	}

	diff := map[string]int{}

	for _, sliceAElement := range sliceA {
		diff[sliceAElement]++
	}

	for _, sliceBElement := range sliceB {
		if _, ok := diff[sliceBElement]; !ok {
			return false
		}

		diff[sliceBElement]--

		if diff[sliceBElement] == 0 {
			delete(diff, sliceBElement)
		}
	}

	return len(diff) == 0
}

type flowImpl struct {
	*v1alpha1.L34Route
	service ServiceInstance
}

func (f *flowImpl) GetName() string {
	return fmt.Sprintf("%s.%s", f.L34Route.GetName(), f.service.GetName())
}

func (f *flowImpl) GetSourceCIDRs() []string {
	return f.Spec.SourceCIDRs
}

func (f *flowImpl) GetDestinationCIDRs() []string {
	return f.Spec.DestinationCIDRs
}

func (f *flowImpl) GetSourcePortRanges() []string {
	return f.Spec.SourcePorts
}

func (f *flowImpl) GetDestinationPortRanges() []string {
	return f.Spec.DestinationPorts
}

func (f *flowImpl) GetProtocols() []string {
	protocols := []string{}
	for _, protocol := range f.Spec.Protocols {
		protocols = append(protocols, string(protocol))
	}

	return protocols
}

func (f *flowImpl) GetPriority() int32 {
	return f.Spec.Priority
}

func (f *flowImpl) GetByteMatches() []string {
	return f.Spec.ByteMatches
}
