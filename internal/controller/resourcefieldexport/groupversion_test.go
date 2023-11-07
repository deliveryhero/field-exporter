package resourcefieldexport

import (
	"testing"

	gdpv1alpha1 "github.com/deliveryhero/field-exporter/api/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestGroupVersion(t *testing.T) {
	for _, tc := range []struct {
		name          string
		input         gdpv1alpha1.ResourceRef
		expectGroup   string
		expectVersion string
		expectErr     string
	}{
		{
			name: "bucket",
			input: gdpv1alpha1.ResourceRef{
				APIVersion: "storage.cnrm.cloud.google.com/v1alpha1",
				Kind:       "Bucket",
			},
			expectGroup:   "storage.cnrm.cloud.google.com",
			expectVersion: "v1alpha1",
		},
		{
			name: "bucket malformed",
			input: gdpv1alpha1.ResourceRef{
				APIVersion: "storage.cnrm.cloud.google.com",
			},
			expectErr: "apiVersion storage.cnrm.cloud.google.com is invalid",
		},
		{
			name:      "unsupported resource",
			input:     gdpv1alpha1.ResourceRef{APIVersion: "dbcluster.rds.services.k8s.aws/v1alpha2"},
			expectErr: "unsupported apiVersion: dbcluster.rds.services.k8s.aws/v1alpha2, needs to be part of cnrm.cloud.google.com",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			group, version, err := groupVersion(tc.input)
			if tc.expectErr != "" {
				require.ErrorContains(t, err, tc.expectErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectGroup, group)
			require.Equal(t, tc.expectVersion, version)
		})
	}
}
