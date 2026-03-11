#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
STACK_NAME="${STACK_NAME:-crackoa}"
IMAGE_TAG="${IMAGE_TAG:-swarm-local}"
IMAGE_PREFIX="${IMAGE_PREFIX:-crackoa}"
API_IMAGE="${API_IMAGE:-${IMAGE_PREFIX}/api:${IMAGE_TAG}}"
WEB_IMAGE="${WEB_IMAGE:-${IMAGE_PREFIX}/web:${IMAGE_TAG}}"
API_ENV_FILE="${API_ENV_FILE:-${ROOT_DIR}/api/.env}"
STACK_FILE="${ROOT_DIR}/deploy/swarm/stack.yml"
PUSH_IMAGES="${PUSH_IMAGES:-0}"

required_bins=(docker grep cut tr)
for bin in "${required_bins[@]}"; do
    if ! command -v "${bin}" >/dev/null 2>&1; then
        echo "missing required command: ${bin}" >&2
        exit 1
    fi
done

read_env_value() {
    local key="$1"
    local file="$2"
    local line

    line="$(grep -E "^[[:space:]]*${key}=" "${file}" | tail -n 1 || true)"
    line="${line#*=}"
    line="$(printf '%s' "${line}" | tr -d '\r')"

    if [[ "${line}" == \"*\" && "${line}" == *\" ]]; then
        line="${line:1:${#line}-2}"
    elif [[ "${line}" == \'*\' && "${line}" == *\' ]]; then
        line="${line:1:${#line}-2}"
    fi

    printf '%s' "${line}"
}

ensure_secret() {
    local name="$1"
    local value="$2"

    if docker secret inspect "${name}" >/dev/null 2>&1; then
        docker secret rm "${name}" >/dev/null
    fi

    if [[ -z "${value}" ]]; then
        printf '\n' | docker secret create "${name}" - >/dev/null
        return 0
    fi

    printf '%s' "${value}" | docker secret create "${name}" - >/dev/null
}

wait_for_stack_removal() {
    local stack="$1"
    local attempt

    for attempt in {1..30}; do
        if [[ -z "$(docker stack services "${stack}" --format '{{.Name}}' 2>/dev/null || true)" ]]; then
            return 0
        fi
        sleep 2
    done

    echo "timed out waiting for stack ${stack} to stop" >&2
    exit 1
}

if [[ ! -f "${API_ENV_FILE}" ]]; then
    echo "missing API env file: ${API_ENV_FILE}" >&2
    echo "create it from api/.env.example before deploying" >&2
    exit 1
fi

access_key="$(read_env_value ACCESS_KEY "${API_ENV_FILE}")"
jwt_secret="$(read_env_value JWT_SECRET "${API_ENV_FILE}")"
openai_api_key="$(read_env_value OPENAI_API_KEY "${API_ENV_FILE}")"
gemini_api_key="$(read_env_value GEMINI_API_KEY "${API_ENV_FILE}")"

if [[ -z "${access_key}" ]]; then
    echo "ACCESS_KEY must be set in ${API_ENV_FILE}" >&2
    exit 1
fi

if [[ -z "${jwt_secret}" ]]; then
    echo "JWT_SECRET must be set in ${API_ENV_FILE}" >&2
    exit 1
fi

swarm_state="$(docker info --format '{{.Swarm.LocalNodeState}}')"
if [[ "${swarm_state}" == "inactive" ]]; then
    docker swarm init >/dev/null
fi

if docker stack ls --format '{{.Name}}' | grep -Fx "${STACK_NAME}" >/dev/null 2>&1; then
    docker stack rm "${STACK_NAME}" >/dev/null
    wait_for_stack_removal "${STACK_NAME}"
fi

docker build -t "${API_IMAGE}" -f "${ROOT_DIR}/api/Dockerfile" "${ROOT_DIR}/api"
docker build -t "${WEB_IMAGE}" -f "${ROOT_DIR}/web/Dockerfile" "${ROOT_DIR}/web"

if [[ "${PUSH_IMAGES}" == "1" ]]; then
    docker push "${API_IMAGE}"
    docker push "${WEB_IMAGE}"
fi

ensure_secret crackoa_access_key "${access_key}"
ensure_secret crackoa_jwt_secret "${jwt_secret}"
ensure_secret crackoa_openai_api_key "${openai_api_key}"
ensure_secret crackoa_gemini_api_key "${gemini_api_key}"

export API_IMAGE WEB_IMAGE
docker stack deploy --compose-file "${STACK_FILE}" "${STACK_NAME}" >/dev/null

echo "stack ${STACK_NAME} deployed"
echo "images:"
echo "  api: ${API_IMAGE}"
echo "  web: ${WEB_IMAGE}"
echo "services:"
docker stack services "${STACK_NAME}"
