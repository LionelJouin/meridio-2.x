ARG BUILD_VERSION=0.0.0-unknown

FROM golang:1.23 AS build

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-extldflags -static -X github.com/lioneljouin/meridio-experiment/pkg/cli.Version=${BUILD_VERSION}" -o stateless-load-balancer ./cmd/stateless-load-balancer

FROM ubuntu:22.04 AS nfqlb-build

RUN apt-get update && apt install -y git build-essential libmnl-dev libnetfilter-queue-dev libpcap-dev curl \
    && git clone --depth 1 --branch 1.2.0 https://github.com/Nordix/nfqueue-loadbalancer.git

WORKDIR /nfqueue-loadbalancer

ENV ARCHIVE=/tmp

RUN curl -L https://netfilter.org/projects/libmnl/files/libmnl-1.0.4.tar.bz2 > /tmp/libmnl-1.0.4.tar.bz2 \
    && chmod 777 /tmp/* \
    && /nfqueue-loadbalancer/nfqlb.sh libmnl_unpack \
    && /nfqueue-loadbalancer/nfqlb.sh libmnl_build \
    && make -C src -j8 X=/bin/nfqlb static

FROM alpine

RUN apk update && apk add iproute2 tcpdump nftables

COPY --from=build /app/stateless-load-balancer .
COPY --from=nfqlb-build /bin/nfqlb /bin/nfqlb

CMD ["./stateless-load-balancer", "run"]
