#! /bin/bash

sysctl -w net.ipv6.conf.all.disable_ipv6=0
sysctl -w net.ipv4.fib_multipath_hash_policy=1
sysctl -w net.ipv6.fib_multipath_hash_policy=1
sysctl -w net.ipv6.conf.all.forwarding=1
sysctl -w net.ipv4.conf.all.forwarding=1
sysctl -w net.ipv6.conf.all.accept_dad=0

ip link add link eth0 name vlan1 type vlan id $1
ip link set vlan1 up
ip addr add 169.254.100.150/24 dev vlan1
ip addr add 100:100::150/64 dev vlan1
ip addr add 200.100.0.100/32 dev vlan1
ip addr add 200:100::100/128 dev vlan1

ethtool -K eth0 tx off

sh -c "echo \"PS1='VPN Gateway/Traffic-Generator | VLAN:$1> '\" >> ~/.bashrc"

/usr/sbin/bird -d -c /etc/bird/bird-gw.conf
