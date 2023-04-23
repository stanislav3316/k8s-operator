/*
Copyright 2023 StanislavBolsun.

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

package controllers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	homeworkv1alpha1 "github.com/stanislav3316/DummyOperator/api/v1alpha1"
)

// DummyReconciler reconciles a Dummy object
type DummyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=homework.anynines.com,resources=dummies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homework.anynines.com,resources=pods,namespace=d-op-system,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=homework.anynines.com,resources=dummies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=homework.anynines.com,resources=dummies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Dummy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *DummyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	dummy := &homeworkv1alpha1.Dummy{}
	err := r.Client.Get(ctx, req.NamespacedName, dummy)

	if err != nil {
		logger.Error(err, "failed to get Dummy resource")
		return ctrl.Result{}, err
	}

	logger.Info("Dummy instance:", "Got name - ", dummy.Name, " namespace - ", req.Namespace, "msg - ", dummy.Spec.Message)

	dummy.Status.SpecEcho = dummy.Spec.Message

	if err := r.Client.Status().Update(ctx, dummy); err != nil {
		logger.Error(err, "Failed to update Dummy status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully updated Dummy status")

	// Create a Pod for the Dummy if it doesn't exist
	pod := &v1.Pod{}
	err = r.Client.Get(ctx, types.NamespacedName{Name: dummy.Name}, pod)
	if err != nil && errors.IsNotFound(err) {
		pod = r.newPod(dummy)
		if err := r.Client.Create(ctx, pod); err != nil {
			log.Log.Error(err, "unable to create Pod for Dummy", "pod", pod)
			return ctrl.Result{}, err
		}

		dummy.Status.PodStatus = "Running"
		if err := r.Client.Status().Update(ctx, dummy); err != nil {
			logger.Error(err, "Failed to update Dummy status")
			return ctrl.Result{}, err
		}

		log.Log.Info("created Pod for Dummy", "pod", pod)
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Log.Error(err, "unable to fetch Pod for Dummy")
		return ctrl.Result{}, err
	}

	dummy.Status.PodStatus = string(pod.Status.Phase)
	if err := r.Client.Status().Update(ctx, dummy); err != nil {
		logger.Error(err, "Failed to update Dummy status")
		return ctrl.Result{}, err
	}

	// Delete the Pod if the Dummy is being deleted
	if !dummy.ObjectMeta.DeletionTimestamp.IsZero() {
		err = r.Client.Delete(ctx, pod)
		if err != nil {
			log.Log.Error(err, "unable to delete Pod for Dummy", "pod", pod)
			return ctrl.Result{}, err
		}
		log.Log.Info("deleted Pod for Dummy", "pod", pod)
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *DummyReconciler) newPod(dummy *homeworkv1alpha1.Dummy) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dummy.Name,
			Namespace: dummy.Namespace,
			Labels:    map[string]string{"app": "dummy", "dummy-name": dummy.Name},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}

	_ = ctrl.SetControllerReference(dummy, pod, r.Scheme)
	return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *DummyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homeworkv1alpha1.Dummy{}).
		Owns(&v1.Pod{}).
		Complete(r)
}
