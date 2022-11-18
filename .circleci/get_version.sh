#!/bin/bash
set -eo pipefail

if [[ "${CIRCLE_TAG}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "${CIRCLE_TAG}"
    exit 0
fi

# shellcheck disable=SC2068,SC2046
LAST_RELEASE="$(git describe --always --tags $(git rev-list --tags) | grep -E '^[0-9]+\.[0-9]+\.[0-9]+$' | head -n 1)"
NEXT_RELEASE="$(echo "${LAST_RELEASE}" | awk -F. '/[0-9]+\./{$NF++;print}' OFS=.)"
if [[ "${CIRCLE_BRANCH}" == "main" ]]; then
    echo "${NEXT_RELEASE}-alpha.${CIRCLE_BUILD_NUM}"
    exit 0
fi

echo "${NEXT_RELEASE}-branch.${BRANCH_NAME_TRUNCATED}.${CIRCLE_BUILD_NUM}"
exit 0
