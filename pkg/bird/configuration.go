package bird

import (
	"fmt"

	"github.com/lioneljouin/meridio-experiment/apis/v1alpha1"
)

// getConfig combines the logs, base, vips and gateways configs.
func (b *Bird) getConfig(vips []string, gateways []Gateway) string {
	conf := fmt.Sprintf("%s\n\n%s",
		b.logConfig(),
		fmt.Sprintf(baseConfig, defaultKernelTableID, defaultKernelTableID),
	)

	vipsConfig := vipsConfig(vips)
	if vipsConfig != "" {
		conf = fmt.Sprintf("%s\n\n%s", conf, vipsConfig)
	}

	gatewaysConfig := gatewaysConfig(gateways)
	if gatewaysConfig != "" {
		conf = fmt.Sprintf("%s\n\n%s", conf, gatewaysConfig)
	}

	return conf
}

func (b *Bird) logConfig() string {
	conf := ""

	if b.LogFileSize > 0 {
		conf += fmt.Sprintf("log \"%s\" %v \"%s\" { debug, trace, info, remote, warning, error, auth, fatal, bug };\n",
			b.LogFile, b.LogFileSize, b.LogFileBackup)
	}

	if b.LogEnabled {
		conf += "log stderr all;"
	} else {
		conf += "log stderr { error, fatal, bug, warning };"
	}

	return conf
}

// vipsConfig create static routes for VIP addresses in BIRD config.
//
// VIP addresses are configured as static routes in BIRD. They are
// only advertised to BGP peers and not synced into local network stack.
//
// Note: VIPs shall be advertised only if external connectivity is OK.
func vipsConfig(vips []string) string {
	ipv4, ipv6 := "", ""

	for _, vip := range vips {
		if isIPv6CIDR(vip) {
			ipv6 += fmt.Sprintf(vipRouteTemplate, vip)
		} else if isIPv4CIDR(vip) {
			ipv4 += fmt.Sprintf(vipRouteTemplate, vip)
		}
	}

	conf := ""

	if ipv4 != "" {
		conf += fmt.Sprintf(vipsTemplate, "VIP4", "ipv4", ipv4)
	}

	if ipv6 != "" {
		conf += fmt.Sprintf(vipsTemplate, "VIP6", "ipv6", ipv6)
	}

	return conf
}

// gatewaysConfig creates BGP proto part of the BIRD config for each gateway to connect with them
//
// BGP is restricted to the external interface.
// Only VIP related routes are announced to peer, and only default routes are accepted.
//
// Note: When VRRP IPs are configured, BGP sessions won't import any routes from external
// peers, as external routes are going to be taken care of by static default routes (VRRP IPs
// as next hops).
func gatewaysConfig(gateways []Gateway) string {
	conf := ""

	for _, gateway := range gateways {
		conf += gatewayConfig(gateway)
		conf += "\n\n"
	}

	conf += bfdTemplate

	return conf
}

func gatewayConfig(gateway Gateway) string {
	conf := ""

	switch gateway.GetProtocol() {
	case v1alpha1.BGP:
		conf += bgpConfig(gateway)
	case v1alpha1.Static: // todo: static
	}

	return conf
}

func bgpConfig(gateway Gateway) string {
	ipFamily := ""

	if isIPv4(gateway.GetAddress()) {
		ipFamily = "ipv4"
	} else if isIPv6(gateway.GetAddress()) {
		ipFamily = "ipv6"
	}

	localPort := defaultLocalPort
	if gateway.GetBgpSpec().GetLocalPort() != nil {
		localPort = *gateway.GetBgpSpec().GetLocalPort()
	}

	localASN := defaultLocalASN
	if gateway.GetBgpSpec().GetLocalPort() != nil {
		localASN = *gateway.GetBgpSpec().GetLocalASN()
	}

	remotePort := defaultRemotePort
	if gateway.GetBgpSpec().GetRemotePort() != nil {
		remotePort = *gateway.GetBgpSpec().GetRemotePort()
	}

	remoteASN := defaultRemoteASN
	if gateway.GetBgpSpec().GetRemotePort() != nil {
		remoteASN = *gateway.GetBgpSpec().GetRemoteASN()
	}

	return fmt.Sprintf(bgpTemplate,
		gateway.GetName(),
		gateway.GetInterface(),
		localPort,
		localASN,
		gateway.GetAddress(),
		remotePort,
		remoteASN,
		bfdConfig(gateway.GetBgpSpec().GetBfdSpec()),
		defaultBGPHoldTime,
		ipFamily,
	)
}

func bfdConfig(bfd BfdSpec) string {
	if bfd.GetSwitch() == nil || !*bfd.GetSwitch() {
		return "\tbfd off;"
	}

	conf := ""

	if bfd.GetMinRx() != "" {
		conf += fmt.Sprintf("\t\tmin rx interval %s;\n", bfd.GetMinRx())
	}

	if bfd.GetMinTx() != "" {
		conf += fmt.Sprintf("\t\tmin tx interval %s;\n", bfd.GetMinTx())
	}

	if bfd.GetMultiplier() != nil {
		conf += fmt.Sprintf("\t\tmultiplier %d;\n", *bfd.GetMultiplier())
	}

	if conf != "" {
		conf = fmt.Sprintf("\n%s\t", conf)
	}

	return fmt.Sprintf(bgpBfdTemplate, conf)
}
