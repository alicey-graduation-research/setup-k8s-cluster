#! /bin/bash
set -e

# change dir
cd `dirname $0`


GOOS=linux GOARCH=amd64 go build -o ./amd64_notification.token.o notification_token.go
GOOS=linux GOARCH=arm64 go build -o ./arm64_notification.token.o notification_token.go