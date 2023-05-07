package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/testing"
)

// ListStatefulSets will look for statefulSets in the given namespace that match the given filters and return them. This will
// fail the test if there is an error.
func ListStatefulSets(t testing.TestingT, options *KubectlOptions, filters metav1.ListOptions) []appsv1.StatefulSet {
	statefulset, err := ListStatefulSetsE(t, options, filters)
	require.NoError(t, err)
	return statefulset
}

// ListStatefulSetsE will look for statefulsets in the given namespace that match the given filters and return them.
func ListStatefulSetsE(t testing.TestingT, options *KubectlOptions, filters metav1.ListOptions) ([]appsv1.StatefulSet, error) {
	clientset, err := GetKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	statefulsets, err := clientset.AppsV1().StatefulSets(options.Namespace).List(context.Background(), filters)
	if err != nil {
		return nil, err
	}
	return statefulsets.Items, nil
}

// GetStatefulSet returns a Kubernetes statefulset resource in the provided namespace with the given name. This will
// fail the test if there is an error.
func GetStatefulSet(t testing.TestingT, options *KubectlOptions, statefulesetName string) *appsv1.StatefulSet {
	statefulset, err := GetStatefulSetE(t, options, statefulesetName)
	require.NoError(t, err)
	return statefulset
}

// GetStatefulSetE returns a Kubernetes statefulset resource in the provided namespace with the given name.
func GetStatefulSetE(t testing.TestingT, options *KubectlOptions, statefulsetName string) (*appsv1.StatefulSet, error) {
	clientset, err := GetKubernetesClientFromOptionsE(t, options)
	if err != nil {
		return nil, err
	}
	return clientset.AppsV1().StatefulSets(options.Namespace).Get(context.Background(), statefulsetName, metav1.GetOptions{})
}

// WaitUntilStatefulSetAvailable waits until all pods within the statefulset are ready and started,
// retrying the check for the specified amount of times, sleeping
// for the provided duration between each try.
// This will fail the test if there is an error.
func WaitUntilStatefulSetAvailable(t testing.TestingT, options *KubectlOptions, statefulsetName string, retries int, sleepBetweenRetries time.Duration) {
	require.NoError(t, WaitUntilStatefulSetAvailableE(t, options, statefulsetName, retries, sleepBetweenRetries))
}

// WaitUntilStatefulSetAvailableE waits until all pods within the statefulset are ready and started,
// retrying the check for the specified amount of times, sleeping
// for the provided duration between each try.
func WaitUntilStatefulSetAvailableE(
	t testing.TestingT,
	options *KubectlOptions,
	statefulsetName string,
	retries int,
	sleepBetweenRetries time.Duration,
) error {
	statusMsg := fmt.Sprintf("Wait for statefulset %s to be provisioned.", statefulsetName)
	message, err := retry.DoWithRetryE(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			statefulset, err := GetStatefulSetE(t, options, statefulsetName)
			if err != nil {
				return "", err
			}
			if !IsStatefulSetAvailable(statefulset) {
				return "", NewStatefulSetNotAvailableError(statefulset)
			}
			return "Statefulset is now available", nil
		},
	)
	if err != nil {
		logger.Logf(t, "Timedout waiting for Statefulset to be provisioned: %s", err)
		return err
	}
	logger.Logf(t, message)
	return nil
}

// IsStatefulSetAvailable returns true if all pods within the deployment are ready and started
func IsStatefulSetAvailable(statefulset *appsv1.StatefulSet) bool {
	var acutalReplicas int32
	acutalReplicas = 1
	if statefulset.Spec.Replicas != nil {
		acutalReplicas = *statefulset.Spec.Replicas
	}
	return (statefulset.Status.ReadyReplicas == acutalReplicas)
}
