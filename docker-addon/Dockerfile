# Builder stage
ARG GOARCH
ARG GOARM
ARG BUILD_PLATFORM
ARG FINAL_IMAGE

FROM golang:1.17-alpine AS builder

ENV GOPROXY=https://proxy.golang.org
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GOARCH=${GOARCH}
ENV GOARM=${GOARM}
ENV FINAL_IMAGE=${FINAL_IMAGE}

WORKDIR /app
ADD . .

ARG GOARCH
ENV GOARCH=${GOARCH}
ENV GOARM=${GOARM}

RUN apk add build-base

COPY go.mod go.sum ./
RUN go mod download

# Build
RUN CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc GOOS=linux go build -v -o haargos-out .

# Final stage
FROM "$FINAL_IMAGE"
WORKDIR /root/

COPY --chmod=755 docker-addon/run.sh ./run.sh
COPY --from=builder /app/haargos-out ./haargos

LABEL \
  io.hass.version="1.0.72" \
  io.hass.type="addon" \
  io.hass.arch="armhf|aarch64|i386|amd64"

CMD ["./run.sh"]
