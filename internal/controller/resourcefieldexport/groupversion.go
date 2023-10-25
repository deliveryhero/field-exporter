package resourcefieldexport

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/deliveryhero/field-exporter/api/v1alpha1"
)

const (
	ccGroupSuffix = "cnrm.cloud.google.com"
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

	if !strings.HasSuffix(gv.Group, ccGroupSuffix) {
		return "", "", fmt.Errorf("unsupported apiVersion: %s, needs to be part of %s", fromAPIVersion, ccGroupSuffix)
	}
	return gv.Group, gv.Version, nil
}
