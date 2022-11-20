#!/bin/bash
set -eu

# change dir
cd `dirname $0`

# setup worker
mkdir -p /opt/k8s-session-observer/
cp ./$(dpkg --print-architecture)_node_join.o /opt/k8s-session-observer/k8s-session-observer.o
chmod u+x /opt/k8s-session-observer/k8s-session-observer.o

sudo tee /etc/systemd/system/k8s-session-observer.service <<EOF
[Unit]
Description=k8s-session-observer
After=network-online.target
[Service]
ExecStart=/opt/k8s-session-observer/k8s-session-observer.o
WorkingDirectory=/opt/k8s-session-observer/
Restart=always
User=$(id -u -n)
Group=$(groups | cut -d ' ' -f 1)
[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable k8s-session-observer.service
sudo systemctl start k8s-session-observer.service