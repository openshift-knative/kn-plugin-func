#!/usr/bin/env bash
#
# This script generates the productized Dockerfiles
#

set -o errexit
set -o nounset
set -o pipefail

function install_generate_hack_tool() {
  tmp_dir=$(mktemp -d)

	git clone https://github.com/openshift-knative/hack.git "${tmp_dir}"
	cd "${tmp_dir}" && \
	  go install github.com/openshift-knative/hack/cmd/generate && \
	  cd - && rm -rf "${tmp_dir}"

	return $?
}

repo_root_dir=$(dirname "$(realpath "${BASH_SOURCE[0]}")")/../..

install_generate_hack_tool || exit 1

$(go env GOPATH)/bin/generate \
  --root-dir "${repo_root_dir}" \
  --generators dockerfile \
  --dockerfile-image-builder-fmt "registry.ci.openshift.org/openshift/release:rhel-9-release-golang-%s-openshift-4.17" \
  --includes cmd/func-util \
  --additional-packages socat \
  --additional-packages tar \
  --sym-link-names /usr/local/bin/deploy \
  --sym-link-names /usr/local/bin/scaffold \
  --sym-link-names /usr/local/bin/s2i
