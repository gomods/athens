FROM gobuffalo/buffalo:v0.12.6 as builder

RUN mkdir -p $GOPATH/src/github.com/gomods/athens/cmd/olympus
WORKDIR $GOPATH/src/github.com/gomods/athens/cmd/olympus

# this will cache the npm install step, unless package.json changes
ADD cmd/olympus/package.json .
ADD cmd/olympus/yarn.lock .
RUN yarn install --no-progress

WORKDIR $GOPATH/src/github.com/gomods/athens
ADD . .

RUN cd cmd/olympus && buffalo build -s -o /bin/app


FROM golang:1.11-rc-alpine
RUN apk add --no-cache bash git openssh ca-certificates

COPY --from=builder /bin/app /bin/app

# Comment out to run the binary in "production" mode:
# ENV GO_ENV=production

# Bind the app to 0.0.0.0 so it can be seen from outside the container
ENV ADDR=0.0.0.0

EXPOSE 3000

# Comment out to run the migrations before running the binary:
# CMD /bin/app migrate; /bin/app
CMD exec /bin/app
