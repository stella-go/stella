#!/bin/bash

#
# Usage: build.sh [dev|beta|product]
#

echo "Building ..."

PROJECT_DIR=$(dirname "$0")
cd "${PROJECT_DIR}" || exit 1
PROJECT_DIR=$(pwd)
PROFILE="product"
if [ -n "$1" ]; then
  PROFILE=$1
fi
echo "Building profile: ${PROFILE}"

CONFIG_DIR="${PROJECT_DIR}/config"
PROFILE_DIR="${PROJECT_DIR}/profiles/${PROFILE}"
RESOURCES_DIR="${PROJECT_DIR}/resources"
LIBS_DIR="${PROJECT_DIR}/lib"
SCRIPTS_DIR="${PROJECT_DIR}/scripts"

CONFIG_FILE="${PROFILE_DIR}/application.yml"

if [ ! -f "${CONFIG_FILE}" ]; then
  CONFIG_FILE="${PROJECT_DIR}/config/application.yml"
fi
if [ ! -f "${CONFIG_FILE}" ]; then
  echo "ERROR: Not Found config file."
  exit 1
fi

APPLICATION=$(grep -v '^\s*#' "${CONFIG_FILE}" | sed '/application: /!d;s/.*: //' | tr -d '\r')
VERSION=$(grep -v '^\s*#' "${CONFIG_FILE}" | sed '/version: /!d;s/.*: //' | tr -d '\r')

if [ -z "${APPLICATION}" ]; then
  echo "ERROR: Not Found APPLICATION defined in ${CONFIG_FILE}"
  exit 1
fi
if [ -z "${VERSION}" ]; then
  echo "ERROR: Not Found VERSION defined in ${CONFIG_FILE}"
  exit 1
fi

echo "Building app name: ${APPLICATION} ${VERSION}"

TARGET="${PROJECT_DIR}/target"
rm -rf "$TARGET"

ASSEMBLY="${TARGET}/${APPLICATION}"
mkdir -p "${ASSEMBLY}" "${ASSEMBLY}/bin" "${ASSEMBLY}/config" "${ASSEMBLY}/resources" "${ASSEMBLY}/lib"

echo "Building copy files..."

if [ -d "$CONFIG_DIR" ]; then
  FILES=$(ls "${CONFIG_DIR}")
  if [ -n "${FILES}" ]; then
    cp -rfL "${CONFIG_DIR}"/* "${ASSEMBLY}/config"
  fi
fi
if [ -d "$PROFILE_DIR" ]; then
  FILES=$(ls "${PROFILE_DIR}")
  if [ -n "${FILES}" ]; then
    cp -rfL "${PROFILE_DIR}"/* "${ASSEMBLY}/config"
  fi
fi
if [ -d "$RESOURCES_DIR" ]; then
  FILES=$(ls "${RESOURCES_DIR}")
  if [ -n "${FILES}" ]; then
    cp -rfL "${RESOURCES_DIR}"/* "${ASSEMBLY}/resources"
  fi
fi
if [ -d "$LIBS_DIR" ]; then
  FILES=$(ls "${LIBS_DIR}")
  if [ -n "${FILES}" ]; then
    cp -rfL "${LIBS_DIR}"/* "${ASSEMBLY}/lib"
  fi
fi
if [ -d "$SCRIPTS_DIR" ]; then
  FILES=$(ls "${SCRIPTS_DIR}")
  if [ -n "${FILES}" ]; then
    cp -rfL "${SCRIPTS_DIR}"/* "${ASSEMBLY}/bin"
  fi
fi

echo "Building main..."

if [ -z "${GOPATH}" ]; then
  export GOPATH="${HOME}/go"
fi

if [ -n "${GOROOT}" ];then
  GO="${GOROOT}/bin/go"
else
  GO=$(which go)
fi
if [ -z "$GO" ];then
  echo "ERROR: Not Found \`go\` command defined in system."
  exit 1
fi
if ! $GO generate ; then
    echo "Build Generate failed."
fi

if ! $GO build -trimpath -ldflags="-s -w" -o "${ASSEMBLY}/bin/${APPLICATION}" . ; then
    echo "Build main failed."
    exit 1
fi

if upx -V > /dev/null 2>&1 ; then
    upx -q "${ASSEMBLY}/bin/${APPLICATION}" > /dev/null 2>&1
    echo "Building upx compress success"
fi
chmod -R a+x "${ASSEMBLY}"/bin
echo "Building zip..."
cd "${TARGET}" || exit
zip -qry "${APPLICATION}.zip" "${APPLICATION}" || exit 1

echo "Build success. Assembly is ${ASSEMBLY}.zip"

