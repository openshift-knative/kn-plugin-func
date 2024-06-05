#!/usr/bin/env bash

set -e

FUNC_S2I="pkg/pipelines/resources/tekton/task/func-s2i/0.1/func-s2i"
FUNC_BUILDPACKS="pkg/pipelines/resources/tekton/task/func-buildpacks/0.2/func-buildpacks"
FUNC_DEPLOY="pkg/pipelines/resources/tekton/task/func-deploy/0.1/func-deploy"

sed -e 's/^kind: .*/kind: Task/' "${FUNC_S2I}.yaml" > "${FUNC_S2I}-pac.yaml"
sed -e 's/^kind: .*/kind: Task/' "${FUNC_BUILDPACKS}.yaml" > "${FUNC_BUILDPACKS}-pac.yaml"
sed -e 's/^kind: .*/kind: Task/' "${FUNC_DEPLOY}.yaml" > "${FUNC_DEPLOY}-pac.yaml"


sed -r -i 's#/func-(\w+)\.yaml#/func-\1-pac.yaml#g' pkg/pipelines/tekton/templates.go
