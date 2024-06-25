#!/usr/bin/env bash

set -e

FUNC_S2I="pkg/pipelines/resources/tekton/task/func-s2i/0.1/func-s2i"
FUNC_BUILDPACKS="pkg/pipelines/resources/tekton/task/func-buildpacks/0.2/func-buildpacks"
FUNC_DEPLOY="pkg/pipelines/resources/tekton/task/func-deploy/0.1/func-deploy"

for fname in "${FUNC_S2I}" "${FUNC_BUILDPACKS}" "${FUNC_DEPLOY}"; do
    cp "${fname}.yaml" "${fname}-pac.yaml"
    sed -i 's/^kind: Task$/kind: ClusterTask/g' "${fname}.yaml" 
done

sed -r -i 's#/func-(\w+)\.yaml#/func-\1-pac.yaml#g' pkg/pipelines/tekton/templates.go
