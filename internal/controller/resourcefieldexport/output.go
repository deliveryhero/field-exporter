package resourcefieldexport

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *Reconciler) writeToSecret(ctx context.Context, name string, namespace string, values map[string]string) error {
	logger := log.FromContext(ctx)
	var targetSecret v1.Secret
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &targetSecret)
	if err != nil {
		// todo: disambiguate on type of error
		logger.Error(err, "failed to get target Secret",
			"name", name,
			"namespace", namespace)
		return err
	}
	secretCopy := targetSecret.DeepCopy()
	if secretCopy.Data == nil {
		secretCopy.Data = make(map[string][]byte)
	}
	for k, v := range values {
		secretCopy.Data[k] = []byte(v)
	}
	err = r.Update(ctx, secretCopy)
	if err != nil {
		logger.Error(err, "failed to update Secret",
			"name", name,
			"namespace", namespace,
			"keyCount", len(values))
		return err
	}
	logger.Info("successfully updated Secret",
		"name", name,
		"namespace", namespace,
		"keyCount", len(values))
	return nil
}

func (r *Reconciler) writeToConfigMap(ctx context.Context, name string, namespace string, values map[string]string) error {
	logger := log.FromContext(ctx)
	var targetConfigMap v1.ConfigMap
	err := r.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &targetConfigMap)
	if err != nil {
		// todo: disambiguate on type of error
		logger.Error(err, "failed to get target ConfigMap",
			"name", name,
			"namespace", namespace)
		return err
	}
	cmCopy := targetConfigMap.DeepCopy()
	if cmCopy.Data == nil {
		cmCopy.Data = make(map[string]string)
	}
	for k, v := range values {
		cmCopy.Data[k] = v
	}
	err = r.Update(ctx, cmCopy)
	if err != nil {
		logger.Error(err, "failed to update ConfigMap",
			"name", name,
			"namespace", namespace,
			"keyCount", len(values))
		return err
	}
	logger.Info("successfully updated ConfigMap",
		"name", name,
		"namespace", namespace,
		"keyCount", len(values))
	return nil
}
