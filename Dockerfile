FROM golang:1.13-alpine AS builder

WORKDIR "/opt/go/src/github.com/seeruk/bedrockd"

ADD go.mod ./
ADD go.sum ./

RUN set -x \
    && go mod download

ADD . .
RUN set -x \
    && set -o pipefail \
    # && go test -cover ./... | grep -v "^?" \
    && CGO_ENABLED=0 go build \
        -ldflags "-s -w" \
        -o bedrockd \
        ./cmd/bedrockd/main.go

FROM ubuntu:bionic
LABEL maintainer="Elliot Wright <wright.elliot@gmail.com>"

ARG SERVER_VERSION=1.12.1.1

COPY docker/entrypoint.sh /opt/mcbuild/entrypoint.sh

RUN set -x \
    && apt-get update \
    && apt-get install -y curl unzip wget \
    && rm -rf /var/lib/apt/lists/* \
    && useradd -d /home/mcserver -u 1000 -m -s /bin/bash mcserver \
    && mkdir -p /opt/mcbuild \
    && mkdir -p /opt/mcserver \
    && cd /opt/mcbuild \
    && wget https://minecraft.azureedge.net/bin-linux/bedrock-server-${SERVER_VERSION}.zip \
    && unzip bedrock-server-${SERVER_VERSION}.zip \
    && rm bedrock-server-${SERVER_VERSION}.zip \
    && chown -R mcserver: /opt/mcbuild \
    && chown -R mcserver: /opt/mcserver \
    && chmod +x /opt/mcbuild/entrypoint.sh

COPY --chown=mcserver --from=builder \
    /opt/go/src/github.com/seeruk/bedrockd/bedrockd /opt/mcbuild/bedrockd

USER mcserver

VOLUME /opt/mcserver

ENTRYPOINT ["/opt/mcbuild/entrypoint.sh"]
