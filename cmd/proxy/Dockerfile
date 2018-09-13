FROM golang:1.11-alpine AS builder

RUN mkdir -p $GOPATH/src/github.com/gomods/athens
WORKDIR $GOPATH/src/github.com/gomods/athens

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -o /bin/app ./cmd/proxy
RUN ./scripts/create_default_config.sh

FROM alpine

ENV GO111MODULE=on
COPY --from=builder /bin/app /bin/app
COPY --from=builder /go/src/github.com/gomods/athens/cmd/proxy/locales /go/src/github.com/gomods/athens/cmd/proxy/locales
COPY --from=builder /go/src/github.com/gomods/athens/cmd/proxy/assets /go/src/github.com/gomods/athens/cmd/proxy/assets
COPY --from=builder /usr/local/go/bin/go /bin/go

RUN apk update &&  \
		apk add --no-cache bzr git mercurial openssh-client subversion procps

ENV GO_ENV=production

EXPOSE 3000

ENTRYPOINT /bin/app -config_file=config.toml
