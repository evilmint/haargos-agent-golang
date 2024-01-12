ARG GOARCH
ARG GOARM
ARG FINAL_IMAGE
ARG HAARGOS_VERSION
ARG COMPILER_LINK
ARG COMPILER_BINARY_NAME
ARG RUN_SCRIPT_NAME

# Builder stage
FROM golang:1.17-alpine AS builder

ARG GOARCH
ARG GOARM
ARG FINAL_IMAGE
ARG HAARGOS_VERSION
ARG COMPILER_LINK
ARG COMPILER_BINARY_NAME
ARG RUN_SCRIPT_NAME

ENV GOPROXY=https://proxy.golang.org \
    GO111MODULE=on \
    GOOS=linux \
    CGO_ENABLED=0 \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    FINAL_IMAGE=${FINAL_IMAGE} \
    COMPILER_LINK=${COMPILER_LINK} \
    COMPILER_BINARY_NAME=${COMPILER_BINARY_NAME} \
    GOARCH=${GOARCH} \
    GOARM=${GOARM} \
    COMPILER_LINK=${COMPILER_LINK} \
    COMPILER_BINARY_NAME=${COMPILER_BINARY_NAME}

WORKDIR /app
ADD . .

RUN apk add gcc g++ build-base

COPY go.mod go.sum ./
RUN go mod download

RUN env
RUN wget -O ~/compiler.tgz "$COMPILER_LINK"
RUN mkdir ~/compiler && tar -xvf ~/compiler.tgz -C ~/compiler --strip-components=1

# Build
RUN ls /
RUN ls ~/compiler
RUN CGO_ENABLED=1 CC=~/compiler/bin/$COMPILER_BINARY_NAME GOOS=linux go build -v -o haargos-out .

# Final stage
FROM "$FINAL_IMAGE"
WORKDIR /root/

ARG HAARGOS_VERSION
ARG RUN_SCRIPT_NAME

COPY --chmod=755 "docker-addon/$RUN_SCRIPT_NAME" ./run.sh
COPY --chmod=755 "docker-addon/templates" ./templates
COPY --chmod=755 "docker-addon/static" ./static
COPY --from=builder /app/haargos-out ./haargos

RUN echo "HAARGOS_VERSION is '$HAARGOS_VERSION'"
RUN echo "$HAARGOS_VERSION" > ./VERSION

LABEL \
  io.hass.version="$HAARGOS_VERSION" \
  io.hass.type="addon" \
  io.hass.arch="armhf|aarch64|i386|amd64"

CMD ["./run.sh"]
