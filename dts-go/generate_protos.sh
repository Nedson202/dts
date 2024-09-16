#!/bin/bash

# Exit on any error
set -e
set -x  # Enable command echoing

# Function to check and install Go packages
install_go_package() {
    if ! command -v $1 &> /dev/null; then
        echo "$1 is not installed. Installing..."
        go install $2
    else
        echo "$1 is already installed."
    fi
}

# Install required Go packages
install_go_package protoc-gen-go "google.golang.org/protobuf/cmd/protoc-gen-go@latest"
install_go_package protoc-gen-go-grpc "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
install_go_package protoc-gen-grpc-gateway "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest"

export PATH="$PATH:$(go env GOPATH)/bin"

# Ensure protoc is installed
command -v protoc >/dev/null 2>&1 || { echo >&2 "protoc is required but not installed. Please install it manually. Aborting."; exit 1; }

# Create necessary directories
mkdir -p pkg/job pkg/scheduler pkg/execution

# Generate Go code from proto files
protoc -I. \
    -I./api/proto \
    --go_out=. --go-grpc_out=. \
    --grpc-gateway_out=. \
    --go_opt=module=github.com/nedson202/dts-go \
    --go-grpc_opt=module=github.com/nedson202/dts-go \
    --grpc-gateway_opt=module=github.com/nedson202/dts-go \
    api/proto/*.proto

echo "Proto files generated and import paths updated successfully."

# Run go mod tidy to update dependencies
go mod tidy
go work sync

echo "Dependencies updated. Please review the changes and commit them."
