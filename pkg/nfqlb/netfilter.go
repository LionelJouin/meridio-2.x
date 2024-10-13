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

package nfqlb

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

// netfilterAdaptor configures nftables to direct IP packets whos destination
// address matches netfilter IP Sets ipv4DestinationSet, ipv6DestinationSet to
// the configured target netfilter queue(s).
//
// Supports udpate of the IP Sets ipv4DestinationSet, ipv6DestinationSet.

/* Example config:
table inet table-nfqlb {
	set ipv4-vips {
		type ipv4_addr
		flags interval
		elements = { 20.0.0.1, 40.0.0.0/24 }
	}

	set ipv6-vips {
		type ipv6_addr
		flags interval
		elements = { 2000::1 }
	}

	chain nfqlb {
		type filter hook prerouting priority filter; policy accept;
		ip daddr @ipv4-vips counter packets 15364 bytes 3948540 queue num 0-3
		ip6 daddr @ipv6-vips counter packets 14800 bytes 4443820 queue num 0-3
	}

	chain nfqlb-local {
		type filter hook output priority filter; policy accept;
		meta l4proto icmp ip daddr @ipv4-vips counter packets 1 bytes 576 queue num 0-3
		meta l4proto ipv6-icmp ip6 daddr @ipv6-vips counter packets 0 bytes 0 queue num 0-3
	}
}
*/

func newNetfilterQueue(nftqueueNum uint16, nftqueueTotal uint16, queueFlag expr.QueueFlag) (*netfilterQueue, error) {
	nfQueue := &netfilterQueue{
		nftqueueNum:   nftqueueNum,
		nftqueueTotal: nftqueueTotal,
		nftqueueFlag:  queueFlag,
	}

	if err := nfQueue.configure(); err != nil {
		nfQueue.logger.Error(err, "configure")

		return nil, err
	}

	return nfQueue, nil
}

type netfilterQueue struct {
	table              *nftables.Table
	chain              *nftables.Chain
	localchain         *nftables.Chain
	ipv4Rule           *nftables.Rule
	ipv6Rule           *nftables.Rule
	ipv4DestinationSet *nftables.Set
	ipv6DestinationSet *nftables.Set
	nftqueueFlag       expr.QueueFlag
	nftqueueNum        uint16 // start of nqueue range
	nftqueueTotal      uint16 // number of nfqueues in use
	logger             logr.Logger
}

// delete removes nftables chains rules.
func (nfq *netfilterQueue) delete() error {
	conn := &nftables.Conn{}

	conn.FlushTable(nfq.table)
	conn.DelTable(nfq.table)

	err := conn.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush (delete): %w", err)
	}

	return nil
}

func (nfq *netfilterQueue) configure() error {
	if nfq.table == nil {
		// create nf table
		if err := nfq.configureTable(); err != nil {
			return err
		}
	}

	if err := nfq.configureSets(); err != nil {
		return err
	}

	if err := nfq.configureChainAndRules(); err != nil {
		return err
	}

	return nfq.configureLocalChainAndRules()
}

// configureTable creates netfilter table if not yet present.
func (nfq *netfilterQueue) configureTable() error {
	conn := &nftables.Conn{}

	table := conn.AddTable(&nftables.Table{
		Name:   tableName,
		Family: nftables.TableFamilyINet,
	})

	err := conn.Flush()
	if err != nil {
		return fmt.Errorf("netfilterQueue: nftable: %w", err)
	}

	nfq.table = table

	return nil
}

// configureSets creates nftables Sets for both IPv4 and IPv6 destination addresses.
func (nfq *netfilterQueue) configureSets() error {
	ipv4Set, err := createSets(nfq.table, ipv4)
	if err != nil {
		return fmt.Errorf("create ipv4 set %w", err)
	}

	ipv6Set, err := createSets(nfq.table, ipv6)
	if err != nil {
		return fmt.Errorf("create ipv6 set %w", err)
	}

	nfq.ipv4DestinationSet = ipv4Set
	nfq.ipv6DestinationSet = ipv6Set

	return nil
}

// configureChainAndRules adds nftables rules to direct incoming packets with matching dst address to targetNFQueue.
//
//nolint:gomnd,dupl,funlen,nolintlint
func (nfq *netfilterQueue) configureChainAndRules() error {
	conn := &nftables.Conn{}

	nfq.chain = conn.AddChain(&nftables.Chain{
		Name:     chainName,
		Table:    nfq.table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookPrerouting,
		Priority: nftables.ChainPriorityFilter,
	})

	if rules, _ := conn.GetRules(nfq.table, nfq.chain); len(rules) != 0 {
		conn.FlushChain(nfq.chain)
	}

	// nft add rule inet table-nfqlb nfqlb ip daddr @ipv4Vips counter queue num 0-3 fanout
	ipv4Rule := &nftables.Rule{
		Table: nfq.table,
		Chain: nfq.chain,
		Exprs: []expr.Any{
			// [ meta load nfproto => reg 1 ]
			&expr.Meta{
				Key:      expr.MetaKeyNFPROTO,
				Register: 1,
			},
			// [ cmp eq reg 1 0x00000002 ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.AF_INET},
			},
			// [ payload load 4b @ network header + 16 => reg 1 ]
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       16,
				Len:          4,
			},
			// [ lookup reg 1 set vips ]
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        nfq.ipv4DestinationSet.Name,
				SetID:          nfq.ipv4DestinationSet.ID,
			},
			// [ counter pkts 0 bytes 0 ]
			&expr.Counter{
				Bytes:   0,
				Packets: 0,
			},
			// [ queue num 1 ]
			&expr.Queue{
				Num:   nfq.nftqueueNum,
				Total: nfq.nftqueueTotal,
				Flag:  nfq.nftqueueFlag,
			},
		},
	}
	nfq.ipv4Rule = conn.AddRule(ipv4Rule)

	// nft add rule inet table-nfqlb nfqlb ip6 daddr @ipv6Vips counter queue num 0-3 fanout
	ipv6Rule := &nftables.Rule{
		Table: nfq.table,
		Chain: nfq.chain,
		Exprs: []expr.Any{
			// [ meta load nfproto => reg 1 ]
			&expr.Meta{
				Key:      expr.MetaKeyNFPROTO,
				Register: 1,
			},
			// [ cmp eq reg 1 0x0000000a ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.AF_INET6},
			},
			// [ payload load 16b @ network header + 24 => reg 1 ]
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       24,
				Len:          16,
			},
			// [ lookup reg 1 set flow-a-daddrs-v6 ]
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        nfq.ipv6DestinationSet.Name,
				SetID:          nfq.ipv6DestinationSet.ID,
			},
			// [ counter pkts 0 bytes 0 ]
			&expr.Counter{
				Bytes:   0,
				Packets: 0,
			},
			// [ queue num 1 ]
			&expr.Queue{
				Num:   nfq.nftqueueNum,
				Total: nfq.nftqueueTotal,
				Flag:  nfq.nftqueueFlag,
			},
		},
	}
	nfq.ipv6Rule = conn.AddRule(ipv6Rule)

	err := conn.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush (configureChainAndRules): %w", err)
	}

	return nil
}

// configureLocalChainAndRules adds nftables rules to direct locally generated ICMP unreachable reply packets
// with matching dst address to targetNFQueue (e.g. in case next-hop IP had lower PMTU)
// TODO: consider adding filter to only allow the unreachable and fragmentation related packets to match.
//
//nolint:gomnd,dupl,funlen
func (nfq *netfilterQueue) configureLocalChainAndRules() error {
	conn := &nftables.Conn{}

	nfq.localchain = conn.AddChain(&nftables.Chain{
		Name:     localChainName,
		Table:    nfq.table,
		Type:     nftables.ChainTypeFilter,
		Hooknum:  nftables.ChainHookOutput,
		Priority: nftables.ChainPriorityFilter,
	})

	if rules, _ := conn.GetRules(nfq.table, nfq.localchain); len(rules) != 0 {
		conn.FlushChain(nfq.localchain)
	}

	// nft add rule inet table-nfqlb nfqlb-local ip meta l4proto icmp daddr @ipv4Vips counter queue num 0-3 fanout.
	ipv4Rule := &nftables.Rule{
		Table: nfq.table,
		Chain: nfq.localchain,
		Exprs: []expr.Any{
			// [ meta load nfproto => reg 1 ]
			&expr.Meta{
				Key:      expr.MetaKeyNFPROTO,
				Register: 1,
			},
			// [ cmp eq reg 1 0x00000002 ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.AF_INET},
			},
			// [ meta load l4proto => reg 1 ]
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: 1,
			},
			// [ cmp eq reg 1 0x00000001 ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_ICMP},
			},
			// [ payload load 4b @ network header + 16 => reg 1 ]
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       16,
				Len:          4,
			},
			// [ lookup reg 1 set vips ]
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        nfq.ipv4DestinationSet.Name,
				SetID:          nfq.ipv4DestinationSet.ID,
			},
			// // [ payload load 2b @ transport header + 2 => reg 1 ]
			// &expr.Payload{
			// 	DestRegister: 1,
			// 	Base:         expr.PayloadBaseTransportHeader,
			// 	Offset:       0,
			// 	Len:          1,
			// },
			// // [ cmp eq reg 1 0x00000003 ]
			// &expr.Cmp{
			// 	Op:       expr.CmpOpEq,
			// 	Register: 1,
			// 	Data:     []byte{0x3},
			// },
			// [ counter pkts 0 bytes 0 ]
			&expr.Counter{
				Bytes:   0,
				Packets: 0,
			},
			// [ queue num 1 ]
			&expr.Queue{
				Num:   nfq.nftqueueNum,
				Total: nfq.nftqueueTotal,
				Flag:  nfq.nftqueueFlag,
			},
		},
	}
	nfq.ipv4Rule = conn.AddRule(ipv4Rule)

	// nft add rule inet table-nfqlb nfqlb-local ip6 meta l4proto icmpv6 daddr @ipv6Vips counter queue num 0-3 fanout.
	ipv6Rule := &nftables.Rule{
		Table: nfq.table,
		Chain: nfq.localchain,
		Exprs: []expr.Any{
			// [ meta load nfproto => reg 1 ]
			&expr.Meta{
				Key:      expr.MetaKeyNFPROTO,
				Register: 1,
			},
			// [ cmp eq reg 1 0x0000000a ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.AF_INET6},
			},
			// [ meta load l4proto => reg 1 ]
			&expr.Meta{
				Key:      expr.MetaKeyL4PROTO,
				Register: 1,
			},
			// [ cmp eq reg 1 0x0000003a ]
			&expr.Cmp{
				Op:       expr.CmpOpEq,
				Register: 1,
				Data:     []byte{unix.IPPROTO_ICMPV6},
			},
			// [ payload load 16b @ network header + 24 => reg 1 ]
			&expr.Payload{
				DestRegister: 1,
				Base:         expr.PayloadBaseNetworkHeader,
				Offset:       24,
				Len:          16,
			},
			// [ lookup reg 1 set flow-a-daddrs-v6 ]
			&expr.Lookup{
				SourceRegister: 1,
				SetName:        nfq.ipv6DestinationSet.Name,
				SetID:          nfq.ipv6DestinationSet.ID,
			},
			// [ counter pkts 0 bytes 0 ]
			&expr.Counter{
				Bytes:   0,
				Packets: 0,
			},
			// [ queue num 1 ]
			&expr.Queue{
				Num:   nfq.nftqueueNum,
				Total: nfq.nftqueueTotal,
				Flag:  nfq.nftqueueFlag,
			},
		},
	}
	nfq.ipv6Rule = conn.AddRule(ipv6Rule)

	err := conn.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush (configureLocalChainAndRules): %w", err)
	}

	return nil
}

// setDestinationCIDRs updates nftables Set based on the VIPs so that all traffic with VIP destination
// could be handled by the user space application connected to the configured queue(s).
func (nfq *netfilterQueue) setDestinationCIDRs(cidrList []string) error {
	ipv4s, ipv6s := getIPv4AndIPv6(cidrList)
	ipv4Elements := ipNetsToSetElements(ipv4s)
	ipv6Elements := ipNetsToSetElements(ipv6s)

	currentIpv4Elements, err := getCurrentElements(nfq.ipv4DestinationSet)
	if err != nil {
		return err
	}

	currentIpv6Elements, err := getCurrentElements(nfq.ipv6DestinationSet)
	if err != nil {
		return err
	}

	err = updateSet(nfq.ipv4DestinationSet, currentIpv4Elements, ipv4Elements)
	if err != nil {
		return err
	}

	err = updateSet(nfq.ipv6DestinationSet, currentIpv6Elements, ipv6Elements)
	if err != nil {
		return err
	}

	return nil
}
