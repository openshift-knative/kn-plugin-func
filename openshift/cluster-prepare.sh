#!/usr/bin/env bash
#
# Prepare Cluster on Openshift CI
# - Creates testing Namespace
# - Setup Openshift Serverless and Openshift Pipelines

set -o errexit
set -o nounset
set -o pipefail

BASEDIR=$(dirname "$0")
INSTALL_SERVERLESS="${INSTALL_SERVERLESS:-true}"
INSTALL_PIPELINES="${INSTALL_PIPELINES:-true}"

go env
source "$(go run knative.dev/hack/cmd/script e2e-tests.sh)"

# Prepare Namespace
TEST_NAMESPACE="${TEST_NAMESPACE:-knfunc-oncluster-test-$(head -c 128 </dev/urandom | LC_CTYPE=C tr -dc 'a-z0-9' | fold -w 6 | head -n 1)}"
oc new-project "${TEST_NAMESPACE}" || true
oc project "${TEST_NAMESPACE}"

# Installs Openshift Serverless
if [ "$INSTALL_SERVERLESS" == "true" ] ; then
  header "Installing Openshift Serverless"
  oc apply -f "${BASEDIR}/deploy/serverless-subscription.yaml"
  wait_until_pods_running openshift-serverless

  subheader "Installing Serving and Eventing"
  oc apply -f "${BASEDIR}/deploy/knative-serving.yaml"
  oc apply -f "${BASEDIR}/deploy/knative-eventing.yaml"
  oc wait --for=condition=Ready --timeout=10m knativeserving knative-serving -n knative-serving
  oc wait --for=condition=Ready --timeout=10m knativeeventing knative-eventing -n knative-eventing
fi

# Installs Openshift Pipelines
if [ "$INSTALL_PIPELINES" == "true" ] ; then
  header "Installing Openshift Pipelines"
  oc apply -f "${BASEDIR}/deploy/pipelines-subscription.yaml"
  wait_until_pods_running openshift-pipelines
fi

 # TEMPORARY WORKAROUND: Disable affinity assistant to prevent pod scheduling issues
subheader "Disabling affinity assistant (temporary workaround)"
oc patch configmap feature-flags -n openshift-pipelines \
  -p '{"data":{"disable-affinity-assistant":"true", "coschedule":"disabled"}}' \
  --type=merge

# Patch domain template to match tests check
oc patch -n knative-serving cm/config-network --patch '{"data":{"domain-template":"{{.Name}}-{{.Namespace}}-ksvc.{{.Domain}}"}}'

