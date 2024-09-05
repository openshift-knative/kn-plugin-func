ARG OCP_VERSION=4.16
ARG GOLANG_VERSION=1.21
FROM registry.ci.openshift.org/ocp/${OCP_VERSION}:tools as tools

FROM registry.ci.openshift.org/ocp/builder:rhel-9-golang-${GOLANG_VERSION}-openshift-${OCP_VERSION}

COPY --from=tools /usr/bin/oc /usr/bin/
RUN ln -s /usr/bin/oc /usr/bin/kubectl

# Reset the goflags to avoid the -mod=vendor flag
ENV GOFLAGS=
