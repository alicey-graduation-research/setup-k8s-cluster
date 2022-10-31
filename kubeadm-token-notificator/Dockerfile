# go-builder
FROM golang:1.19.1-buster AS go-builder
WORKDIR /opt/app
COPY ./notification_token.go ./app.go
RUN go build app.go 

# running-image
FROM ubuntu:20.04

RUN apt-get update \
    && apt-get install -y curl apt-transport-https gnupg2 \
    && curl -s https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

RUN echo "deb https://apt.kubernetes.io/ kubernetes-xenial main" >> /etc/apt/sources.list.d/kubernetes.list
RUN apt-get update \
    && apt-get install -y kubeadm=1.24.1-00 kubectl=1.24.1-00 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt/app
COPY --from=go-builder /opt/app/app ./app
RUN chmod u+x ./app

ENTRYPOINT ["/opt/app/app"]
CMD [""]