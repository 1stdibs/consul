#!/bin/bash
pushd $(dirname ${BASH_SOURCE[0]}) > /dev/null
SCRIPT_DIR=$(pwd)
pushd ../.. > /dev/null
SOURCE_DIR=$(pwd)
popd > /dev/null
pushd ../functions > /dev/null
FN_DIR=$(pwd)
popd > /dev/null
popd > /dev/null

source "${SCRIPT_DIR}/functions.sh"

# set up the docker images we're going to build consul with
refresh_docker_images "${SOURCE_DIR}"

# we're only building consul for an alpine container, so we only need one OS and ARCH
XC_OS="linux"
XC_ARCH="amd64"

build_consul_release "${SOURCE_DIR}" "consul-build-go"
