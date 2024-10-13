FROM ubuntu:24.04

RUN apt-get update -y \
  && apt-get install -y iproute2 tcpdump net-tools iputils-ping netcat-traditional wget screen xz-utils strace nftables ipvsadm iptables bird2 ethtool conntrack dnsutils jq iperf iperf3

ADD https://github.com/Nordix/ctraffic/releases/download/v1.7.0/ctraffic.gz ctraffic.gz
RUN gunzip ctraffic.gz \
  && chmod a+x ctraffic

ADD https://github.com/Nordix/mconnect/releases/download/v2.2.0/mconnect.xz mconnect.xz
RUN unxz mconnect.xz \
  && chmod a+x mconnect

CMD ["tail", "-f", "/dev/null"]