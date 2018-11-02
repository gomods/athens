FROM alpine:3.8

ENV HUGO_VERSION=0.50
ENV HUGO_BINARY=hugo_${HUGO_VERSION}_Linux-64bit.tar.gz

ADD https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_BINARY} /tmp

RUN tar -xf /tmp/${HUGO_BINARY} -C /tmp \
    && mv /tmp/hugo /usr/local/bin/hugo \
    && rm -rf /tmp/hugo_${HUGO_VERSION}_linux_amd64 \
    && rm -rf /tmp/${HUGO_BINARY} \
    && rm -rf /tmp/LICENSE.md \
    && rm -rf /tmp/README.md \
    && apk upgrade --update \
    && apk add --no-cache git asciidoctor libc6-compat libstdc++ ca-certificates

WORKDIR /src

CMD ["hugo", "server", "-s", "/src", "-b", "http://localhost:1313", "--bind", "0.0.0.0"]

EXPOSE 1313
