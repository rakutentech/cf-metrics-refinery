#!/bin/bash
set -e

# Change the directory to root
DIR=$(cd $(dirname ${0})/.. && pwd)
cd ${DIR}

# Run base script
source ./ci/base.sh

# Change directory to repo
# We need to be in GOPATH/src to use vendor directory
cd /gopath/src/${REPO_PATH}

# Run the unit test
echo "Check go fmt & go vet errors"
make go-fmt
make go-vet