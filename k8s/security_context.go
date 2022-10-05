package k8s

import (
	"context"

	"github.com/Masterminds/semver"
	corev1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
)

func defaultSecurityContext(ctx context.Context, config *restclient.Config) *corev1.SecurityContext {
	runAsNonRoot := true
	return &corev1.SecurityContext{
		Privileged:               new(bool),
		AllowPrivilegeEscalation: new(bool),
		RunAsNonRoot:             &runAsNonRoot,
		Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile:           getSeccompProfile(ctx, config),
	}
}

func getSeccompProfile(ctx context.Context, config *restclient.Config) (profile *corev1.SeccompProfile) {
	defer func() {
		if r := recover(); r != nil {
			profile = nil
		}
	}()

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil
	}

	clusterOperatorResource := schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "clusteroperators",
	}

	openShiftAPIServer, err := dynClient.
		Resource(clusterOperatorResource).
		Get(ctx, "openshift-apiserver", metaV1.GetOptions{})

	if err != nil {
		return nil
	}

	serverVersion := openShiftAPIServer.Object["status"].(map[string]interface{})["versions"].([]interface{})[0].(map[string]interface{})["version"].(string)
	v, err := semver.NewVersion(serverVersion)
	if err != nil {
		return nil
	}
	if v.Compare(semver.MustParse("4.11")) >= 0 {
		return &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}
	}
	return nil
}
