#!/bin/bash
set -eu

# マスタコントロールプレーンの初期化
kubeadm init
sleep 10
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
kubectl apply -f ./k8s-manifests/calico.yaml