#!/usr/bin/env bash
#
# This script generates the productized Dockerfiles
#

set -o errexit
set -o nounset
set -o pipefail

repo_root_dir=$(dirname "$(realpath "${BASH_SOURCE[0]}")")/../..

go run github.com/openshift-knative/hack/cmd/generate@latest \
  --root-dir "${repo_root_dir}" \
  --generators dockerfile \
  --dockerfile-image-builder-fmt "registry.ci.openshift.org/openshift/release:rhel-8-release-golang-%s-openshift-4.19" \
  --includes cmd/func-util \
  --template-name "func-util"
