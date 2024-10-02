#!/usr/bin/env bash
#
# This script generates the productized Dockerfiles
#

set -o errexit
set -o nounset
set -o pipefail

function install_generate_hack_tool() {
  GOFLAGS='' go install github.com/openshift-knative/hack/cmd/generate@latest

  return $?
}

repo_root_dir=$(dirname "$(realpath "${BASH_SOURCE[0]}")")/../..

install_generate_hack_tool || exit 1

"$(go env GOPATH)"/bin/generate \
  --root-dir "${repo_root_dir}" \
  --generators dockerfile \
  --dockerfile-image-builder-fmt "registry.ci.openshift.org/openshift/release:rhel-8-release-golang-%s-openshift-4.17" \
  --includes cmd/func-util \
  --template-name "func-util"
