/*
Copyright 2022.
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
	"encoding/json"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	cachev1alpha1 "memcached-operator/api/v1alpha1"
)

// MemcachedReconciler reconciles a Memcached object
type MemcachedReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Memcached object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *MemcachedReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Memcached instance
	memcached := &cachev1alpha1.Memcached{}
	err := r.Get(ctx, req.NamespacedName, memcached)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Memcached resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Memcached")
		return ctrl.Result{}, err
	}

	// Check if the chaosblade already exists, if not create a new one
	found := &unstructured.Unstructured{}
	found.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "chaosblade.io",
		Kind:    "ChaosBlade",
		Version: "v1alpha1",
	})
	err = r.Get(ctx, types.NamespacedName{Name: memcached.Name}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new chaosblade
		cb := r.chaosbladeForMemcached(memcached)
		log.Info("Creating a new Chaosblade", "Chaosblade.Name", memcached.Name)
		err = r.Create(ctx, cb)
		if err != nil {
			log.Error(err, "Failed to create new Chaosblade", "Chaosblade.Name", memcached.Name)
			return ctrl.Result{}, err
		}
		// Update the Memcached status with the chaoblade status
		time.Sleep(5 * time.Second)

		_ = r.Get(ctx, types.NamespacedName{Name: memcached.Name}, cb)
		cbStatus, _, _ := unstructured.NestedMap(cb.Object, "status")
		cbstatusData, err := json.Marshal(cbStatus)
		if err != nil {
			log.Error(err, "Failed to Marshal.")
			return ctrl.Result{}, err
		}

		tmpmStatus := &cachev1alpha1.MemcachedStatus{}
		err = json.Unmarshal(cbstatusData, tmpmStatus)
		if err != nil {
			log.Error(err, "Failed to Unmarshal.")
			return ctrl.Result{}, err
		}

		// Update status.Nodes if needed
		if !reflect.DeepEqual(tmpmStatus, memcached.Status) {
			memcached.Status = *tmpmStatus
			err := r.Status().Update(ctx, memcached)
			if err != nil {
				log.Error(err, "Failed to update Memcached status")
				return ctrl.Result{}, err
			}
		}

		// Chaosblade created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	// size := memcached.Spec.Size
	// if *found.Spec.Replicas != size {
	// 	found.Spec.Replicas = &size
	// 	err = r.Update(ctx, found)
	// 	if err != nil {
	// 		log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	// Ask to requeue after 1 minute in order to give enough time for the
	// 	// pods be created on the cluster side and the operand be able
	// 	// to do the next update step accurately.
	// 	return ctrl.Result{RequeueAfter: time.Minute}, nil
	// }

	// Update the Memcached status with the pod names
	// List the pods for this memcached's deployment
	// 	podList := &corev1.PodList{}
	// 	listOpts := []client.ListOption{
	// 		client.InNamespace(memcached.Namespace),
	// 		client.MatchingLabels(labelsForMemcached(memcached.Name)),
	// 	}
	// 	if err = r.List(ctx, podList, listOpts...); err != nil {
	// 		log.Error(err, "Failed to list pods", "Memcached.Namespace", memcached.Namespace, "Memcached.Name", memcached.Name)
	// 		return ctrl.Result{}, err
	// 	}
	// 	podNames := getPodNames(podList.Items)

	// // Update status.Nodes if needed
	// if !reflect.DeepEqual(podNames, memcached.Status.Nodes) {
	// 	memcached.Status.Nodes = podNames
	// 	err := r.Status().Update(ctx, memcached)
	// 	if err != nil {
	// 		log.Error(err, "Failed to update Memcached status")
	// 		return ctrl.Result{}, err
	// 	}
	// }

	return ctrl.Result{}, nil
}

// chaosbladeForMemcached returns a memcached Chaosblade object
func (r *MemcachedReconciler) chaosbladeForMemcached(m *cachev1alpha1.Memcached) *unstructured.Unstructured {
	// ls := labelsForMemcached(m.Name)
	// replicas := m.Spec.Size
	cb := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": m.Name,
			},
			"spec": m.Spec,
		},
	}
	cb.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "chaosblade.io",
		Kind:    "ChaosBlade",
		Version: "v1alpha1",
	})

	// cb := &unstructured.Unstructured{
	// 	Object: map[string]interface{}{
	// 		"metadata": map[string]interface{}{
	// 			"name": m.Name,
	// 		},
	// 		"spec": map[string]interface{}{
	// 			"experiments": []map[string]interface{}{
	// 				{
	// 					"scope":  "node",
	// 					"target": "cpu",
	// 					"action": "fullload",
	// 					"matchers": []map[string]interface{}{
	// 						{
	// 							"name": "names",
	// 							"value": []string{
	// 								"k8s-node1",
	// 							},
	// 						},
	// 						{
	// 							"name": "cpu-percent",
	// 							"value": []string{
	// 								"10",
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// cb.SetGroupVersionKind(schema.GroupVersionKind{
	// 	Group:   "chaosblade.io",
	// 	Kind:    "ChaosBlade",
	// 	Version: "v1alpha1",
	// })

	// dep := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      m.Name,
	// 		Namespace: m.Namespace,
	// 	},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Replicas: &replicas,
	// 		Selector: &metav1.LabelSelector{
	// 			MatchLabels: ls,
	// 		},
	// 		Template: corev1.PodTemplateSpec{
	// 			ObjectMeta: metav1.ObjectMeta{
	// 				Labels: ls,
	// 			},
	// 			Spec: corev1.PodSpec{
	// 				// Ensure restrictive standard for the Pod.
	// 				// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
	// 				SecurityContext: &corev1.PodSecurityContext{
	// 					// WARNING: Ensure that the image used defines an UserID in the Dockerfile
	// 					// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
	// 					// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
	// 					// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
	// 					// "RunAsUser" fields empty.
	// 					RunAsNonRoot: &[]bool{true}[0],
	// 					// Please ensure that you can use SeccompProfile and do NOT use
	// 					// this field if your project must work on old Kubernetes
	// 					// versions < 1.19 or on vendors versions which
	// 					// do NOT support this field by default (i.e. Openshift < 4.11)
	// 					SeccompProfile: &corev1.SeccompProfile{
	// 						Type: corev1.SeccompProfileTypeRuntimeDefault,
	// 					},
	// 				},
	// 				Containers: []corev1.Container{{
	// 					Image: "memcached:1.4.36-alpine",
	// 					Name:  "memcached",
	// 					// Ensure restrictive context for the container
	// 					// More info: https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted
	// 					SecurityContext: &corev1.SecurityContext{
	// 						// WARNING: Ensure that the image used defines an UserID in the Dockerfile
	// 						// otherwise the Pod will not run and will fail with "container has runAsNonRoot and image has non-numeric user"".
	// 						// If you want your workloads admitted in namespaces enforced with the restricted mode in OpenShift/OKD vendors
	// 						// then, you MUST ensure that the Dockerfile defines a User ID OR you MUST leave the "RunAsNonRoot" and
	// 						// "RunAsUser" fields empty.
	// 						RunAsNonRoot:             &[]bool{true}[0],
	// 						AllowPrivilegeEscalation: &[]bool{false}[0],
	// 						Capabilities: &corev1.Capabilities{
	// 							Drop: []corev1.Capability{
	// 								"ALL",
	// 							},
	// 						},
	// 						// The memcached image does not use a non-zero numeric user as the default user.
	// 						// Due to RunAsNonRoot field being set to true, we need to force the user in the
	// 						// container to a non-zero numeric user. We do this using the RunAsUser field.
	// 						// However, if you are looking to provide solution for K8s vendors like OpenShift
	// 						// be aware that you can not run under its restricted-v2 SCC if you set this value.
	// 						RunAsUser: &[]int64{1000}[0],
	// 					},
	// 					Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
	// 					Ports: []corev1.ContainerPort{{
	// 						ContainerPort: 11211,
	// 						Name:          "memcached",
	// 					}},
	// 				}},
	// 			},
	// 		},
	// 	},
	// }
	// Set Memcached instance as the owner and controller
	// ctrl.SetControllerReference(m, dep, r.Scheme)
	return cb
}

// labelsForMemcached returns the labels for selecting the resources
// belonging to the given memcached CR name.
func labelsForMemcached(name string) map[string]string {
	return map[string]string{"app": "memcached", "memcached_cr": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *MemcachedReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.Memcached{}).
		// Owns(&appsv1.Deployment{}).
		Complete(r)
}
