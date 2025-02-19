FROM golang:1.23.5-bookworm

WORKDIR /tmp

# Install Helm
ENV HELM_VERSION=2.13.0
RUN curl -sLO https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz && \
    tar -zxvf helm-v${HELM_VERSION}-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/

# Install a tiny azure client
ENV AZCLI_VERSION=v0.3.2
RUN curl -sLo /usr/local/bin/az https://github.com/carolynvs/az-cli/releases/download/$AZCLI_VERSION/az-linux-amd64 && \
chmod +x /usr/local/bin/az

WORKDIR /go/src/github.com/gomods/athens
