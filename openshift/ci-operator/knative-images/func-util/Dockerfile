# DO NOT EDIT! Generated Dockerfile for cmd/func-util.
ARG GO_BUILDER=registry.ci.openshift.org/openshift/release:rhel-8-release-golang-1.23-openshift-4.19
ARG GO_RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal

FROM $GO_BUILDER as builder

WORKDIR /workspace
COPY . .

ENV CGO_ENABLED=1
ENV GOEXPERIMENT=strictfipsruntime
ENV GOFLAGS=''

RUN go build -tags strictfipsruntime,exclude_graphdriver_btrfs -o /usr/bin/main ./cmd/func-util

FROM $GO_RUNTIME

ARG VERSION=knative-v1.17

RUN microdnf install -y socat tar

COPY --from=builder /usr/bin/main /usr/bin/func-util
COPY LICENSE /licenses/

RUN ln -s /usr/bin/func-util /usr/local/bin/deploy && \
	ln -s /usr/bin/func-util /usr/local/bin/scaffold && \
	ln -s /usr/bin/func-util /usr/local/bin/s2i && \
	ln -s /usr/bin/func-util /usr/local/bin/s2i-generate

USER 65532

LABEL \
      com.redhat.component="openshift-serverless-1-kn-plugin-func-func-util-rhel9-container" \
      name="openshift-serverless-1/kn-plugin-func-func-util-rhel9" \
      version=$VERSION \
      summary="Red Hat OpenShift Serverless 1 Kn Plugin Func Func Util" \
      maintainer="serverless-support@redhat.com" \
      description="Red Hat OpenShift Serverless 1 Kn Plugin Func Func Util" \
      io.k8s.display-name="Red Hat OpenShift Serverless 1 Kn Plugin Func Func Util" \
      io.k8s.description="Red Hat OpenShift Serverless Kn Plugin Func Func Util" \
      io.openshift.tags="func-util" \
      cpe="cpe:/a:redhat:openshift_serverless:1.37::el9"

ENTRYPOINT ["/usr/bin/bash"]
