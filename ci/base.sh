#!/bin/bash
set -e

# Check go version
go version

# Setup GOPATH 
export ORG_PATH="github.com/rakutentech"
export REPO_PATH="${ORG_PATH}/cf-metrics-refinery"

mkdir -p /gopath/src/${ORG_PATH}

ln -s ${PWD} /gopath/src/${REPO_PATH}

cd /gopath/src/${REPO_PATH}