# Builder stage
FROM --platform=linux/amd64 golang:1.17 AS builder
ENV GOPROXY=https://proxy.golang.org
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN apt-get update && apt-get upgrade -y \
    && apt-get install -y libsqlite3-dev build-essential g++-x86-64-linux-gnu libc6-dev-amd64-cross \
    && apt-get install -y gcc-multilib g++-multilib libc6-dev-i386 \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory and add application code
WORKDIR /app
ADD . .

# Copy go mod and sum files and fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

# Set environment for cgo
ENV CGO_ENABLED=1

# Build for amd64
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -o app-amd64 .

# Build for 386
RUN CGO_ENABLED=1 GOOS=linux GOARCH=386 go build -v -o app-386 .

# Final stage
FROM multiarch/ubuntu-core:amd64-bionic
WORKDIR /root/
COPY --from=builder /app/app-amd64 ./app-amd64
COPY --from=builder /app/app-386 ./app-386
CMD ["./app-amd64"]
