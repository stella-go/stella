#!/bin/bash
#
# Usage: start.sh
#
echo "Starting..."

BIN_DIR=$(dirname "$0")
cd "${BIN_DIR}" || exit
BIN_DIR=$(pwd)
cd ..
DEPLOY_DIR=$(pwd)
CONFIG_DIR=$DEPLOY_DIR/config
LIB_DIR=$DEPLOY_DIR/lib

APP=$(grep -v '^\s*#' "${CONFIG_DIR}/application.yml" | sed '/application: /!d;s/.*: //' | tr -d '\r')
echo "APP=${APP}"
SERVER_PORT=$(grep -v '^\s*#' "${CONFIG_DIR}/application.yml" | sed '/port: /!d;s/.*: //' | tr -d '\r')
echo "SERVER_PORT=${SERVER_PORT}"
LOG_PATH=$(grep -v '^\s*#' "${CONFIG_DIR}/application.yml" | sed '/^  path: /!d;s/.*: //' | tr -d '\r')
if [ -z "$LOG_PATH" ]; then
  LOG_PATH=$(grep -v '^\s*#' "${CONFIG_DIR}/application.yml" | sed '/logger.path: /!d;s/.*: //' | tr -d '\r')
  if [ -z "${LOG_PATH}" ]; then
    LOG_PATH=${DEPLOY_DIR}/logs
  fi
fi
echo "LOG_PATH=${LOG_PATH}"

if [ -z "${APP}" ]; then
  echo "ERROR: Not Found APP defined in ${CONFIG_DIR}/application.yml"
  exit 1
fi

if [ -z "${SERVER_PORT}" ]; then
  echo "ERROR: Not Found SERVER_PORT defined in ${CONFIG_DIR}/application.yml"
  exit 1
fi

if [ -z "${LOG_PATH}" ]; then
  echo "ERROR: Not Found LOG_PATH defined in ${CONFIG_DIR}/application.yml"
  exit 1
fi

if [ ! -d "$LOG_PATH" ]; then
    mkdir -p "$LOG_PATH"
fi
if [ ! -d "$LOG_PATH" ]; then
  echo "ERROR: Please check LOG_PATH=$LOG_PATH is ok?"
  exit 1
fi
EXE="${BIN_DIR}/${APP}"
echo "EXE=${EXE}"

PIDS=$(ps -ef | grep "${EXE}" | grep -v grep | awk '{print $2}')
if [ -n "${PIDS}" ]; then
    echo "ERROR: The application ${APP} already started!"
    echo "PID: ${PIDS}"
    exit 1
fi

SERVER_PORT_COUNT=$(netstat -nltp 2>/dev/null | grep -wc "$SERVER_PORT")
if [ ${SERVER_PORT_COUNT} -gt 0 ]; then
    echo "ERROR: The port $SERVER_PORT already used!"
    exit 1
fi

STDOUT_FILE=${LOG_PATH}/console.log

export LD_LIBRARY_PATH="${LIB_DIR}:${LD_LIBRARY_PATH}"
echo "Starting the ${APP} ..."
nohup "${EXE}" > "${STDOUT_FILE}" 2>&1 &

COUNT=0
while [ $COUNT -lt 1 ]; do
    sleep 1
    if [ -n "${SERVER_PORT}" ]; then
        COUNT=$(netstat -nltp 2>/dev/null | grep -cw "$SERVER_PORT")
    fi
    if [ $COUNT -lt 1 ]; then
        COUNT=$(ps -ef | grep "${EXE}" | grep -v grep | awk '{print $2}' | wc -l)
    fi
    if [ $COUNT -gt 0 ]; then
        break
    fi
    echo "check $(date '+%Y-%m-%d %H:%M:%S')"
done

PIDS=$(ps -ef | grep "${EXE}" | grep -v grep | awk '{print $2}')

echo "OK!"
echo "PID: ${PIDS}"