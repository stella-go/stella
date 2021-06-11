#!/bin/bash

#
# Usage: build.sh [dev|beta|product]
#

echo "Building..."

PROJECT_DIR=$(dirname "$0")
cd "${PROJECT_DIR}" || exit 1
PROJECT_DIR=$(pwd)
echo "Project dir: ${PROJECT_DIR}"
PROFILE="product"
if [ -n "$1" ]; then
  PROFILE=$1
fi
echo "Building profile: ${PROFILE}"

PROFILE_DIR="${PROJECT_DIR}/profiles/${PROFILE}"
RESOURCES_DIR="${PROJECT_DIR}/resources"
LIBS_DIR="${PROJECT_DIR}/lib"
SCRIPTS_DIR="${PROJECT_DIR}/scripts"

CONFIG_FILE="${PROFILE_DIR}/application.yml"

if [ ! -f "${CONFIG_FILE}" ]; then
  CONFIG_FILE="${PROJECT_DIR}/config/application.yml"
fi
if [ ! -f "${CONFIG_FILE}" ]; then
  CONFIG_FILE="${PROJECT_DIR}/application.yml"
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

echo "Building AppName: ${APPLICATION} ${VERSION}"

TARGET="${PROJECT_DIR}/target"
rm -rf "$TARGET"

ASSEMBLY="${TARGET}/${APPLICATION}"
mkdir -p "${ASSEMBLY}" "${ASSEMBLY}"/bin "${ASSEMBLY}"/config "${ASSEMBLY}"/resources "${ASSEMBLY}"/lib

echo "Building copy files..."
if [ -d "$RESOURCES_DIR" ]; then
  FILES=$(ls "${RESOURCES_DIR}")
  if [ -n "${FILES}" ]; then
    cp "${RESOURCES_DIR}"/* "${ASSEMBLY}"/resources
  fi
fi
if [ -d "$SCRIPTS_DIR" ]; then
  FILES=$(ls "${SCRIPTS_DIR}")
  if [ -n "${FILES}" ]; then
    cp "${SCRIPTS_DIR}"/* "${ASSEMBLY}"/bin
  fi
fi
if [ -d "$PROFILE_DIR" ]; then
  FILES=$(ls "${PROFILE_DIR}")
  if [ -n "${FILES}" ]; then
    cp "${PROFILE_DIR}"/* "${ASSEMBLY}"/config
  fi
fi
if [ -d "$LIBS_DIR" ]; then
  FILES=$(ls "${LIBS_DIR}")
  if [ -n "${FILES}" ]; then
    cp "${LIBS_DIR}"/* "${ASSEMBLY}"/lib
  fi
fi

echo "Building main..."

if [ -z "${GOPATH}" ]; then
  export GOPATH="${HOME}/go"
fi

if [ -n "${GOROOT}" ];then
  GO=${GOROOT}/bin/go
else
  GO=$(which go)
fi
if [ -z "$GO" ];then
  echo "ERROR: Not Found \`go\` command defined in system."
  exit 1
fi
if ! $GO build -o "${ASSEMBLY}"/bin/"${APPLICATION}" . ; then
    echo "Build main failed."
    exit 1
fi

chmod -R a+x "${ASSEMBLY}"/bin
echo "Building zip..."
cd "${TARGET}" || exit
zip -qry "${ASSEMBLY}".zip "${ASSEMBLY}" || exit 1

echo "Build success. Assembly is ${ASSEMBLY}.zip"

