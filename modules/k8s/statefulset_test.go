//go:build kubeall || kubernetes
// +build kubeall kubernetes

// NOTE: we have build tags to differentiate kubernetes tests from non-kubernetes tests. This is done because minikube
// is heavy and can interfere with docker related tests in terratest. Specifically, many of the tests start to fail with
// `connection refused` errors from `minikube`. To avoid overloading the system, we run the kubernetes tests and helm
// tests separately from the others. This may not be necessary if you have a sufficiently powerful machine.  We
// recommend at least 4 cores and 16GB of RAM if you want to run all the tests together.

package k8s

import (
	"fmt"
	"time"

	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetStatefulSetEReturnsError(t *testing.T) {
	t.Parallel()

	options := NewKubectlOptions("", "", "")
	_, err := GetStatefulSetE(t, options, "nginx-statefulset")
	require.Error(t, err)
}

func TestGetStatefulSets(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "", uniqueID)
	configData := fmt.Sprintf(ExampleStatefulSetYAMLTemplate, uniqueID)
	KubectlApplyFromString(t, options, configData)
	defer KubectlDeleteFromString(t, options, configData)

	statefulset := GetStatefulSet(t, options, "nginx-statefulset")
	require.Equal(t, statefulset.Name, "nginx-statefulset")
	require.Equal(t, statefulset.Namespace, uniqueID)
}

func TestListStatefulSets(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "", uniqueID)
	configData := fmt.Sprintf(ExampleStatefulSetYAMLTemplate, uniqueID)
	KubectlApplyFromString(t, options, configData)
	defer KubectlDeleteFromString(t, options, configData)

	statefulsets := ListStatefulSets(t, options, metav1.ListOptions{})
	require.Equal(t, len(statefulsets), 1)

	statefulset := statefulsets[0]
	require.Equal(t, statefulset.Name, "nginx-statefulset")
	require.Equal(t, statefulset.Namespace, uniqueID)
}

func TestWaitUntilStatefulSetAvailable(t *testing.T) {
	t.Parallel()

	uniqueID := strings.ToLower(random.UniqueId())
	options := NewKubectlOptions("", "", uniqueID)
	configData := fmt.Sprintf(ExampleStatefulSetYAMLTemplate, uniqueID)
	KubectlApplyFromString(t, options, configData)
	defer KubectlDeleteFromString(t, options, configData)

	WaitUntilStatefulSetAvailable(t, options, "nginx-statefulset", 60, 1*time.Second)
}

func TestTestIsStatefulSetAvailable(t *testing.T) {
	testCases := []struct {
		title          string
		statefulset    *appsv1.StatefulSet
		expectedResult bool
	}{
		{
			title: "TestIsStatefulSetAvailableWithStatusReadyReplicasEqualToSpecReplicas",
			statefulset: &appsv1.StatefulSet{
				Status: appsv1.StatefulSetStatus{
					Replicas: 1,
					ReadyReplicas: 1,
				},
			},
			expectedResult: true,
		},
		{
			title: "TestIsStatefulSetAvailableWithStatusReadyReplicasNotEqualToSpecReplicas",
			statefulset: &appsv1.StatefulSet{
				Status: appsv1.StatefulSetStatus{
					Replicas: 1,
					ReadyReplicas: 0,
				},
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()
			actualResult := IsStatefulSetAvailable(tc.statefulset)
			require.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

const ExampleStatefulSetYAMLTemplate = `---
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nginx-statefulset
  labels:
    app: nginx
spec:
  updateStrategy:
    rollingUpdate:
      partition: 0
  serviceName: nginx
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.15.7
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /
            port: 80
`
