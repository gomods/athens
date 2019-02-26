#!/bin/bash

./main & # run the just-built athens server
sleep 3 # wait for it to spin up
mkdir -p ~/happy
mkdir -p ~/emptygopath # ensure the gopath has no modules cached.
cd ~/happy
git clone https://github.com/athens-artifacts/happy-path.git
cd happy-path
GOPATH=~/emptygopath GOPROXY=http://localhost:3000 go build