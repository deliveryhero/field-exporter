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

package v1alpha1

import (
	"errors"
	"fmt"

	"github.com/itchyny/gojq"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/deliveryhero/field-exporter/internal/resourcemanager"
)

// log is for logging in this package.
var (
	resourcefieldexportlog = logf.Log.WithName("resourcefieldexport-resource")
	resourceValidator      *resourcemanager.ResourceManager
)

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *ResourceFieldExport) SetupWebhookWithManager(mgr ctrl.Manager, validator *resourcemanager.ResourceManager) error {
	resourceValidator = validator
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-gdp-deliveryhero-io-v1alpha1-resourcefieldexport,mutating=false,failurePolicy=fail,sideEffects=None,groups=gdp.deliveryhero.io,resources=resourcefieldexports,verbs=create;update,versions=v1alpha1,name=vresourcefieldexport.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ResourceFieldExport{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ResourceFieldExport) ValidateCreate() (admission.Warnings, error) {
	resourcefieldexportlog.Info("validate create", "name", r.Name, "namespace", r.Namespace)
	return r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ResourceFieldExport) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	resourcefieldexportlog.Info("validate update", "name", r.Name, "namespace", r.Namespace)
	return r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ResourceFieldExport) ValidateDelete() (admission.Warnings, error) {
	resourcefieldexportlog.Info("validate delete", "name", r.Name, "namespace", r.Namespace)
	return nil, nil
}

func (r *ResourceFieldExport) validate() (admission.Warnings, error) {
	var errs []error
	errs = append(errs, resourceValidator.Validate(r.Spec.From.APIVersion, r.Spec.From.Kind))
	for _, o := range r.Spec.Outputs {
		_, err := gojq.Parse(o.Key)
		if err != nil {
			errs = append(errs, fmt.Errorf("output key %s is invalid: %w", o.Key, err))
		}
	}
	return nil, errors.Join(errs...)
}
