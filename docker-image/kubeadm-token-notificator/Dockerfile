FROM ubuntu:20.04

RUN apt-get update \
    && apt-get install -y curl apt-transport-https gnupg2 \
    && curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" >> /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update \
    && apt-get install -y kubeadm=1.24.1-00 kubectl=1.24.1-00 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/app
ADD notification_token ./app
RUN chmod u+x ./app

ENTRYPOINT ["/opt/app/app"]
CMD [""]