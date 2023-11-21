# Builder stage
ARG GOARCH
ARG GOARM
ARG BUILD_PLATFORM
ARG FINAL_IMAGE

ARG COMPILER_LINK
ARG COMPILER_BINARY_NAME

FROM golang:1.17-alpine AS builder

ENV GOPROXY=https://proxy.golang.org
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV CGO_ENABLED=0
ENV GOARCH=${GOARCH}
ENV GOARM=${GOARM}
ENV FINAL_IMAGE=${FINAL_IMAGE}
ENV COMPILER_LINK=${COMPILER_LINK}
ENV COMPILER_BINARY_NAME=${COMPILER_BINARY_NAME}

WORKDIR /app
ADD . .

ARG GOARCH
ENV GOARCH=${GOARCH}
ENV GOARM=${GOARM}
ENV COMPILER_LINK=${COMPILER_LINK}
ENV COMPILER_BINARY_NAME=${COMPILER_BINARY_NAME}

RUN apk add gcc g++ build-base

COPY go.mod go.sum ./
RUN go mod download

RUN env
RUN wget -O ~/compiler.tgz ${COMPILER_LINK}
RUN mkdir ~/compiler && tar -xvf ~/compiler.tgz -C ~/compiler --strip-components=1

# Build
RUN ls /
RUN ls ~/compiler
RUN CGO_ENABLED=1 CC=~/compiler/bin/$COMPILER_BINARY_NAME GOOS=linux go build -v -o haargos-out .

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
