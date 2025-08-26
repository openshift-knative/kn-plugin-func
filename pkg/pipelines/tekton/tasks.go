package tekton

import (
	"fmt"
	"strings"
)

var DeployerImage = "ghcr.io/knative/func-utils:latest"

func getBuildpackTask() string {
	return `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: func-buildpacks
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/categories: Image Build
    tekton.dev/pipelines.minVersion: "0.17.0"
    tekton.dev/tags: image-build
    tekton.dev/displayName: "Knative Functions Buildpacks"
    tekton.dev/platforms: "linux/amd64"
spec:
  description: >-
    The Knative Functions Buildpacks task builds source into a container image and pushes it to a registry,
    using Cloud Native Buildpacks. This task is based on the Buildpacks Tekton task v 0.4.

  workspaces:
    - name: source
      description: Directory where application source is located.
    - name: cache
      description: Directory where cache is stored (when no cache image is provided).
      optional: true
    - name: dockerconfig
      description: >-
        An optional workspace that allows providing a .docker/config.json file
        for Buildpacks lifecycle binary to access the container registry.
        The file should be placed at the root of the Workspace with name config.json.
      optional: true

  params:
    - name: APP_IMAGE
      description: The name of where to store the app image.
    - name: REGISTRY
      description: The registry associated with the function image.
    - name: BUILDER_IMAGE
      description: The image on which builds will run (must include lifecycle and compatible buildpacks).
    - name: SOURCE_SUBPATH
      description: A subpath within the "source" input where the source to build is located.
      default: ""
    - name: ENV_VARS
      type: array
      description: Environment variables to set during _build-time_.
      default: []
    - name: RUN_IMAGE
      description: Reference to a run image to use.
      default: ""
    - name: CACHE_IMAGE
      description: The name of the persistent app cache image (if no cache workspace is provided).
      default: ""
    - name: SKIP_RESTORE
      description: Do not write layer metadata or restore cached layers.
      default: "false"
    - name: USER_ID
      description: The user ID of the builder image user.
      default: "1001"
    - name: GROUP_ID
      description: The group ID of the builder image user.
      default: "0"
      ##############################################################
      #####  "default" has been changed to "0" for Knative Functions
    - name: PLATFORM_DIR
      description: The name of the platform directory.
      default: empty-dir

  results:
    - name: IMAGE_DIGEST
      description: The digest of the built "APP_IMAGE".

  stepTemplate:
    env:
      - name: CNB_PLATFORM_API
        value: "0.10"

  steps:
    - name: prepare
      image: docker.io/library/bash:5.1.4@sha256:b208215a4655538be652b2769d82e576bc4d0a2bb132144c060efc5be8c3f5d6
      args:
        - "--env-vars"
        - "$(params.ENV_VARS[*])"
      script: |
        #!/usr/bin/env bash
        set -e

        if [[ "$(workspaces.cache.bound)" == "true" ]]; then
          echo "> Setting permissions on '$(workspaces.cache.path)'..."
          chown -R "$(params.USER_ID):$(params.GROUP_ID)" "$(workspaces.cache.path)"
        fi

        #######################################################
        #####  "/emptyDir" has been added for Knative Functions
        for path in "/tekton/home" "/layers" "/emptyDir" "$(workspaces.source.path)"; do
          echo "> Setting permissions on '$path'..."
          chown -R "$(params.USER_ID):$(params.GROUP_ID)" "$path"

          if [[ "$path" == "$(workspaces.source.path)" ]]; then
              chmod 775 "$(workspaces.source.path)"
          fi
        done

        echo "> Parsing additional configuration..."
        parsing_flag=""
        envs=()
        for arg in "$@"; do
            if [[ "$arg" == "--env-vars" ]]; then
                echo "-> Parsing env variables..."
                parsing_flag="env-vars"
            elif [[ "$parsing_flag" == "env-vars" ]]; then
                envs+=("$arg")
            fi
        done

        echo "> Processing any environment variables..."
        ENV_DIR="/platform/env"

        echo "--> Creating 'env' directory: $ENV_DIR"
        mkdir -p "$ENV_DIR"

        for env in "${envs[@]}"; do
            IFS='=' read -r key value <<< "$env"
            if [[ "$key" != "" && "$value" != "" ]]; then
                path="${ENV_DIR}/${key}"
                echo "--> Writing ${path}..."
                echo -n "$value" > "$path"
            fi
        done

        ############################################
        ##### Added part for Knative Functions #####
        ############################################

        func_file="$(workspaces.source.path)/func.yaml"
        if [ "$(params.SOURCE_SUBPATH)" != "" ]; then
          func_file="$(workspaces.source.path)/$(params.SOURCE_SUBPATH)/func.yaml"
        fi
        echo "--> Saving 'func.yaml'"
        cp $func_file /emptyDir/func.yaml

        ############################################

      volumeMounts:
        - name: layers-dir
          mountPath: /layers
        - name: $(params.PLATFORM_DIR)
          mountPath: /platform
          ########################################################
          #####   "/emptyDir" has been added for Knative Functions
        - name: empty-dir
          mountPath: /emptyDir

    - name: create
      image: $(params.BUILDER_IMAGE)
      imagePullPolicy: Always
      command: ["/cnb/lifecycle/creator"]
      env:
        - name: DOCKER_CONFIG
          value: $(workspaces.dockerconfig.path)
      args:
        - "-app=$(workspaces.source.path)/$(params.SOURCE_SUBPATH)"
        - "-cache-dir=$(workspaces.cache.path)"
        - "-cache-image=$(params.CACHE_IMAGE)"
        - "-uid=$(params.USER_ID)"
        - "-gid=$(params.GROUP_ID)"
        - "-layers=/layers"
        - "-platform=/platform"
        - "-report=/layers/report.toml"
        - "-skip-restore=$(params.SKIP_RESTORE)"
        - "-previous-image=$(params.APP_IMAGE)"
        - "-run-image=$(params.RUN_IMAGE)"
        - "$(params.APP_IMAGE)"
      volumeMounts:
        - name: layers-dir
          mountPath: /layers
        - name: $(params.PLATFORM_DIR)
          mountPath: /platform
      securityContext:
        runAsUser: 1001
        #################################################################
        #####  "runAsGroup" has been changed to "0" for Knative Functions
        runAsGroup: 0

    - name: results
      image: docker.io/library/bash:5.1.4@sha256:b208215a4655538be652b2769d82e576bc4d0a2bb132144c060efc5be8c3f5d6
      script: |
        #!/usr/bin/env bash
        set -e
        cat /layers/report.toml | grep "digest" | cut -d'"' -f2 | cut -d'"' -f2 | tr -d '\n' | tee $(results.IMAGE_DIGEST.path)

        ############################################
        ##### Added part for Knative Functions #####
        ############################################

        digest=$(cat $(results.IMAGE_DIGEST.path))

        func_file="$(workspaces.source.path)/func.yaml"
        if [ "$(params.SOURCE_SUBPATH)" != "" ]; then
          func_file="$(workspaces.source.path)/$(params.SOURCE_SUBPATH)/func.yaml"
        fi

        if [[ ! -f "$func_file" ]]; then
          echo "--> Restoring 'func.yaml'"
          mkdir -p "$(workspaces.source.path)/$(params.SOURCE_SUBPATH)"
          cp /emptyDir/func.yaml $func_file
        fi

        echo ""
        sed -i "s|^image:.*$|image: $(params.APP_IMAGE)|" "$func_file"
        echo "Function image name: $(params.APP_IMAGE)"

        sed -i "s/^imageDigest:.*$/imageDigest: $digest/" "$func_file"
        echo "Function image digest: $digest"

        sed -i "s|^registry:.*$|registry: $(params.REGISTRY)|" "$func_file"
        echo "Function image registry: $(params.REGISTRY)"

        ############################################
      volumeMounts:
        - name: layers-dir
          mountPath: /layers
          ########################################################
          #####   "/emptyDir" has been added for Knative Functions
        - name: empty-dir
          mountPath: /emptyDir

  volumes:
    - name: empty-dir
      emptyDir: {}
    - name: layers-dir
      emptyDir: {}
`
}

func getS2ITask() string {
	return fmt.Sprintf(`apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: func-s2i
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/pipelines.minVersion: "0.17.0"
    tekton.dev/categories: Image Build
    tekton.dev/tags: image-build
    tekton.dev/platforms: "linux/amd64"
spec:
  description: >-
    Knative Functions Source-to-Image (S2I) is a toolkit and workflow for building reproducible
    container images from source code

    S2I produces images by injecting source code into a base S2I container image
    and letting the container prepare that source code for execution. The base
    S2I container images contains the language runtime and build tools needed for
    building and running the source code.

  params:
    - name: BUILDER_IMAGE
      description: The location of the s2i builder image.
    - name: IMAGE
      description: Reference of the image S2I will produce.
    - name: REGISTRY
      description: The registry associated with the function image.
      default: ""
    - name: PATH_CONTEXT
      description: The location of the path to run s2i from.
      default: .
    - name: TLSVERIFY
      description: Verify the TLS on the registry endpoint (for push/pull to a non-TLS registry)
      default: "true"
    - name: LOGLEVEL
      description: Log level when running the S2I binary
      default: "0"
    - name: ENV_VARS
      type: array
      description: Environment variables to set during _build-time_.
      default: []
    - name: S2I_IMAGE_SCRIPTS_URL
      description: The URL containing the default assemble and run scripts for the builder image.
      default: "image:///usr/libexec/s2i"
  workspaces:
    - name: source
    - name: cache
      description: Directory where cache is stored (e.g. local mvn repo).
      optional: true
    - name: sslcertdir
      optional: true
    - name: dockerconfig
      description: >-
        An optional workspace that allows providing a .docker/config.json file
        for Buildah to access the container registry.
        The file should be placed at the root of the Workspace with name config.json.
      optional: true
  results:
    - name: IMAGE_DIGEST
      description: Digest of the image just built.
  steps:
    - name: generate
      image: %s
      workingDir: $(workspaces.source.path)
      args: ["$(params.ENV_VARS[*])"]
      script: |
        echo "Processing Build Environment Variables"
        echo "" > /env-vars/env-file
        for var in "$@"
        do
            if [[ "$var" != "=" ]]; then
                echo "$var" >> /env-vars/env-file
            fi
        done

        echo "Generated Build Env Var file"
        echo "------------------------------"
        cat /env-vars/env-file
        echo "------------------------------"

        /usr/local/bin/s2i --loglevel=$(params.LOGLEVEL) build --keep-symlinks $(params.PATH_CONTEXT) $(params.BUILDER_IMAGE) \
        --image-scripts-url $(params.S2I_IMAGE_SCRIPTS_URL) \
        --as-dockerfile /gen-source/Dockerfile.gen --environment-file /env-vars/env-file

        echo "Preparing func.yaml for later deployment"
        func_file="$(workspaces.source.path)/func.yaml"
        if [ "$(params.PATH_CONTEXT)" != "" ]; then
          func_file="$(workspaces.source.path)/$(params.PATH_CONTEXT)/func.yaml"
        fi
        sed -i "s|^registry:.*$|registry: $(params.REGISTRY)|" "$func_file"
        echo "Function image registry: $(params.REGISTRY)"

        s2iignore_file="$(dirname "$func_file")/.s2iignore"
        [ -f "$s2iignore_file" ] || echo "node_modules" >> "$s2iignore_file"

      volumeMounts:
        - mountPath: /gen-source
          name: gen-source
        - mountPath: /env-vars
          name: env-vars
    - name: build
      image: registry.redhat.io/rhel8/buildah@sha256:a1e5cc0fb334e333e5eab69689223e8bd1f0c060810d260603b26cf8c0da2023
      workingDir: /gen-source
      script: |
        TLS_VERIFY_FLAG=""
        if [ "$(params.TLSVERIFY)" = "false" ] || [ "$(params.TLSVERIFY)" = "0" ]; then
          TLS_VERIFY_FLAG="--tls-verify=false"
        fi

        [[ "$(workspaces.sslcertdir.bound)" == "true" ]] && CERT_DIR_FLAG="--cert-dir $(workspaces.sslcertdir.path)"
        ARTIFACTS_CACHE_PATH="$(workspaces.cache.path)/mvn-artifacts"
        [ -d "${ARTIFACTS_CACHE_PATH}" ] || mkdir "${ARTIFACTS_CACHE_PATH}"
        buildah ${CERT_DIR_FLAG} bud --storage-driver=vfs ${TLS_VERIFY_FLAG} --layers \
          -v "${ARTIFACTS_CACHE_PATH}:/tmp/artifacts/:rw,z,U" \
          -f /gen-source/Dockerfile.gen -t $(params.IMAGE) .

        [[ "$(workspaces.dockerconfig.bound)" == "true" ]] && export DOCKER_CONFIG="$(workspaces.dockerconfig.path)"
        buildah ${CERT_DIR_FLAG} push --storage-driver=vfs ${TLS_VERIFY_FLAG} --digestfile $(workspaces.source.path)/image-digest \
          $(params.IMAGE) docker://$(params.IMAGE)

        cat $(workspaces.source.path)/image-digest | tee /tekton/results/IMAGE_DIGEST
      volumeMounts:
      - name: varlibcontainers
        mountPath: /var/lib/containers
      - mountPath: /gen-source
        name: gen-source
      securityContext:
        capabilities:
          add: ["SETFCAP"]
  volumes:
    - emptyDir: {}
      name: varlibcontainers
    - emptyDir: {}
      name: gen-source
    - emptyDir: {}
      name: env-vars
`, DeployerImage)
}

func getDeployTask() string {
	return fmt.Sprintf(`apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: func-deploy
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/pipelines.minVersion: "0.12.1"
    tekton.dev/categories: CLI
    tekton.dev/tags: cli
    tekton.dev/platforms: "linux/amd64"
spec:
  description: >-
    This Task performs a deploy operation using the Knative "func"" CLI
  params:
    - name: path
      description: Path to the function project
      default: ""
    - name: image
      description: Container image to be deployed
      default: ""
  workspaces:
    - name: source
      description: The workspace containing the function project
    - name: cache
      optional: true
    - name: sslcertdir
      optional: true
    - name: dockerconfig
      optional: true
  steps:
    - name: func-deploy
      image: "%s"
      script: |
        deploy $(params.path) "$(params.image)"
`, DeployerImage)
}

func getScaffoldTask() string {
	return fmt.Sprintf(`apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: func-scaffold
  labels:
    app.kubernetes.io/version: "0.1"
  annotations:
    tekton.dev/pipelines.minVersion: "0.12.1"
    tekton.dev/categories: CLI
    tekton.dev/tags: cli
    tekton.dev/platforms: "linux/amd64"
spec:
  params:
    - name: path
      description: Path to the function project
      default: ""
  workspaces:
    - name: source
      description: The workspace containing the function project
    - name: cache
      optional: true
    - name: sslcertdir
      optional: true
    - name: dockerconfig
      optional: true
  steps:
    - name: func-scaffold
      image: %s
      script: |
        scaffold $(params.path)
`, DeployerImage)
}

func getGitCloneTask() string {
	return `apiVersion: tekton.dev/v1
kind: Task
metadata:
  name: git-clone
  labels:
    app.kubernetes.io/version: "0.10"
  annotations:
    tekton.dev/pipelines.minVersion: "0.38.0"
    tekton.dev/categories: Git
    tekton.dev/tags: git
    tekton.dev/displayName: "git clone"
    tekton.dev/platforms: "linux/amd64,linux/s390x,linux/ppc64le,linux/arm64"
spec:
  description: >-
    These Tasks are Git tasks to work with repositories used by other tasks
    in your Pipeline.

    The git-clone Task will clone a repo from the provided url into the
    output Workspace. By default the repo will be cloned into the root of
    your Workspace. You can clone into a subdirectory by setting this Task's
    subdirectory param. This Task also supports sparse checkouts. To perform
    a sparse checkout, pass a list of comma separated directory patterns to
    this Task's sparseCheckoutDirectories param.
  workspaces:
    - name: output
      description: The git repo will be cloned onto the volume backing this Workspace.
    - name: cache
      optional: true
    - name: sslcertdir
      optional: true
    - name: dockerconfig
      optional: true
    - name: ssh-directory
      optional: true
      description: |
        A .ssh directory with private key, known_hosts, config, etc. Copied to
        the user's home before git commands are executed. Used to authenticate
        with the git remote when performing the clone. Binding a Secret to this
        Workspace is strongly recommended over other volume types.
    - name: basic-auth
      optional: true
      description: |
        A Workspace containing a .gitconfig and .git-credentials file. These
        will be copied to the user's home before any git commands are run. Any
        other files in this Workspace are ignored. It is strongly recommended
        to use ssh-directory over basic-auth whenever possible and to bind a
        Secret to this Workspace over other volume types.
    - name: ssl-ca-directory
      optional: true
      description: |
        A workspace containing CA certificates, this will be used by Git to
        verify the peer with when fetching or pushing over HTTPS.
  params:
    - name: url
      description: Repository URL to clone from.
      type: string
    - name: revision
      description: Revision to checkout. (branch, tag, sha, ref, etc...)
      type: string
      default: ""
    - name: refspec
      description: Refspec to fetch before checking out revision.
      default: ""
    - name: submodules
      description: Initialize and fetch git submodules.
      type: string
      default: "true"
    - name: depth
      description: Perform a shallow clone, fetching only the most recent N commits.
      type: string
      default: "1"
    - name: sslVerify
      description: Set the http.sslVerify global git config. Setting this to false is not advised unless you are sure that you trust your git remote.
      type: string
      default: "true"
    - name: crtFileName
      description: file name of mounted crt using ssl-ca-directory workspace. default value is ca-bundle.crt.
      type: string
      default: "ca-bundle.crt"
    - name: subdirectory
      description: Subdirectory inside the output Workspace to clone the repo into.
      type: string
      default: ""
    - name: sparseCheckoutDirectories
      description: Define the directory patterns to match or exclude when performing a sparse checkout.
      type: string
      default: ""
    - name: deleteExisting
      description: Clean out the contents of the destination directory if it already exists before cloning.
      type: string
      default: "true"
    - name: httpProxy
      description: HTTP proxy server for non-SSL requests.
      type: string
      default: ""
    - name: httpsProxy
      description: HTTPS proxy server for SSL requests.
      type: string
      default: ""
    - name: noProxy
      description: Opt out of proxying HTTP/HTTPS requests.
      type: string
      default: ""
    - name: verbose
      description: Log the commands that are executed during git-clone's operation.
      type: string
      default: "true"
    - name: gitInitImage
      description: The image providing the git-init binary that this Task runs.
      type: string
      default: "registry.redhat.io/openshift-pipelines/pipelines-git-init-rhel9@sha256:39eb71a9d59951659d27ada3789309c89ef13f89e60666fc4e4b4cf62894a4ad"
    - name: userHome
      description: |
        Absolute path to the user's home directory.
      type: string
      default: "/home/git"
  results:
    - name: commit
      description: The precise commit SHA that was fetched by this Task.
    - name: url
      description: The precise URL that was fetched by this Task.
    - name: committer-date
      description: The epoch timestamp of the commit that was fetched by this Task.
  steps:
    - name: clone
      image: "$(params.gitInitImage)"
      env:
      - name: HOME
        value: "$(params.userHome)"
      - name: PARAM_URL
        value: $(params.url)
      - name: PARAM_REVISION
        value: $(params.revision)
      - name: PARAM_REFSPEC
        value: $(params.refspec)
      - name: PARAM_SUBMODULES
        value: $(params.submodules)
      - name: PARAM_DEPTH
        value: $(params.depth)
      - name: PARAM_SSL_VERIFY
        value: $(params.sslVerify)
      - name: PARAM_CRT_FILENAME
        value: $(params.crtFileName)
      - name: PARAM_SUBDIRECTORY
        value: $(params.subdirectory)
      - name: PARAM_DELETE_EXISTING
        value: $(params.deleteExisting)
      - name: PARAM_HTTP_PROXY
        value: $(params.httpProxy)
      - name: PARAM_HTTPS_PROXY
        value: $(params.httpsProxy)
      - name: PARAM_NO_PROXY
        value: $(params.noProxy)
      - name: PARAM_VERBOSE
        value: $(params.verbose)
      - name: PARAM_SPARSE_CHECKOUT_DIRECTORIES
        value: $(params.sparseCheckoutDirectories)
      - name: PARAM_USER_HOME
        value: $(params.userHome)
      - name: WORKSPACE_OUTPUT_PATH
        value: $(workspaces.output.path)
      - name: WORKSPACE_SSH_DIRECTORY_BOUND
        value: $(workspaces.ssh-directory.bound)
      - name: WORKSPACE_SSH_DIRECTORY_PATH
        value: $(workspaces.ssh-directory.path)
      - name: WORKSPACE_BASIC_AUTH_DIRECTORY_BOUND
        value: $(workspaces.basic-auth.bound)
      - name: WORKSPACE_BASIC_AUTH_DIRECTORY_PATH
        value: $(workspaces.basic-auth.path)
      - name: WORKSPACE_SSL_CA_DIRECTORY_BOUND
        value: $(workspaces.ssl-ca-directory.bound)
      - name: WORKSPACE_SSL_CA_DIRECTORY_PATH
        value: $(workspaces.ssl-ca-directory.path)
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
      volumeMounts:
        - name: user-home
          mountPath: $(params.userHome)
      script: |
        #!/usr/bin/env sh
        set -eu

        if [ "${PARAM_VERBOSE}" = "true" ] ; then
          set -x
        fi

        if [ "${WORKSPACE_BASIC_AUTH_DIRECTORY_BOUND}" = "true" ] ; then
          cp "${WORKSPACE_BASIC_AUTH_DIRECTORY_PATH}/.git-credentials" "${PARAM_USER_HOME}/.git-credentials"
          cp "${WORKSPACE_BASIC_AUTH_DIRECTORY_PATH}/.gitconfig" "${PARAM_USER_HOME}/.gitconfig"
          chmod 400 "${PARAM_USER_HOME}/.git-credentials"
          chmod 400 "${PARAM_USER_HOME}/.gitconfig"
        fi

        if [ "${WORKSPACE_SSH_DIRECTORY_BOUND}" = "true" ] ; then
          cp -R "${WORKSPACE_SSH_DIRECTORY_PATH}" "${PARAM_USER_HOME}"/.ssh
          chmod 700 "${PARAM_USER_HOME}"/.ssh
          chmod -R 400 "${PARAM_USER_HOME}"/.ssh/*
        fi

        if [ "${WORKSPACE_SSL_CA_DIRECTORY_BOUND}" = "true" ] ; then
           export GIT_SSL_CAPATH="${WORKSPACE_SSL_CA_DIRECTORY_PATH}"
           if [ "${PARAM_CRT_FILENAME}" != "" ] ; then
              export GIT_SSL_CAINFO="${WORKSPACE_SSL_CA_DIRECTORY_PATH}/${PARAM_CRT_FILENAME}"
           fi
        fi
        CHECKOUT_DIR="${WORKSPACE_OUTPUT_PATH}/${PARAM_SUBDIRECTORY}"

        cleandir() {
          # Delete any existing contents of the repo directory if it exists.
          #
          # We don't just "rm -rf ${CHECKOUT_DIR}" because ${CHECKOUT_DIR} might be "/"
          # or the root of a mounted volume.
          if [ -d "${CHECKOUT_DIR}" ] ; then
            # Delete non-hidden files and directories
            rm -rf "${CHECKOUT_DIR:?}"/*
            # Delete files and directories starting with . but excluding ..
            rm -rf "${CHECKOUT_DIR}"/.[!.]*
            # Delete files and directories starting with .. plus any other character
            rm -rf "${CHECKOUT_DIR}"/..?*
          fi
        }

        if [ "${PARAM_DELETE_EXISTING}" = "true" ] ; then
          cleandir || true
        fi

        test -z "${PARAM_HTTP_PROXY}" || export HTTP_PROXY="${PARAM_HTTP_PROXY}"
        test -z "${PARAM_HTTPS_PROXY}" || export HTTPS_PROXY="${PARAM_HTTPS_PROXY}"
        test -z "${PARAM_NO_PROXY}" || export NO_PROXY="${PARAM_NO_PROXY}"

        git config --global --add safe.directory "${WORKSPACE_OUTPUT_PATH}"
        /ko-app/git-init \
          -url="${PARAM_URL}" \
          -revision="${PARAM_REVISION}" \
          -refspec="${PARAM_REFSPEC}" \
          -path="${CHECKOUT_DIR}" \
          -sslVerify="${PARAM_SSL_VERIFY}" \
          -submodules="${PARAM_SUBMODULES}" \
          -depth="${PARAM_DEPTH}" \
          -sparseCheckoutDirectories="${PARAM_SPARSE_CHECKOUT_DIRECTORIES}"
        cd "${CHECKOUT_DIR}"
        RESULT_SHA="$(git rev-parse HEAD)"
        EXIT_CODE="$?"
        if [ "${EXIT_CODE}" != 0 ] ; then
          exit "${EXIT_CODE}"
        fi
        RESULT_COMMITTER_DATE="$(git log -1 --pretty=%ct)"
        printf "%s" "${RESULT_COMMITTER_DATE}" > "$(results.committer-date.path)"
        printf "%s" "${RESULT_SHA}" > "$(results.commit.path)"
        printf "%s" "${PARAM_URL}" > "$(results.url.path)"
  volumes:
    - name: user-home
      emptyDir: {}
`
}

// GetClusterTasks returns multi-document yaml containing tekton tasks used by func.
func GetClusterTasks() string {
	tasks := getBuildpackTask() + "\n---\n" + getS2ITask() + "\n---\n" + getDeployTask() + "\n---\n" + getScaffoldTask()
	tasks = strings.Replace(tasks, "kind: Task", "kind: ClusterTask", -1)
	tasks = strings.ReplaceAll(tasks, "apiVersion: tekton.dev/v1", "apiVersion: tekton.dev/v1beta1")
	return tasks
}
