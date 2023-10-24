package resourcefieldexport

import (
	"context"
	"errors"
	"github.com/deliveryhero/field-exporter/api/v1alpha1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime"
	"time"
)

func (r *Reconciler) degradedStatusWithRetry(ctx context.Context, exports *v1alpha1.ResourceFieldExport, trigger error, requeueAfter time.Duration) (controllerruntime.Result, error) {
	exports = exports.DeepCopy()
	conditions := exports.Status.Conditions
	found := -1
	for i, c := range conditions {
		if c.Type == readyCondition {
			found = i
		}
	}
	if found < 0 {
		conditions = append(conditions, v1alpha1.Condition{
			Type:   readyCondition,
			Status: v1.ConditionFalse,
		})
		found = len(conditions) - 1
	}
	conditions[found].LastTransitionTime = now()
	conditions[found].Message = pointer.String(trigger.Error())

	err := r.Status().Update(ctx, exports)
	return controllerruntime.Result{RequeueAfter: requeueAfter}, errors.Join(trigger, err)
}

func (r *Reconciler) degradedStatus(ctx context.Context, exports *v1alpha1.ResourceFieldExport, trigger error) (controllerruntime.Result, error) {
	return r.degradedStatusWithRetry(ctx, exports, trigger, 0) // 0 for time.Duration indicates the error will not be re-queued
}

func now() *metav1.Time {
	n := metav1.Now()
	return &n
}
