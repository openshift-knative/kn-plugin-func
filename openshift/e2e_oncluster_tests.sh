#!/usr/bin/env bash
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Runs func remote test using 'func' binary built from source
#

set -o errexit
set -o nounset
set -o pipefail

source "$(go run knative.dev/hack/cmd/script e2e-tests.sh)"

pushd "$(dirname "$0")/.."

export BUILD_NUMBER=${BUILD_NUMBER:-$(head -c 128 < /dev/urandom | LC_CTYPE=C tr -dc 'a-z0-9' | head -c 8)}
export ARTIFACT_DIR="${ARTIFACT_DIR:-$(dirname "$(mktemp -d -u)")/build-${BUILD_NUMBER}}"
export ARTIFACTS="${ARTIFACTS:-${ARTIFACT_DIR}}/kn-func/e2e-oncluster-tests"
mkdir -p "${ARTIFACTS}"

# Prepare Build Parameters
FUNC_UTILS_IMG="${FUNC_UTILS_IMG:-ghcr.io/knative/func-utils:latest}"
LDFLAGS="-X knative.dev/func/pkg/k8s.SocatImage=${FUNC_UTILS_IMG}"
LDFLAGS+=" -X knative.dev/func/pkg/k8s.TarImage=${FUNC_UTILS_IMG}"
LDFLAGS+=" -X knative.dev/func/pkg/pipelines/tekton.DeployerImage=${FUNC_UTILS_IMG}"
export GOFLAGS="'-ldflags=${LDFLAGS}'"
FUNC_REPO_REF="${FUNC_REPO_REF:-openshift-knative/kn-plugin-func}"
FUNC_REPO_BRANCH_REF="${FUNC_REPO_BRANCH_REF:-release-next}"

# Build 'func' binary
echo "=== building func binary"
env FUNC_REPO_REF="${FUNC_REPO_REF}" FUNC_REPO_BRANCH_REF="${FUNC_REPO_BRANCH_REF}" make build

# Patch Tests for Openshift
sed -i 's|http://%v.%s.%s|https://%v-%s.%s|' ./e2e/e2e_*.go
sed -i 's|"--builder", "pack"|"--builder", "s2i"|' ./e2e/e2e_*.go
sed -i 's|--builder=pack|--builder=s2i|' ./e2e/e2e_*.go

# Execute on Remote tests
echo "=== running func e2e Remote tests"

export FUNC_E2E_NAMESPACE="$(oc project -q)"
export FUNC_E2E_DOMAIN="$(oc get ingresses.config cluster -o jsonpath='{.spec.domain}')"
export FUNC_E2E_BIN="$(pwd)/func"
export FUNC_E2E_CLUSTER_REGISTRY="image-registry.openshift-image-registry.svc:5000/${FUNC_E2E_NAMESPACE}"
export FUNC_E2E_KUBECONFIG="$KUBECONFIG"

go_test_e2e -v -timeout 15m -tags="e2e" -run TestRemote_Deploy ./e2e || fail_test 'kn-func e2e tests'
ret=$?

popd
exit $ret
