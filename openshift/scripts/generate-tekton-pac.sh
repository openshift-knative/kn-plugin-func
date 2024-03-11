#!/usr/bin/env bash

set -e

FUNC_S2I="pkg/pipelines/resources/tekton/task/func-s2i/0.1/func-s2i"
FUNC_BUILDPACKS="pkg/pipelines/resources/tekton/task/func-buildpacks/0.2/func-buildpacks"
FUNC_DEPLOY="pkg/pipelines/resources/tekton/task/func-deploy/0.1/func-deploy"

sed -e 's/^kind: .*/kind: Task/' "${FUNC_S2I}.yaml" > "${FUNC_S2I}-pac.yaml"
sed -e 's/^kind: .*/kind: Task/' "${FUNC_BUILDPACKS}.yaml" > "${FUNC_BUILDPACKS}-pac.yaml"
sed -e 's/^kind: .*/kind: Task/' "${FUNC_DEPLOY}.yaml" > "${FUNC_DEPLOY}-pac.yaml"

cat << EOF | git apply -
diff --git a/pkg/pipelines/tekton/templates.go b/pkg/pipelines/tekton/templates.go
index 6a92e62f..379d55e1 100644
--- a/pkg/pipelines/tekton/templates.go
+++ b/pkg/pipelines/tekton/templates.go
@@ -93,9 +93,9 @@ var (

 	taskBasePath = "https://raw.githubusercontent.com/" +
 		FuncRepoRef + "/" + FuncRepoBranchRef + "/pkg/pipelines/resources/tekton/task/"
-	BuildpackTaskURL = taskBasePath + "func-buildpacks/0.2/func-buildpacks.yaml"
-	S2ITaskURL       = taskBasePath + "func-s2i/0.1/func-s2i.yaml"
-	DeployTaskURL    = taskBasePath + "func-deploy/0.1/func-deploy.yaml"
+	BuildpackTaskURL = taskBasePath + "func-buildpacks/0.2/func-buildpacks-pac.yaml"
+	S2ITaskURL       = taskBasePath + "func-s2i/0.1/func-s2i-pac.yaml"
+	DeployTaskURL    = taskBasePath + "func-deploy/0.1/func-deploy-pac.yaml"
 )

 type templateData struct {
EOF
