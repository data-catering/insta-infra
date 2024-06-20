#!/bin/bash
#
# Copyright 2018-2023 contributors to the Marquez project
# SPDX-License-Identifier: Apache-2.0
#
# Usage: $ ./init.sh

set -e

java -jar /usr/src/app/marquez-api-*.jar seed --url "${MARQUEZ_URL:-http://localhost:5000}" --metadata /tmp/data/metadata.json
