#! /bin/bash
set -e

# change dir
cd `dirname $0`

bash kubeadm-token-notificator/build.sh

GOOS=linux GOARCH=amd64 go build -o ./amd64_node_join.o node_join.go
GOOS=linux GOARCH=arm64 go build -o ./arm64_node_join.o node_join.go