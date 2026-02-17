package resourcemanager

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestRMValidate(t *testing.T) {
	registeredGVKs := []schema.GroupVersionKind{
		{Version: "v1", Kind: "Secret"},
		{Group: "redis.cnrm.cloud.google.com", Version: "v1beta1", Kind: "RedisInstance"},
		{Group: "sql.cnrm.cloud.google.com", Version: "v1beta1", Kind: "SQLInstance"},
		{Group: "rds.services.k8s.aws", Version: "v1alpha1", Kind: "DBCluster"},
	}
	for _, tc := range []struct {
		name       string
		apiVersion string
		kind       string
		expectErr  string
	}{
		{
			name:       "valid resource, invalid group",
			apiVersion: "v1",
			kind:       "Secret",
			expectErr:  "unsupported GroupVersion v1, group must have suffix in [cnrm.cloud.google.com services.k8s.aws]",
		},
		{
			name:       "invalid resource, invalid group",
			apiVersion: "v1",
			kind:       "RedisCluster",
			expectErr:  "unsupported GroupVersion v1, group must have suffix in [cnrm.cloud.google.com services.k8s.aws]",
		},
		{
			name:       "invalid resource, valid group",
			apiVersion: "sql.cnrm.cloud.google.com/v1beta1",
			kind:       "RedisInstance",
			expectErr:  "unsupported resource: sql.cnrm.cloud.google.com/v1beta1, Kind=RedisInstance",
		},
		{
			name:       "valid resource, valid group",
			apiVersion: "sql.cnrm.cloud.google.com/v1beta1",
			kind:       "SQLInstance",
		},
		{
			name:       "valid aws resource, valid group",
			apiVersion: "rds.services.k8s.aws/v1alpha1",
			kind:       "DBCluster",
		},
		{
			name:       "invalid aws resource, valid group",
			apiVersion: "rds.services.k8s.aws/v1alpha1",
			kind:       "S3Bucket",
			expectErr:  "unsupported resource: rds.services.k8s.aws/v1alpha1, Kind=S3Bucket",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rm, err := NewResourceManager(&testPreferredResources{gvks: registeredGVKs})
			require.NoError(t, err)
			err = rm.Validate(tc.apiVersion, tc.kind)
			if tc.expectErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.expectErr)
		})
	}
}

type testPreferredResources struct {
	gvks []schema.GroupVersionKind
}

func (r *testPreferredResources) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	apiResources := make([]*metav1.APIResourceList, 0, len(r.gvks))
	for _, gvk := range r.gvks {
		apiResources = append(apiResources, &metav1.APIResourceList{
			GroupVersion: gvk.GroupVersion().Identifier(),
			APIResources: []metav1.APIResource{
				{Kind: gvk.Kind},
			},
		})
	}
	return apiResources, nil
}

func TestRMResources(t *testing.T) {
	secretGVK := schema.GroupVersionKind{Version: "v1", Kind: "Secret"}
	redisGVK := schema.GroupVersionKind{Group: "redis.cnrm.cloud.google.com", Version: "v1beta1", Kind: "RedisInstance"}
	clusterGVK := schema.GroupVersionKind{Group: "redis.cnrm.cloud.google.com", Version: "v1beta1", Kind: "RedisCluster"}
	sqlGVK := schema.GroupVersionKind{Group: "sql.cnrm.cloud.google.com", Version: "v1beta1", Kind: "SQLInstance"}
	userGVK := schema.GroupVersionKind{Group: "sql.cnrm.cloud.google.com", Version: "v1beta1", Kind: "SQLUser"}
	DBClusterGVK := schema.GroupVersionKind{Group: "rds.services.k8s.aws", Version: "v1alpha1", Kind: "DBCluster"}
	for _, tc := range []struct {
		name           string
		registeredGVKs []schema.GroupVersionKind
		expectedGVKs   []schema.GroupVersionKind
	}{
		{
			name: "no config connector types registered",
			registeredGVKs: []schema.GroupVersionKind{
				secretGVK,
			},
			expectedGVKs: []schema.GroupVersionKind{},
		},
		{
			name: "one config connector type registered",
			registeredGVKs: []schema.GroupVersionKind{
				secretGVK,
				sqlGVK,
			},
			expectedGVKs: []schema.GroupVersionKind{
				sqlGVK,
			},
		},
		{
			name: "multiple types registered",
			registeredGVKs: []schema.GroupVersionKind{
				secretGVK,
				sqlGVK,
				clusterGVK,
				userGVK,
				redisGVK,
				DBClusterGVK,
			},
			expectedGVKs: []schema.GroupVersionKind{
				sqlGVK,
				clusterGVK,
				userGVK,
				redisGVK,
				DBClusterGVK,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rm, err := NewResourceManager(&testPreferredResources{gvks: tc.registeredGVKs})
			require.NoError(t, err)
			gvks := rm.Resources()
			require.ElementsMatch(t, tc.expectedGVKs, gvks)
		})
	}
}
