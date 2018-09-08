FROM golang:1.11-alpine

RUN mkdir -p $GOPATH/src/github.com/gomods/athens
WORKDIR $GOPATH/src/github.com/gomods/athens

COPY . .

RUN apk update &&  \
		apk add --no-cache bzr git mercurial openssh-client subversion procps && \
		GO111MODULE=on CGO_ENABLED=0 go build -mod=vendor -o /bin/app ./cmd/proxy

ENV GO_ENV=production

EXPOSE 3000

ENTRYPOINT ["/bin/app"]
