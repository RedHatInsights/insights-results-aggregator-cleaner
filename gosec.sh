#!/bin/bash

# Copyright 2020, 2021 Red Hat, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

GO_SEC_ARGS=""

VERBOSE=false
if [[ $* == *verbose* ]]; then
    VERBOSE=true
fi

if [[ $* != *verbose* ]]; then
    GO_SEC_ARGS="-quiet"
fi

cd "$(dirname "$0")" || exit

echo -e "${BLUE}Security issues detection${NC}"

if ! [ -x "$(command -v gosec)" ]
then
    echo -e "${BLUE}Installing ${NC}"
    GO111MODULE=off go get github.com/securego/gosec/cmd/gosec 2> /dev/null
fi

if ! gosec $GO_SEC_ARGS ./...
then
    echo -e "${RED}[FAIL]${NC} Potential security issues detected!"
    exit 1
else
    echo -e "${GREEN}[OK]${NC} No potential security issues has been detected"
    exit 0
fi
