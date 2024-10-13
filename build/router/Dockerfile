ARG BUILD_VERSION=0.0.0-unknown

FROM golang:1.22 as build

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-extldflags -static -X github.com/lioneljouin/meridio-experiment/pkg/cli.Version=${BUILD_VERSION}" -o router ./cmd/router

FROM alpine

RUN apk update && apk add iproute2 tcpdump nftables bird

RUN mkdir -p /run/bird && mkdir -p /etc/bird

COPY --from=build /app/router .

CMD ["./router", "run"]
