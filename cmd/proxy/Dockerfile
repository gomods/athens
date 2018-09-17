FROM golang:1.11-alpine AS builder

WORKDIR $GOPATH/src/github.com/gomods/athens

COPY . .

RUN apk update && \ 
	apk add --no-cache bash && \
	GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -o /bin/app ./cmd/proxy && \
	./scripts/create_default_config.sh

FROM alpine

ENV GO111MODULE=on

COPY --from=builder /bin/app /bin/app
COPY --from=builder /go/src/github.com/gomods/athens/config.toml /config/config.toml
COPY --from=builder /go/src/github.com/gomods/athens/cmd/proxy/locales /go/src/github.com/gomods/athens/cmd/proxy/locales
COPY --from=builder /go/src/github.com/gomods/athens/cmd/proxy/assets /go/src/github.com/gomods/athens/cmd/proxy/assets
COPY --from=builder /usr/local/go/bin/go /bin/go

RUN apk update &&  \
		apk add --no-cache bzr git mercurial openssh-client subversion procps && \
		mkdir -p /usr/local/go

ENV GO_ENV=production

EXPOSE 3000

ENTRYPOINT ["app", "-config_file=/config/config.toml"]
