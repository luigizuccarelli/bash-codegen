package somenamecontroller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/org/some-operator/api/v1alpha2"
)

func (r *SomeNameReconciler) ensureNOMC(ctx context.Context, instance *v1alpha2.SomeName) (*v1alpha2.SomeNameMachineConfig, error) {
	nameSpace := types.NamespacedName{Name: instance.Name}

	desired := r.desiredNOMC(instance, nameSpace)
	if err := controllerutil.SetControllerReference(instance, desired, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set the controller reference for somenamemachineconfig %q: %w", nameSpace.Name, err)
	}

	currentNOMC, err := r.currentNOMC(ctx, nameSpace)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get somenamemachineconfig %q due to: %w", nameSpace.Name, err)
	} else if err != nil && errors.IsNotFound(err) {

		// create NOMC since it doesn't exist
		if err := r.createNOMC(ctx, desired); err != nil {
			return nil, fmt.Errorf("failed to create somenamemachineconfig %q: %w", instance.Name, err)
		}
		r.Log.V(1).Info("created somenamemachineconfig", "nomc.namespace", instance.Namespace, "nomc.name", instance.Name)
		return r.currentNOMC(ctx, nameSpace)
	}

	return r.updateNOMC(ctx, currentNOMC, desired)
}

// currentNOMC checks if the SomeNameMachineConfig exists
func (r *SomeNameReconciler) currentNOMC(ctx context.Context, nameSpace types.NamespacedName) (*v1alpha2.SomeNameMachineConfig, error) {
	mc := &v1alpha2.SomeNameMachineConfig{}
	if err := r.Get(ctx, nameSpace, mc); err != nil {
		return nil, err
	}
	return mc, nil
}

// desiredNOMC returns a SomeNameMachineConfig object
func (r *SomeNameReconciler) desiredNOMC(instance *v1alpha2.SomeName, nameSpace types.NamespacedName) *v1alpha2.SomeNameMachineConfig {
	return &v1alpha2.SomeNameMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: nameSpace.Name,
		},
		Spec: r.desiredNOMCSpec(instance),
	}
}

// createNOMC creates the SomeNameMachineConfig
func (r *SomeNameReconciler) createNOMC(ctx context.Context, instance *v1alpha2.SomeNameMachineConfig) error {
	return r.Create(ctx, instance)
}

// desiredNOMCSpec returns a SomeNameMachineConfigSpec object
func (r *SomeNameReconciler) desiredNOMCSpec(instance *v1alpha2.SomeName) v1alpha2.SomeNameMachineConfigSpec {
	s := v1alpha2.SomeNameMachineConfigSpec{}
	if instance.Spec.Type == v1alpha2.CrioKubeletSomeNameType {
		s.Debug.EnableCrioProfiling = true
	}
	if len(instance.Spec.NodeSelector) != 0 {
		s.NodeSelector = instance.Spec.NodeSelector
	}
	// TODO: ebpf, custom will go here
	return s
}

func (r *SomeNameReconciler) updateNOMC(ctx context.Context, current, desired *v1alpha2.SomeNameMachineConfig) (*v1alpha2.SomeNameMachineConfig, error) {
	updatedNOMC := current.DeepCopy()
	updated := false

	if !cmp.Equal(current.ObjectMeta.OwnerReferences, desired.ObjectMeta.OwnerReferences) {
		updatedNOMC.ObjectMeta.OwnerReferences = desired.ObjectMeta.OwnerReferences
		updated = true
	}

	if current.Spec.Debug.EnableCrioProfiling != desired.Spec.Debug.EnableCrioProfiling {
		updatedNOMC.Spec.Debug.EnableCrioProfiling = desired.Spec.Debug.EnableCrioProfiling
		updated = true
	}

	if updated {
		return updatedNOMC, r.Update(ctx, updatedNOMC)
	}

	return updatedNOMC, nil
}

func (r *SomeNameReconciler) deleteNOMC(ctx context.Context, someName *v1alpha2.SomeName) error {
	mc := &v1alpha2.SomeNameMachineConfig{}
	mc.Name = someName.Name
	if err := r.Delete(ctx, mc); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete somenamemachineconfig %s/%s: %w", mc.Namespace, mc.Name, err)
	}
	r.Log.V(1).Info("deleted somenamemachineconfig", "nomc.name", mc.Name)
	return nil
}
