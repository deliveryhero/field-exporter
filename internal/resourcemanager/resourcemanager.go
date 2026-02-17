package resourcemanager

import (
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	supportedAPIGroups = map[string]struct{}{
		"alloydb.cnrm.cloud.google.com": {},
		"iam.cnrm.cloud.google.com":     {},
		"redis.cnrm.cloud.google.com":   {},
		"sql.cnrm.cloud.google.com":     {},
		"storage.cnrm.cloud.google.com": {},
		"rds.services.k8s.aws":          {},
		"elasticache.services.k8s.aws":  {},
		"dynamodb.services.k8s.aws":     {},
	}
)

type PreferredResources interface {
	ServerPreferredResources() ([]*metav1.APIResourceList, error)
}

func NewResourceManager(client PreferredResources) (*ResourceManager, error) {
	resources, err := supportedResources(client)
	if err != nil {
		return nil, err
	}
	return &ResourceManager{supportedResources: resources}, nil
}

type ResourceManager struct {
	supportedResources map[schema.GroupVersionKind]struct{}
}

func (r *ResourceManager) Validate(apiVersion, kind string) error {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return err
	}

	if _, ok := supportedAPIGroups[gv.Group]; !ok {
		supportedSuffixes := []string{"cnrm.cloud.google.com", "services.k8s.aws"}
		sort.Strings(supportedSuffixes)
		return fmt.Errorf("unsupported GroupVersion %s, group must have suffix in %v", gv, supportedSuffixes)
	}

	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}

	if _, ok := r.supportedResources[gvk]; !ok {
		return fmt.Errorf("unsupported resource: %s", gvk)
	}
	return nil
}

func supportedResources(client PreferredResources) (map[schema.GroupVersionKind]struct{}, error) {
	preferredResources, err := client.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	output := make(map[schema.GroupVersionKind]struct{})
	for _, resourceList := range preferredResources {
		gv, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			return nil, err
		}
		if _, ok := supportedAPIGroups[gv.Group]; !ok {
			continue
		}
		for _, r := range resourceList.APIResources {
			output[schema.GroupVersionKind{
				Group:   gv.Group,
				Version: gv.Version,
				Kind:    r.Kind,
			}] = struct{}{}
		}
	}
	return output, nil
}

func (r *ResourceManager) Resources() []schema.GroupVersionKind {
	output := make([]schema.GroupVersionKind, 0, len(r.supportedResources))
	for gvk := range r.supportedResources {
		output = append(output, gvk)
	}
	return output
}
