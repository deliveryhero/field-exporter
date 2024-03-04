package resourcefieldexport

import (
	"context"
	"errors"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	controllerruntime "sigs.k8s.io/controller-runtime"

	"github.com/deliveryhero/field-exporter/api/v1alpha1"
)

func (r *Reconciler) degradedStatus(ctx context.Context, exports *v1alpha1.ResourceFieldExport, trigger error) (controllerruntime.Result, error) {
	exports = exports.DeepCopy()
	conditions := exports.Status.Conditions
	found := -1
	updateNeeded := true
	for i, c := range conditions {
		if c.Type == readyCondition {
			found = i
		}
		if c.Status == v1.ConditionFalse && c.Message != nil && *c.Message == trigger.Error() {
			updateNeeded = false
		}
	}
	if found < 0 {
		conditions = append(conditions, v1alpha1.Condition{
			Type: readyCondition,
		})
		found = len(conditions) - 1
	}
	var err error
	if updateNeeded {
		conditions[found].LastTransitionTime = now()
		conditions[found].Message = ptr.To(trigger.Error())
		conditions[found].Status = v1.ConditionFalse
		exports.Status.Conditions = conditions
		err = r.Status().Update(ctx, exports)
	}
	return controllerruntime.Result{}, errors.Join(trigger, err)
}

func (r *Reconciler) readyStatus(ctx context.Context, exports *v1alpha1.ResourceFieldExport) (controllerruntime.Result, error) {
	exports = exports.DeepCopy()
	conditions := exports.Status.Conditions
	found := -1
	updateNeeded := true
	for i, c := range conditions {
		if c.Type == readyCondition {
			found = i
		}
		if c.Status == v1.ConditionTrue && c.Message != nil && *c.Message == "Fields Synced" {
			updateNeeded = false
		}
	}
	if found < 0 {
		conditions = append(conditions, v1alpha1.Condition{
			Type: readyCondition,
		})
		found = len(conditions) - 1
	}
	var err error
	if updateNeeded {
		conditions[found].LastTransitionTime = now()
		conditions[found].Message = ptr.To("Fields Synced")
		conditions[found].Status = v1.ConditionTrue
		exports.Status.Conditions = conditions
		err = r.Status().Update(ctx, exports)
	}
	return controllerruntime.Result{}, err
}

func now() *metav1.Time {
	n := metav1.Now()
	return &n
}
