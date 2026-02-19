package tekton

import (
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	"sigs.k8s.io/yaml"
)

func GetDevConsolePipelines() string {
	return GetNodeJSPipeline()
}

func GetNodeJSPipeline() string {
	pipelineYaml := `apiVersion: tekton.dev/v1
kind: Pipeline
metadata:
  name: devconsole-nodejs-function-pipeline
  namespace: openshift
  labels:
    function.knative.dev: "true"
    function.knative.dev/name: viewer
    function.knative.dev/runtime: nodejs
spec:
  params:
    - description: Git repository that hosts the function project
      name: GIT_REPO
      type: string
    - description: Git revision to build
      name: GIT_REVISION
      type: string
    - description: Path where the function project is
      name: PATH_CONTEXT
      type: string
      default: .
    - description: Function image name
      name: IMAGE_NAME
      type: string
    - description: Builder image to be used
      name: BUILDER_IMAGE
      type: string
      default: image-registry.openshift-image-registry.svc:5000/openshift/nodejs:22-minimal-ubi9
    - description: Environment variables to set during build time
      name: BUILD_ENVS
      type: array
      default: []
    - default: 'image:///usr/libexec/s2i'
      description: >-
        URL containing the default assemble and run scripts for the builder
        image.
      name: s2iImageScriptsUrl
      type: string
    - description: Verify TLS when pushing to registry
      name: tlsVerify
      type: string
      default: 'true'
  workspaces:
    - description: Directory where function source is located.
      name: source-workspace
  tasks:
    - name: build
      params:
        - name: GIT_REPOSITORY
          value: $(params.GIT_REPO)
        - name: GIT_REVISION
          value: $(params.GIT_REVISION)
        - name: IMAGE
          value: $(params.IMAGE_NAME)
        - name: REGISTRY
          value: ''
        - name: LOGLEVEL
          value: 0
        - name: PATH_CONTEXT
          value: $(params.PATH_CONTEXT)
        - name: BUILDER_IMAGE
          value: $(params.BUILDER_IMAGE)
        - name: ENV_VARS
          value:
            - '$(params.BUILD_ENVS[*])'
        - name: S2I_IMAGE_SCRIPTS_URL
          value: $(params.s2iImageScriptsUrl)
        - name: TLSVERIFY
          value: $(params.tlsVerify)
      taskRef:
        params:
          - name: kind
            value: task
          - name: name
            value: func-s2i
          - name: namespace
            value: openshift-pipelines
        resolver: cluster
      workspaces:
        - name: source
          workspace: source-workspace
          subPath: "src"
        - name: cache
          workspace: source-workspace
          subPath: "cache"
`
	var pipeline v1.Pipeline
	err := yaml.UnmarshalStrict([]byte(pipelineYaml), &pipeline)
	if err != nil {
		panic(err)
	}
	var task v1.Task
	err = yaml.UnmarshalStrict([]byte(getS2ITask()), &task)
	if err != nil {
		panic(err)
	}
	pipeline.Spec.Tasks[0].TaskRef = nil
	pipeline.Spec.Tasks[0].TaskSpec = &v1.EmbeddedTask{
		TaskSpec: task.Spec,
	}
	bs, err := yaml.Marshal(pipeline)
	if err != nil {
		panic(err)
	}
	return string(bs)
}
