#!/usr/bin/env bash

set -eox pipefail

TEMP_PACKAGE_DIR="${TEMP_PACKAGE_DIR:-}./.cr-release-packages}"
HELM_PACKAGE_BRANCH="${HELM_PACKAGE_BRANCH:-"gh-pages"}"
HELM_STABLE_BRANCH="${HELM_STABLE_BRANCH:-"main"}"
HELM_CHART_DIR_PATH="${HELM_CHART_DIR_PATH:-"deploy/helm"}"

if [[ -z "${VERSION}" ]]; then
  echo "VERSION environment variable should be set"
  exit 1
fi

if [[ -z "${GIT_EMAIL}" ]]; then
  echo "GIT_EMAIL environment variable should be set"
  exit 1
fi
if [[ -z "${GIT_USERNAME}" ]]; then
  echo "GIT_USERNAME environment variable should be set"
  exit 1
fi

rm -rf "${TEMP_PACKAGE_DIR}" || true
mkdir "${TEMP_PACKAGE_DIR}"

helm version -c

helm dependency build deploy/helm
echo "packaging deploy/helm with version: ${VERSION}"
helm package "deploy/helm" -u --version "${VERSION}" --destination "${TEMP_PACKAGE_DIR}"

git config user.email "${GIT_EMAIL}"
git config user.name "${GIT_USERNAME}"
git fetch
git switch "${HELM_PACKAGE_BRANCH}"
if [ "${CIRCLE_BRANCH}" == "${HELM_STABLE_BRANCH}" ]; then
  cp "${TEMP_PACKAGE_DIR}/*" stable/
  pushd stable
  helm repo index .
  popd
else
  cp "${TEMP_PACKAGE_DIR}/*" incubator/
  pushd incubator
  helm repo index .
  popd
fi

git add .
git commit -m "updating helm chart to version ${VERSION}"
# git push --set-upstream origin "${HELM_PACKAGE_BRANCH}" # FIXME remove before merging PR
