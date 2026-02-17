package resourcefieldexport

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/deliveryhero/field-exporter/api/v1alpha1"
)

var (
	supportedGroupSuffixes = []string{
		"cnrm.cloud.google.com",
		"services.k8s.aws",
	}
)

func groupVersion(from v1alpha1.ResourceRef) (string, string, error) {
	fromAPIVersion := from.APIVersion

	gv, err := schema.ParseGroupVersion(fromAPIVersion)
	if err != nil {
		return "", "", err
	}

	if gv.Group == "" {
		return "", "", fmt.Errorf("apiVersion %s is invalid", fromAPIVersion)
	}

	for _, suffix := range supportedGroupSuffixes {
		if strings.HasSuffix(gv.Group, suffix) {
			return gv.Group, gv.Version, nil
		}
	}
	return "", "", fmt.Errorf("unsupported apiVersion: %s, needs to be part of %v", fromAPIVersion, supportedGroupSuffixes)
}
