#!/bin/bash
#
# Usage: restart.sh
#
BIN_DIR=$(dirname "$0")
cd "${BIN_DIR}" || exit
./stop.sh "$@"
./start.sh "$@"