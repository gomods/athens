FROM alpine:3.8 as builder

ARG HUGO_VERSION=0.125.1
ENV HUGO_BINARY=hugo_${HUGO_VERSION}_Linux-64bit.tar.gz

ADD https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/${HUGO_BINARY} /tmp

RUN tar -xf /tmp/${HUGO_BINARY} -C /tmp

FROM alpine:3.8

COPY --from=builder /tmp/hugo /usr/local/bin/hugo

RUN apk upgrade --update \
    && apk add --no-cache git asciidoctor libc6-compat libstdc++ ca-certificates

WORKDIR /src

CMD ["hugo", "server", "-s", "/src", "-b", "http://localhost:1313", "--bind", "0.0.0.0"]

EXPOSE 1313
