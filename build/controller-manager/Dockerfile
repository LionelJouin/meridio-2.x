ARG BUILD_VERSION=0.0.0-unknown

FROM golang:1.23 AS build

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-extldflags -static -X github.com/lioneljouin/meridio-experiment/pkg/cli.Version=${BUILD_VERSION}" -o controller-manager ./cmd/controller-manager

FROM alpine

COPY --from=build /app/controller-manager .

CMD ["./controller-manager", "run"]
