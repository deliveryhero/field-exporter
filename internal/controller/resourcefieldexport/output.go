package resourcefieldexport

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) writeToSecret(ctx context.Context, name string, namespace string, values map[string]string) error {
	var targetSecret v1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &targetSecret)
	if err != nil {
		// todo: disambiguate on type of error
		return err
	}
	secretCopy := targetSecret.DeepCopy()
	if secretCopy.Data == nil {
		secretCopy.Data = make(map[string][]byte)
	}
	for k, v := range values {
		secretCopy.Data[k] = []byte(v)
	}
	return r.Update(ctx, secretCopy)
}

func (r *Reconciler) writeToConfigMap(ctx context.Context, name string, namespace string, values map[string]string) error {
	var targetConfigMap v1.ConfigMap
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &targetConfigMap)
	if err != nil {
		// todo: disambiguate on type of error
		return err
	}
	cmCopy := targetConfigMap.DeepCopy()
	if cmCopy.Data == nil {
		cmCopy.Data = make(map[string]string)
	}
	for k, v := range values {
		cmCopy.Data[k] = v
	}
	return r.Update(ctx, cmCopy)
}
