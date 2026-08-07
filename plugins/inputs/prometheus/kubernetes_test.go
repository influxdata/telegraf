package prometheus

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/influxdata/telegraf/testutil"
)

func initPrometheus() *Prometheus {
	prom := &Prometheus{Log: testutil.Logger{}}
	prom.MonitorKubernetesPodsScheme = "http"
	prom.MonitorKubernetesPodsPort = 9102
	prom.MonitorKubernetesPodsPath = "/metrics"
	prom.MonitorKubernetesPodsMethod = monitorMethodAnnotations
	prom.kubernetesPods = map[podID]urlAndAddress{}
	return prom
}

func TestScrapeURLNoAnnotations(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{}}
	p.Annotations = map[string]string{}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Nil(t, url)
}

func TestScrapeURLNoAnnotationsScrapeConfig(t *testing.T) {
	prom := initPrometheus()
	prom.MonitorKubernetesPodsMethod = monitorMethodSettingsAndAnnotations

	p := pod()
	p.Annotations = map[string]string{}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/metrics", url.String())
}

func TestScrapeURLScrapeConfigCustom(t *testing.T) {
	prom := initPrometheus()
	prom.MonitorKubernetesPodsMethod = monitorMethodSettingsAndAnnotations

	prom.MonitorKubernetesPodsScheme = "https"
	prom.MonitorKubernetesPodsPort = 9999
	prom.MonitorKubernetesPodsPath = "/svc/metrics"
	p := pod()
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "https://127.0.0.1:9999/svc/metrics", url.String())
}

func TestScrapeURLAnnotations(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}
	p := pod()
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/metrics", url.String())
}

func TestScrapeURLAnnotationsScrapeConfig(t *testing.T) {
	prom := initPrometheus()
	prom.MonitorKubernetesPodsMethod = monitorMethodSettingsAndAnnotations
	p := pod()
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/metrics", url.String())
}

func TestScrapeURLAnnotationsCustomPort(t *testing.T) {
	prom := initPrometheus()
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/port": "9000"}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9000/metrics", url.String())
}

func TestScrapeURLAnnotationsCustomPortScrapeConfig(t *testing.T) {
	prom := initPrometheus()
	prom.MonitorKubernetesPodsMethod = monitorMethodSettingsAndAnnotations
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/port": "9000"}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9000/metrics", url.String())
}

func TestScrapeURLAnnotationsCustomPath(t *testing.T) {
	prom := initPrometheus()
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/path": "mymetrics"}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/mymetrics", url.String())
}

func TestScrapeURLAnnotationsCustomPathWithSep(t *testing.T) {
	prom := initPrometheus()
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/path": "/mymetrics"}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/mymetrics", url.String())
}

func TestScrapeURLAnnotationsCustomPathWithQueryParameters(t *testing.T) {
	prom := initPrometheus()
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/path": "/v1/agent/metrics?format=prometheus"}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/v1/agent/metrics?format=prometheus", url.String())
}

func TestScrapeURLAnnotationsCustomPathWithFragment(t *testing.T) {
	prom := initPrometheus()
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/path": "/v1/agent/metrics#prometheus"}
	url, err := getScrapeURL(p, prom)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/v1/agent/metrics#prometheus", url.String())
}

func TestAddPod(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	require.Len(t, prom.kubernetesPods, 1)
}

func TestAddPodScrapeConfig(t *testing.T) {
	prom := initPrometheus()
	prom.MonitorKubernetesPodsMethod = monitorMethodSettingsAndAnnotations

	p := pod()
	p.Annotations = map[string]string{}
	registerPod(p, prom)
	require.Len(t, prom.kubernetesPods, 1)
}

func TestAddMultipleDuplicatePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Name = "Pod2"
	registerPod(p, prom)

	urls, err := prom.getAllURLs()
	require.NoError(t, err)
	require.Len(t, urls, 1)
}

func TestAddMultiplePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Name = "Pod2"
	p.Status.PodIP = "127.0.0.2"
	registerPod(p, prom)
	require.Len(t, prom.kubernetesPods, 2)
}

func TestDeletePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)

	id, err := cache.MetaNamespaceKeyFunc(p)
	require.NoError(t, err)
	unregisterPod(podID(id), prom)
	require.Empty(t, prom.kubernetesPods)
}

func TestKeepDefaultNamespaceLabelName(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)

	id, err := cache.MetaNamespaceKeyFunc(p)
	require.NoError(t, err)
	tags := prom.kubernetesPods[podID(id)].tags
	require.Equal(t, "default", tags["namespace"])
}

func TestChangeNamespaceLabelName(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, PodNamespaceLabelName: "pod_namespace", kubernetesPods: map[podID]urlAndAddress{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)

	id, err := cache.MetaNamespaceKeyFunc(p)
	require.NoError(t, err)
	tags := prom.kubernetesPods[podID(id)].tags
	require.Equal(t, "default", tags["pod_namespace"])
	require.Empty(t, tags["namespace"])
}

func TestPodHasMatchingNamespace(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, PodNamespace: "default"}

	pod := pod()
	pod.Name = "Pod1"
	pod.Namespace = "default"
	shouldMatch := podHasMatchingNamespace(pod, prom)
	require.True(t, shouldMatch)

	pod.Name = "Pod2"
	pod.Namespace = "namespace"
	shouldNotMatch := podHasMatchingNamespace(pod, prom)
	require.False(t, shouldNotMatch)
}

func TestPodHasMatchingLabelSelector(t *testing.T) {
	labelSelectorString := "label0==label0,label1=label1,label2!=label,label3 in (label1,label2, label3),label4 notin (label1, label2,label3),label5,!label6"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesLabelSelector: labelSelectorString}

	pod := pod()
	pod.Labels = make(map[string]string)
	pod.Labels["label0"] = "label0"
	pod.Labels["label1"] = "label1"
	pod.Labels["label2"] = "label2"
	pod.Labels["label3"] = "label3"
	pod.Labels["label4"] = "label4"
	pod.Labels["label5"] = "label5"

	labelSelector, err := labels.Parse(prom.KubernetesLabelSelector)
	require.NoError(t, err)
	require.True(t, podHasMatchingLabelSelector(pod, labelSelector))
}

func TestPodHasMatchingFieldSelector(t *testing.T) {
	fieldSelectorString := "status.podIP=127.0.0.1,spec.restartPolicy=Always,spec.NodeName!=nodeName"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}
	pod := pod()
	pod.Spec.RestartPolicy = "Always"
	pod.Spec.NodeName = "node1000"

	fieldSelector, err := fields.ParseSelector(prom.KubernetesFieldSelector)
	require.NoError(t, err)
	require.True(t, podHasMatchingFieldSelector(pod, fieldSelector))
}

func TestInvalidFieldSelector(t *testing.T) {
	fieldSelectorString := "status.podIP=127.0.0.1,spec.restartPolicy=Always,spec.NodeName!=nodeName,spec.nodeName"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}
	pod := pod()
	pod.Spec.RestartPolicy = "Always"
	pod.Spec.NodeName = "node1000"

	_, err := fields.ParseSelector(prom.KubernetesFieldSelector)
	require.Error(t, err)
}

func TestAnnotationFilters(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{
		"prometheus.io/scrape": "true",
		"includeme":            "true",
		"excludeme":            "true",
		"neutral":              "true",
	}

	cases := []struct {
		desc         string
		include      []string
		exclude      []string
		expectedTags []string
	}{
		{"Just include",
			[]string{"includeme"},
			nil,
			[]string{"includeme"}},
		{"Just exclude",
			nil,
			[]string{"excludeme"},
			[]string{"includeme", "neutral"}},
		{"Include & exclude",
			[]string{"includeme"},
			[]string{"exludeme"},
			[]string{"includeme"}},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}
			prom.PodAnnotationInclude = tc.include
			prom.PodAnnotationExclude = tc.exclude
			require.NoError(t, prom.initFilters())
			registerPod(p, prom)
			for _, pd := range prom.kubernetesPods {
				for _, tagKey := range tc.expectedTags {
					require.Contains(t, pd.tags, tagKey)
				}
			}
		})
	}
}

func TestLabelFilters(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	p.Labels = map[string]string{
		"includeme": "true",
		"excludeme": "true",
		"neutral":   "true",
	}

	cases := []struct {
		desc         string
		include      []string
		exclude      []string
		expectedTags []string
	}{
		{"Just include",
			[]string{"includeme"},
			nil,
			[]string{"includeme"}},
		{"Just exclude",
			nil,
			[]string{"excludeme"},
			[]string{"includeme", "neutral"}},
		{"Include & exclude",
			[]string{"includeme"},
			[]string{"exludeme"},
			[]string{"includeme"}},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			prom := &Prometheus{Log: testutil.Logger{}, kubernetesPods: map[podID]urlAndAddress{}}
			prom.PodLabelInclude = tc.include
			prom.PodLabelExclude = tc.exclude
			require.NoError(t, prom.initFilters())
			registerPod(p, prom)
			for _, pd := range prom.kubernetesPods {
				for _, tagKey := range tc.expectedTags {
					require.Contains(t, pd.tags, tagKey)
				}
			}
		})
	}
}

func TestInformerFactoryRefCounting(t *testing.T) {
	resetInformerFactoryState(t)

	var shutdownCalled atomic.Int32
	mock := &mockSharedInformerFactory{
		onShutdown: func() { shutdownCalled.Add(1) },
	}

	// Simulate two instances registered for the same namespace
	informerfactoryMu.Lock()
	informerfactory = map[string]informers.SharedInformerFactory{"default": mock}
	informerfactoryRefs = map[string]int{"default": 2}
	informerfactoryMu.Unlock()

	// First Stop — should decrement ref count but not shutdown
	_, cancel1 := context.WithCancel(context.Background())
	p1 := &Prometheus{
		MonitorPods:  true,
		PodNamespace: "default",
		cancel:       cancel1,
	}
	p1.Stop()

	// Read under lock, assert outside to avoid deadlock if assertion fails
	informerfactoryMu.Lock()
	refCount := informerfactoryRefs["default"]
	_, exists := informerfactory["default"]
	informerfactoryMu.Unlock()
	require.Equal(t, 1, refCount)
	require.True(t, exists)
	require.Equal(t, int32(0), shutdownCalled.Load())

	// Second Stop — should shutdown and remove
	_, cancel2 := context.WithCancel(context.Background())
	p2 := &Prometheus{
		MonitorPods:  true,
		PodNamespace: "default",
		cancel:       cancel2,
	}
	p2.Stop()

	informerfactoryMu.Lock()
	_, refsExist := informerfactoryRefs["default"]
	_, factoryExist := informerfactory["default"]
	informerfactoryMu.Unlock()
	require.False(t, refsExist)
	require.False(t, factoryExist)
	require.Equal(t, int32(1), shutdownCalled.Load())
}

func TestInformerFactoryMultipleNamespaces(t *testing.T) {
	resetInformerFactoryState(t)

	var shutdownA, shutdownB atomic.Int32
	mockA := &mockSharedInformerFactory{
		onShutdown: func() { shutdownA.Add(1) },
	}
	mockB := &mockSharedInformerFactory{
		onShutdown: func() { shutdownB.Add(1) },
	}

	informerfactoryMu.Lock()
	informerfactory = map[string]informers.SharedInformerFactory{
		"ns-a": mockA,
		"ns-b": mockB,
	}
	informerfactoryRefs = map[string]int{
		"ns-a": 1,
		"ns-b": 1,
	}
	informerfactoryMu.Unlock()

	// Stop instance in ns-a — should shutdown ns-a only
	_, cancelA := context.WithCancel(context.Background())
	pa := &Prometheus{
		MonitorPods:  true,
		PodNamespace: "ns-a",
		cancel:       cancelA,
	}
	pa.Stop()

	informerfactoryMu.Lock()
	_, nsAExists := informerfactory["ns-a"]
	_, nsBExists := informerfactory["ns-b"]
	informerfactoryMu.Unlock()
	require.False(t, nsAExists)
	require.True(t, nsBExists)
	require.Equal(t, int32(1), shutdownA.Load())
	require.Equal(t, int32(0), shutdownB.Load())

	// Stop instance in ns-b — should shutdown ns-b
	_, cancelB := context.WithCancel(context.Background())
	pb := &Prometheus{
		MonitorPods:  true,
		PodNamespace: "ns-b",
		cancel:       cancelB,
	}
	pb.Stop()

	informerfactoryMu.Lock()
	_, nsBStillExists := informerfactory["ns-b"]
	informerfactoryMu.Unlock()
	require.False(t, nsBStillExists)
	require.Equal(t, int32(1), shutdownB.Load())
}

func TestInformerFactoryConcurrentStop(t *testing.T) {
	resetInformerFactoryState(t)

	var shutdownCount atomic.Int32
	mock := &mockSharedInformerFactory{
		onShutdown: func() { shutdownCount.Add(1) },
	}

	const numInstances = 10
	informerfactoryMu.Lock()
	informerfactory = map[string]informers.SharedInformerFactory{"default": mock}
	informerfactoryRefs = map[string]int{"default": numInstances}
	informerfactoryMu.Unlock()

	// Stop all instances concurrently — race detector verifies thread safety
	var wg sync.WaitGroup
	for range numInstances {
		wg.Go(func() {
			_, cancel := context.WithCancel(context.Background())
			p := &Prometheus{
				MonitorPods:  true,
				PodNamespace: "default",
				cancel:       cancel,
			}
			p.Stop()
		})
	}
	wg.Wait()

	informerfactoryMu.Lock()
	_, refsExist := informerfactoryRefs["default"]
	_, factoryExist := informerfactory["default"]
	informerfactoryMu.Unlock()
	require.False(t, refsExist)
	require.False(t, factoryExist)
	require.Equal(t, int32(1), shutdownCount.Load())
}

func pod() *corev1.Pod {
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{}, Status: corev1.PodStatus{}, Spec: corev1.PodSpec{}}
	p.Status.PodIP = "127.0.0.1"
	p.Name = "myPod"
	p.Namespace = "default"
	return p
}

type mockSharedInformerFactory struct {
	informers.SharedInformerFactory
	onShutdown func()
}

func (m *mockSharedInformerFactory) Shutdown() {
	if m.onShutdown != nil {
		m.onShutdown()
	}
}

func resetInformerFactoryState(t *testing.T) {
	t.Helper()
	informerfactoryMu.Lock()
	informerfactory = nil
	informerfactoryRefs = nil
	informerfactoryMu.Unlock()
	t.Cleanup(func() {
		informerfactoryMu.Lock()
		informerfactory = nil
		informerfactoryRefs = nil
		informerfactoryMu.Unlock()
	})
}
