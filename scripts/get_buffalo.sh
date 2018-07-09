#!/bin/bash

BUFFALO_URL="https://github.com/gobuffalo/buffalo/releases/download/v0.12.3/buffalo_0.12.3_linux_amd64.tar.gz"
TAR_GZ="buffalo_0.12.3_linux_amd64.tar.gz"

curl -L -o ${TAR_GZ} ${BUFFALO_URL}
tar -xvf ${TAR_GZ}
mv buffalo-no-sqlite /usr/local/bin/buffalo
