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

package nfqlb_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/containernetworking/plugins/pkg/testutils"
	"github.com/google/nftables"
	"github.com/lioneljouin/meridio-experiment/pkg/nfqlb"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
	"go.uber.org/goleak"
)

// TODO: https://github.com/google/nftables/blob/main/userdata/userdata_cli_interop_test.go#L29

const (
	tableName      = "table-nfqlb"
	chainName      = "nfqlb"
	localChainName = "nfqlb-local"
	ipv4VIPSetName = "ipv4-vips"
	ipv6VIPSetName = "ipv6-vips"

	serviceNameA = "test-a"
	serviceNameB = "test-b"
	flowNameA    = "flow-a"
	flowNameB    = "flow-b"
)

func TestNFQLBStartStop(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	testNS, err := testutils.NewNS()
	assert.Nil(t, err)

	defer func() {
		_ = testNS.Close()
		_ = testutils.UnmountNS(testNS)
	}()

	path, err := os.Getwd()
	assert.Nil(t, err)

	err = testNS.Do(func(ns.NetNS) error {
		nfQueueLoadBalancer, err := nfqlb.New(
			nfqlb.WithNFQLBPath(filepath.Join(path, "testing", "nfqlb")),
			nfqlb.WithQueue("0:1"),
			nfqlb.WithQLength(512),
			nfqlb.WithFanout(true),
		)
		assert.Nil(t, err)
		assert.NotNil(t, nfQueueLoadBalancer)

		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(context.Background())

		wg.Add(1)

		go func() {
			defer wg.Done()

			// execute again in the network namespace is required because of the go routine.
			_ = testNS.Do(func(ns.NetNS) error {
				_ = nfQueueLoadBalancer.Start(ctx)

				return nil
			})
		}()

		conn := &nftables.Conn{}
		tables, err := conn.ListTables()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tables))
		assert.Equal(t, tableName, tables[0].Name)

		set, err := conn.GetSetByName(tables[0], ipv4VIPSetName)
		assert.Nil(t, err)
		assert.NotNil(t, set)

		set, err = conn.GetSetByName(tables[0], ipv6VIPSetName)
		assert.Nil(t, err)
		assert.NotNil(t, set)

		chains, err := conn.ListChains()
		assert.Nil(t, err)
		assert.Equal(t, 2, len(chains))

		cancel()

		wg.Wait()

		tables, err = conn.ListTables()
		assert.Nil(t, err)
		assert.Equal(t, 0, len(tables))

		return nil
	})
	assert.Nil(t, err)
}

func TestAddDeleteService(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	testNS, err := testutils.NewNS()
	assert.Nil(t, err)

	defer func() {
		_ = testNS.Close()
		_ = testutils.UnmountNS(testNS)
	}()

	path, err := os.Getwd()
	assert.Nil(t, err)

	err = testNS.Do(func(ns.NetNS) error {
		nfQueueLoadBalancer, err := nfqlb.New(
			nfqlb.WithNFQLBPath(filepath.Join(path, "testing", "nfqlb")),
		)
		assert.Nil(t, err)
		assert.NotNil(t, nfQueueLoadBalancer)

		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(context.Background())

		wg.Add(1)

		go func() {
			defer wg.Done()

			// execute again in the network namespace is required because of the go routine.
			_ = testNS.Do(func(ns.NetNS) error {
				_ = nfQueueLoadBalancer.Start(ctx)

				return nil
			})
		}()

		// Add Service A
		serviceA, err := nfQueueLoadBalancer.AddService(
			ctx,
			serviceNameA,
			nfqlb.WithMaxTargets(1),
		)
		assert.Nil(t, err)
		assert.NotNil(t, serviceA)

		cmd := exec.CommandContext(
			context.Background(),
			filepath.Join(path, "testing", "nfqlb"),
			"show",
			fmt.Sprintf("--shm=%s", serviceNameA),
		)
		err = cmd.Run()
		assert.Nil(t, err)

		// Add Service B
		serviceB, err := nfQueueLoadBalancer.AddService(
			ctx,
			serviceNameB,
			nfqlb.WithMaxTargets(1),
		)
		assert.Nil(t, err)
		assert.NotNil(t, serviceB)

		cmd = exec.CommandContext(
			context.Background(),
			filepath.Join(path, "testing", "nfqlb"),
			"show",
			fmt.Sprintf("--shm=%s", serviceNameB),
		)
		err = cmd.Run()
		assert.Nil(t, err)

		// Delete Service A
		err = nfQueueLoadBalancer.DeleteService(
			ctx,
			serviceNameA,
		)
		assert.Nil(t, err)

		cmd = exec.CommandContext(
			context.Background(),
			filepath.Join(path, "testing", "nfqlb"),
			"show",
			fmt.Sprintf("--shm=%s", serviceNameA),
		)
		err = cmd.Run()
		assert.NotNil(t, err)

		// Delete Service B
		err = nfQueueLoadBalancer.DeleteService(
			ctx,
			serviceNameB,
		)
		assert.Nil(t, err)

		cmd = exec.CommandContext(
			context.Background(),
			filepath.Join(path, "testing", "nfqlb"),
			"show",
			fmt.Sprintf("--shm=%s", serviceNameB),
		)
		err = cmd.Run()
		assert.NotNil(t, err)

		cancel()

		wg.Wait()

		return nil
	})
	assert.Nil(t, err)
}

func TestAddDeleteFlow(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	testNS, err := testutils.NewNS()
	assert.Nil(t, err)

	defer func() {
		_ = testNS.Close()
		_ = testutils.UnmountNS(testNS)
	}()

	path, err := os.Getwd()
	assert.Nil(t, err)

	err = testNS.Do(func(ns.NetNS) error {
		nfQueueLoadBalancer, err := nfqlb.New(
			nfqlb.WithNFQLBPath(filepath.Join(path, "testing", "nfqlb")),
		)
		assert.Nil(t, err)
		assert.NotNil(t, nfQueueLoadBalancer)

		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(context.Background())

		wg.Add(1)

		go func() {
			defer wg.Done()

			// execute again in the network namespace is required because of the go routine.
			_ = testNS.Do(func(ns.NetNS) error {
				_ = nfQueueLoadBalancer.Start(ctx)

				return nil
			})
		}()

		service, err := nfQueueLoadBalancer.AddService(
			ctx,
			serviceNameA,
			nfqlb.WithMaxTargets(1),
		)
		assert.Nil(t, err)
		assert.NotNil(t, service)

		flowA := &flowMock{
			name:                  flowNameA,
			sourceCIDRs:           []string{},
			destinationCIDRs:      []string{"20.0.0.1/32", "2000::1/128"},
			sourcePortRanges:      []string{},
			destinationPortRanges: []string{},
			protocols:             []string{},
			priority:              1,
			byteMatches:           []string{},
		}
		flowB := &flowMock{
			name:                  flowNameB,
			sourceCIDRs:           []string{},
			destinationCIDRs:      []string{"30.0.0.1/32", "3000::1/128", "20.0.0.1/32", "2000::1/128"},
			sourcePortRanges:      []string{},
			destinationPortRanges: []string{},
			protocols:             []string{},
			priority:              1,
			byteMatches:           []string{},
		}

		conn := &nftables.Conn{}
		tables, err := conn.ListTables()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tables))

		setV4, err := conn.GetSetByName(tables[0], ipv4VIPSetName)
		assert.Nil(t, err)
		assert.NotNil(t, setV4)

		setV6, err := conn.GetSetByName(tables[0], ipv6VIPSetName)
		assert.Nil(t, err)
		assert.NotNil(t, setV6)

		// Add flow A
		err = service.AddFlow(ctx, flowA)
		assert.Nil(t, err)

		elements, err := conn.GetSetElements(setV4)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(elements))
		elements, err = conn.GetSetElements(setV6)
		assert.Nil(t, err)
		assert.Equal(t, 2, len(elements))

		// Add flow B
		err = service.AddFlow(ctx, flowB)
		assert.Nil(t, err)

		elements, err = conn.GetSetElements(setV4)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(elements))
		elements, err = conn.GetSetElements(setV6)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(elements))

		// Delete flow A
		err = service.DeleteFlow(ctx, flowA)
		assert.Nil(t, err)

		elements, err = conn.GetSetElements(setV4)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(elements))
		elements, err = conn.GetSetElements(setV6)
		assert.Nil(t, err)
		assert.Equal(t, 4, len(elements))

		// Delete flow B
		err = service.DeleteFlow(ctx, flowB)
		assert.Nil(t, err)

		elements, err = conn.GetSetElements(setV4)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(elements))
		elements, err = conn.GetSetElements(setV6)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(elements))

		cancel()

		wg.Wait()

		return nil
	})
	assert.Nil(t, err)
}

func TestAddDeleteTarget(t *testing.T) {
	t.Cleanup(func() { goleak.VerifyNone(t) })

	testNS, err := testutils.NewNS()
	assert.Nil(t, err)

	defer func() {
		_ = testNS.Close()
		_ = testutils.UnmountNS(testNS)
	}()

	path, err := os.Getwd()
	assert.Nil(t, err)

	err = testNS.Do(func(ns.NetNS) error {
		cmd := exec.CommandContext(context.Background(), "ip", "link", "add", "dummy1", "type", "dummy")
		err := cmd.Run()
		assert.Nil(t, err)
		cmd = exec.CommandContext(context.Background(), "ip", "addr", "add", "169.255.0.0/24", "dev", "dummy1")
		err = cmd.Run()
		assert.Nil(t, err)
		cmd = exec.CommandContext(context.Background(), "ip", "addr", "add", "fd00::/64", "dev", "dummy1")
		err = cmd.Run()
		assert.Nil(t, err)
		cmd = exec.CommandContext(context.Background(), "ip", "link", "set", "dummy1", "up")
		err = cmd.Run()
		assert.Nil(t, err)

		nfQueueLoadBalancer, err := nfqlb.New(
			nfqlb.WithNFQLBPath(filepath.Join(path, "testing", "nfqlb")),
			nfqlb.WithStartingOffset(5000),
		)
		assert.Nil(t, err)
		assert.NotNil(t, nfQueueLoadBalancer)

		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(context.Background())

		wg.Add(1)

		go func() {
			defer wg.Done()

			// execute again in the network namespace is required because of the go routine.
			_ = testNS.Do(func(ns.NetNS) error {
				_ = nfQueueLoadBalancer.Start(ctx)

				return nil
			})
		}()

		service, err := nfQueueLoadBalancer.AddService(
			ctx,
			serviceNameA,
			nfqlb.WithMaxTargets(2),
		)
		assert.Nil(t, err)
		assert.NotNil(t, service)

		// Add target 1
		err = service.AddTarget(ctx, []string{"169.255.0.1", "fd00::1"}, 0)
		assert.Nil(t, err)
		routes, err := netlink.RouteListFiltered(
			netlink.FAMILY_V4,
			&netlink.Route{
				Gw:    net.ParseIP("169.255.0.1"),
				Table: 5000,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(routes))
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V6,
			&netlink.Route{
				Gw:    net.ParseIP("fd00::1"),
				Table: 5000,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(routes))

		// Add target 2
		err = service.AddTarget(ctx, []string{"169.255.0.2", "fd00::2"}, 1)
		assert.Nil(t, err)
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V4,
			&netlink.Route{
				Gw:    net.ParseIP("169.255.0.2"),
				Table: 5001,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(routes))
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V6,
			&netlink.Route{
				Gw:    net.ParseIP("fd00::2"),
				Table: 5001,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(routes))

		// Add target 3
		err = service.AddTarget(ctx, []string{"169.255.0.3", "fd00::3"}, 2)
		assert.NotNil(t, err)

		// Delete target 1
		err = service.DeleteTarget(ctx, []string{"169.255.0.1", "fd00::1"}, 0)
		assert.Nil(t, err)
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V4,
			&netlink.Route{
				Gw:    net.ParseIP("169.255.0.1"),
				Table: 5000,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(routes))
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V6,
			&netlink.Route{
				Gw:    net.ParseIP("fd00::1"),
				Table: 5000,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(routes))

		// Delete target 2
		err = service.DeleteTarget(ctx, []string{"169.255.0.2", "fd00::2"}, 1)
		assert.Nil(t, err)
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V4,
			&netlink.Route{
				Gw:    net.ParseIP("169.255.0.2"),
				Table: 5001,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(routes))
		routes, err = netlink.RouteListFiltered(
			netlink.FAMILY_V6,
			&netlink.Route{
				Gw:    net.ParseIP("fd00::2"),
				Table: 5001,
			},
			netlink.RT_FILTER_GW|netlink.RT_FILTER_TABLE,
		)
		assert.Nil(t, err)
		assert.Equal(t, 0, len(routes))

		cancel()

		wg.Wait()

		return nil
	})
	assert.Nil(t, err)
}

type flowMock struct {
	name                  string
	sourceCIDRs           []string
	destinationCIDRs      []string
	sourcePortRanges      []string
	destinationPortRanges []string
	protocols             []string
	priority              int32
	byteMatches           []string
}

func (fm *flowMock) GetName() string {
	return fm.name
}

func (fm *flowMock) GetSourceCIDRs() []string {
	return fm.sourceCIDRs
}

func (fm *flowMock) GetDestinationCIDRs() []string {
	return fm.destinationCIDRs
}

func (fm *flowMock) GetSourcePortRanges() []string {
	return fm.sourcePortRanges
}

func (fm *flowMock) GetDestinationPortRanges() []string {
	return fm.destinationPortRanges
}

func (fm *flowMock) GetProtocols() []string {
	return fm.protocols
}

func (fm *flowMock) GetPriority() int32 {
	return fm.priority
}

func (fm *flowMock) GetByteMatches() []string {
	return fm.byteMatches
}
