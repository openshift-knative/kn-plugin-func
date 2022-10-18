package k8s

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func defaultSecurityContext(client *kubernetes.Clientset) *corev1.SecurityContext {
	runAsNonRoot := true
	sc := &corev1.SecurityContext{
		Privileged:               new(bool),
		AllowPrivilegeEscalation: new(bool),
		RunAsNonRoot:             &runAsNonRoot,
		Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile:           nil,
	}
	if info, err := client.ServerVersion(); err == nil {
		var maj, min int64
		maj, err = strconv.ParseInt(info.Major, 10, 64)
		if err != nil {
			maj = -1
		}
		min, err = strconv.ParseInt(info.Minor, 10, 64)
		if err != nil {
			min = -1
		}
		if (maj >= 2) || (maj == 1 && min >= 24) {
			sc.SeccompProfile = &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}
		}
	}
	return sc
}
