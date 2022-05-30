#! /bin/bash

# user check
if [ "$(id -u)" -ne 0 ]; then
    echo "This script must be run as root user"
    exit 1
else
    echo "script start ..."
fi

# change dir
cd `dirname $0`

# swap off
swapoff -a
sed -i -e '/swap/d' /etc/fstab


# iptablesがnftablesバックエンドを使用しないようにする
apt-get update && \
    apt-get install -y iptables arptables ebtables

update-alternatives --set iptables /usr/sbin/iptables-legacy
update-alternatives --set ip6tables /usr/sbin/ip6tables-legacy
update-alternatives --set arptables /usr/sbin/arptables-legacy
update-alternatives --set ebtables /usr/sbin/ebtables-legacy


# install runtime (Containerd)
cat > /etc/modules-load.d/containerd.conf <<EOF
overlay
br_netfilter
EOF

modprobe overlay
modprobe br_netfilter

cat > /etc/sysctl.d/99-kubernetes-cri.conf <<EOF
net.bridge.bridge-nf-call-iptables  = 1
net.ipv4.ip_forward                 = 1
net.bridge.bridge-nf-call-ip6tables = 1
EOF
sysctl --system

apt-get update && \
    apt-get install -y apt-transport-https ca-certificates curl software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -

add-apt-repository \
    "deb [arch=$(dpkg --print-architecture)] https://download.docker.com/linux/ubuntu \
    $(lsb_release -cs) \
    stable"

## install
apt-get update && \
    apt-get install -y containerd.io

## setting containerd
mkdir -p /etc/containerd
containerd config default | tee /etc/containerd/config.toml

systemctl restart containerd


# install component
apt-get update && \
    apt-get install -y apt-transport-https curl
curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

cat <<EOF | tee /etc/apt/sources.list.d/kubernetes.list
deb https://apt.kubernetes.io/ kubernetes-xenial main
EOF

apt-get update && \
    apt-get install -y kubelet kubeadm kubectl && \
    apt-mark hold kubelet kubeadm kubectl

KUBELET_EXTRA_ARGS=--cgroup-driver=/run/containerd/containerd.sock

systemctl daemon-reload
systemctl restart kubelet

# コントロールプレーンの動作
kubeadm init
sleep 10
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
kubectl apply -f ./k8s-manifests/calico.yaml