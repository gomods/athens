FROM gobuffalo/buffalo:v0.12.6 as buffalo-builder

RUN mkdir -p $GOPATH/src/github.com/gomods/athens/cmd/proxy
WORKDIR $GOPATH/src/github.com/gomods/athens/cmd/proxy

# this will cache the npm install step, unless package.json changes
ADD cmd/proxy/package.json .
ADD cmd/proxy/yarn.lock .
RUN yarn install --no-progress

RUN mkdir -p $GOPATH/src/github.com/gomods/athens
WORKDIR $GOPATH/src/github.com/gomods/athens

ADD . .
RUN cd cmd/proxy && buffalo build -s -o /bin/app


FROM golang:1.11-rc-alpine
RUN apk add --no-cache bash git openssh ca-certificates

COPY --from=buffalo-builder /bin/app /bin/app

ENV GO_ENV=production
# Bind the app to 0.0.0.0 (all interfaces) so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

ENTRYPOINT ["/bin/app"]
