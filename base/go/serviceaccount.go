package somenamecontroller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	v1alpha2 "github.com/org/some-operator/api/v1alpha2"
)

const (
	serviceAccountName = "some-name"
)

// ensureServiceAccount ensures that the serviceaccount exists
// Returns a Boolean value indicating whether it exists, a pointer to the
// serviceaccount and an error when relevant
func (r *SomeNameReconciler) ensureServiceAccount(ctx context.Context, someName *v1alpha2.SomeName, ns string) (*corev1.ServiceAccount, error) {
	nameSpace := types.NamespacedName{Namespace: ns, Name: serviceAccountName}

	desired := r.desiredServiceAccount(someName, ns)
	if err := controllerutil.SetControllerReference(someName, desired, r.Scheme); err != nil {
		return nil, fmt.Errorf("failed to set the controller reference for serviceaccount %q: %w", nameSpace, err)
	}

	current, err := r.currentServiceAccount(ctx, nameSpace)
	if err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get serviceaccount %q due to: %w", nameSpace, err)
	} else if err != nil && errors.IsNotFound(err) {

		// creating serviceaccount since it is not found
		if err := r.createServiceAccount(ctx, desired); err != nil {
			return nil, fmt.Errorf("failed to create serviceaccount %q: %w", nameSpace, err)
		}
		r.Log.V(1).Info("successfully created serviceaccount", "sa.name", nameSpace.Name, "sa.namespace", nameSpace.Namespace)
		return r.currentServiceAccount(ctx, nameSpace)
	}

	// update serviceaccount since it already exists
	updated, err := r.updateServiceAccount(ctx, current, desired)
	if err != nil {
		return nil, fmt.Errorf("failed to update serviceaccount %q: %w", nameSpace, err)
	}

	if updated {
		current, err = r.currentServiceAccount(ctx, nameSpace)
		if err != nil {
			return nil, fmt.Errorf("failed to get existing serviceaccount %q: %w", nameSpace, err)
		}
		r.Log.V(1).Info("successfully updated serviceaccount", "sa.name", nameSpace.Name, "sa.namespace", nameSpace.Namespace)
	}
	return current, nil
}

// currentServiceAccount checks that the serviceaccount exists
func (r *SomeNameReconciler) currentServiceAccount(ctx context.Context, nameSpace types.NamespacedName) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	if err := r.Get(ctx, nameSpace, sa); err != nil {
		return nil, err
	}
	return sa, nil
}

// createServiceAccount creates the serviceaccount
func (r *SomeNameReconciler) createServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
	return r.Create(ctx, sa)
}

// desiredServiceAccount returns a serviceaccount object
func (r *SomeNameReconciler) desiredServiceAccount(someName *v1alpha2.SomeName, ns string) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      serviceAccountName,
		},
	}
}

func (r *SomeNameReconciler) updateServiceAccount(ctx context.Context, current, desired *corev1.ServiceAccount) (bool, error) {
	updatedSA := current.DeepCopy()
	var updated bool

	if !cmp.Equal(current.ObjectMeta.OwnerReferences, desired.ObjectMeta.OwnerReferences) {
		updatedSA.ObjectMeta.OwnerReferences = desired.ObjectMeta.OwnerReferences
		updated = true
	}

	if updated {
		if err := r.Update(ctx, updatedSA); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}
