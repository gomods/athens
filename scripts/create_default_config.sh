#!/bin/bash

SCRIPTS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

if [ ! -e "${SCRIPTS_DIR}/../config.toml" ] ; then
    cp "${SCRIPTS_DIR}/../config.example.toml" "${SCRIPTS_DIR}/../config.toml"
fi