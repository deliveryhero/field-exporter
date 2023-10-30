/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resourcefieldexport

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gdpv1alpha1 "github.com/deliveryhero/field-exporter/api/v1alpha1"
)

const (
	readyCondition = "Ready"
)

// Reconciler reconciles a ResourceFieldExport object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gdp.deliveryhero.io,resources=resourcefieldexports,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gdp.deliveryhero.io,resources=resourcefieldexports/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gdp.deliveryhero.io,resources=resourcefieldexports/finalizers,verbs=update
//+kubebuilder:rbac:groups=alloydb.cnrm.cloud.google.com,resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups=iam.cnrm.cloud.google.com,resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups=redis.cnrm.cloud.google.com,resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups=sql.cnrm.cloud.google.com,resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups=storage.cnrm.cloud.google.com,resources=*,verbs=get;list;watch

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	fieldExports := &gdpv1alpha1.ResourceFieldExport{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, fieldExports)

	if err != nil {
		return ctrl.Result{}, err
	}

	fromResource := fieldExports.Spec.From
	group, version, err := groupVersion(fromResource)
	if err != nil {
		return r.degradedStatus(ctx, fieldExports, err)
	}

	objectMap, err := r.resource(ctx, group, version, fromResource.Kind, fromResource.Name, req.Namespace)
	if err != nil {
		return r.degradedStatus(ctx, fieldExports, err)
	}

	if fieldExports.Spec.RequiredFields != nil {
		if err := verifyStatusConditions(ctx, objectMap, fieldExports.Spec.RequiredFields.StatusConditions); err != nil {
			return r.degradedStatus(ctx, fieldExports, err)
		}
	}

	cmValues := make(map[string]string)
	for _, export := range fieldExports.Spec.Outputs {
		value, err := fieldStringValue(ctx, objectMap, export.Path)
		if err != nil {
			return r.degradedStatus(ctx, fieldExports, err)
		}
		cmValues[export.Key] = value
	}

	switch fieldExports.Spec.To.Type {
	case gdpv1alpha1.Secret:
		err = r.writeToSecret(ctx, fieldExports.Spec.To.Name, req.Namespace, cmValues)
	case gdpv1alpha1.ConfigMap:
		err = r.writeToConfigMap(ctx, fieldExports.Spec.To.Name, req.Namespace, cmValues)
	default:
		return r.degradedStatus(ctx, fieldExports, fmt.Errorf("unsupported destination type: %s", fieldExports.Spec.To.Type))
	}

	return r.readyStatus(ctx, fieldExports)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gdpv1alpha1.ResourceFieldExport{}).
		Complete(r)
}

func (r *Reconciler) resource(ctx context.Context, group, version, kind, name, namespace string) (map[string]interface{}, error) {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   group,
		Kind:    kind,
		Version: version,
	})
	err := r.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, u)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s/%s with name %s in namespace %s", group, version, name, namespace)
	}
	return u.Object, nil
}
