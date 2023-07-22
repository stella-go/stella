#!/bin/bash
#
# Usage: stop.sh
#
echo "Stoping..."

BIN_DIR=$(dirname "$0")
cd "${BIN_DIR}" || exit
BIN_DIR=$(pwd)
cd ..
DEPLOY_DIR=$(pwd)
CONFIG_DIR="$DEPLOY_DIR/config"

APP=$(grep -v '^\s*#' "${CONFIG_DIR}/application.yml" | sed '/application: /!d;s/.*: //' | tr -d '\r')
echo "APP=${APP}"
SERVER_PORT=$(grep -v '^\s*#' "${CONFIG_DIR}/application.yml" | sed '/port: /!d;s/.*: //' | tr -d '\r')
echo "SERVER_PORT=${SERVER_PORT}"

EXE=${BIN_DIR}/${APP}
echo "EXE=${EXE}"

PIDS=$(netstat -nltp 2>/dev/null | grep -w "$SERVER_PORT" | awk '{print $7}' | awk '{split($1,a,"/");print a[1]}')
if [ -z "$PIDS" ]; then
    PIDS=$(ps -ef | grep "${EXE}" | grep -v grep | awk '{print $2}')
    echo "PIDS: ${PIDS}"
    if [ -z "${PIDS}" ]; then
        echo "ERROR: The application ${APP} does not started!"
        exit 1
    fi
fi

echo "Stopping the application ${APP}[PIDS: ${PIDS}] ..."
for PID in ${PIDS} ; do
    kill "${PID}" > /dev/null 2>&1
done

MAX_WAIT=10
COUNT=0

while [ ${COUNT} -le ${MAX_WAIT} ]; do
    sleep 1
    for PID in ${PIDS} ; do
        PID_EXIST=$(ps -ef | grep "${PID}" | grep "${EXE}")
        if [ -n "${PID_EXIST}" ]; then
            break 1
        else
            break 2
        fi
    done
    ((COUNT=COUNT+1))
done

for PID in ${PIDS} ; do
    PID_EXIST=$(ps -ef | grep "${PID}" | grep "${EXE}")
    if [ -n "${PID_EXIST}" ]; then
        echo "Force to terminate the application ${APP}[PID: ${PID}] ..."
        kill -9 "${PID}"
    fi
done

echo "OK!"
echo "PID: ${PIDS}"
