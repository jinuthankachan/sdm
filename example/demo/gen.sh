#!/bin/bash
set -e

# 0. Create placeholder to satisfy go mod tidy (bootstrapping problem)
# We need `go mod` to see the 'user' package so it doesn't fail on main.go imports.
mkdir -p proto/user
echo 'package user' > proto/user/placeholder.go

# 1. Vendor the dependencies to make proto files available
go mod tidy
go mod vendor

# 2. Build/Install necessary plugins
# - protoc-gen-sdm (from local source)
(cd ../../ && go build -o example/demo/protoc-gen-sdm ./cmd/protoc-gen-sdm)

# - protoc-gen-go (ensure it's installed)
# We install it to GOBIN or verify it's in path. 
# We can install it to the current dir to be safe.
GOBIN=$(pwd) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

export PATH=$PATH:$(pwd)

# 3. Define include paths
# -I . : Current directory
# -I vendor : Vendored dependencies (where sdm/annotations.proto lives)

# 4. Run protoc
protoc --plugin=protoc-gen-sdm \
       --plugin=protoc-gen-go=$(pwd)/protoc-gen-go \
       --sdm_out=. --sdm_opt=paths=source_relative \
       --go_out=. --go_opt=paths=source_relative \
       -I . \
       -I vendor \
       proto/user/user.proto

# Cleanup placeholder
rm proto/user/placeholder.go

echo "Code generation successful!"
