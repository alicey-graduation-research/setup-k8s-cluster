#!/bin/bash
set -eu

# change dir
cd `dirname $0`

# マスタコントロールプレーンの初期化
kubeadm init
sleep 10
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config

kubectl apply -f ./k8s-manifests/admin-clusterrolebinding.yaml
kubectl apply -f ./k8s-manifests/calico.yaml


# setup token-notificator
cp ./kubeadm-token-notificator/$(dpkg --print-architecture)_notification_token.o /opt/k8s-token-notificator/notification_token.o
chmod u+x /opt/k8s-token-notificator/notification_token.o

sudo tee /etc/systemd/system/k8s-token-notificator.service <<EOF
[Unit]
Description=k8s-token-notificator
After=network-online.target
[Service]
ExecStart=/opt/k8s-token-notificator/notification_token.o
WorkingDirectory=/opt/k8s-token-notificator/
Restart=always
User=$(id -u -n)
Group=$(groups | cut -d ' ' -f 1)
[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable k8s-token-notificator.service
sudo systemctl start k8s-token-notificator.service