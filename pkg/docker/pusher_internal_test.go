//go:build !integration
// +build !integration

package docker

import "testing"

func TestIsInClusterRegistry(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		want     bool
	}{
		{
			name:     "openshift internal registry with port",
			registry: "image-registry.openshift-image-registry.svc:5000",
			want:     true,
		},
		{
			name:     "svc without port",
			registry: "my-registry.my-ns.svc",
			want:     true,
		},
		{
			name:     "fully qualified cluster.local with port",
			registry: "my-registry.my-ns.svc.cluster.local:5000",
			want:     true,
		},
		{
			name:     "external registry",
			registry: "quay.io",
			want:     false,
		},
		{
			name:     "external registry with port",
			registry: "localhost:5000",
			want:     false,
		},
		{
			name:     "host that merely contains svc is not in-cluster",
			registry: "svc.example.com",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isInClusterRegistry(tt.registry); got != tt.want {
				t.Errorf("isInClusterRegistry(%q) = %v, want %v", tt.registry, got, tt.want)
			}
		})
	}
}
