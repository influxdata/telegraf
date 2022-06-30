package prometheus

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/influxdata/telegraf/testutil"
)

func TestScrapeURLNoAnnotations(t *testing.T) {
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{}}
	p.Annotations = map[string]string{}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Nil(t, url)
}

func TestScrapeURLAnnotationsNoScrape(t *testing.T) {
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{}}
	p.Name = "myPod"
	p.Annotations = map[string]string{"prometheus.io/scrape": "false"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Nil(t, url)
}

func TestScrapeURLAnnotations(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/metrics", url.String())
}

func TestScrapeURLAnnotationsCustomPort(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/port": "9000"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9000/metrics", url.String())
}

func TestScrapeURLAnnotationsCustomPath(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "mymetrics"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/mymetrics", url.String())
}

func TestScrapeURLAnnotationsCustomPathWithSep(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "/mymetrics"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/mymetrics", url.String())
}

func TestScrapeURLAnnotationsCustomPathWithQueryParameters(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "/v1/agent/metrics?format=prometheus"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/v1/agent/metrics?format=prometheus", url.String())
}

func TestScrapeURLAnnotationsCustomPathWithFragment(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "/v1/agent/metrics#prometheus"}
	url, err := getScrapeURL(p)
	require.NoError(t, err)
	require.Equal(t, "http://127.0.0.1:9102/v1/agent/metrics#prometheus", url.String())
}

func TestAddPod(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	require.Equal(t, 1, len(prom.kubernetesPods))
}

func TestAddMultipleDuplicatePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Name = "Pod2"
	registerPod(p, prom)
	require.Equal(t, 1, len(prom.kubernetesPods))
}

func TestAddMultiplePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Name = "Pod2"
	p.Status.PodIP = "127.0.0.2"
	registerPod(p, prom)
	require.Equal(t, 2, len(prom.kubernetesPods))
}

func TestDeletePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	unregisterPod(p, prom)
	require.Equal(t, 0, len(prom.kubernetesPods))
}

func TestPodHasMatchingNamespace(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, PodNamespace: "default"}

	pod := pod()
	pod.Name = "Pod1"
	pod.Namespace = "default"
	shouldMatch := podHasMatchingNamespace(pod, prom)
	require.Equal(t, true, shouldMatch)

	pod.Name = "Pod2"
	pod.Namespace = "namespace"
	shouldNotMatch := podHasMatchingNamespace(pod, prom)
	require.Equal(t, false, shouldNotMatch)
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
	require.Equal(t, err, nil)
	require.Equal(t, true, podHasMatchingLabelSelector(pod, labelSelector))
}

func TestPodHasMatchingFieldSelector(t *testing.T) {
	fieldSelectorString := "status.podIP=127.0.0.1,spec.restartPolicy=Always,spec.NodeName!=nodeName"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}
	pod := pod()
	pod.Spec.RestartPolicy = "Always"
	pod.Spec.NodeName = "node1000"

	fieldSelector, err := fields.ParseSelector(prom.KubernetesFieldSelector)
	require.Equal(t, err, nil)
	require.Equal(t, true, podHasMatchingFieldSelector(pod, fieldSelector))
}

func TestInvalidFieldSelector(t *testing.T) {
	fieldSelectorString := "status.podIP=127.0.0.1,spec.restartPolicy=Always,spec.NodeName!=nodeName,spec.nodeName"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}
	pod := pod()
	pod.Spec.RestartPolicy = "Always"
	pod.Spec.NodeName = "node1000"

	_, err := fields.ParseSelector(prom.KubernetesFieldSelector)
	require.NotEqual(t, err, nil)
}

func pod() *corev1.Pod {
	p := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{}, Status: corev1.PodStatus{}, Spec: corev1.PodSpec{}}
	p.Status.PodIP = "127.0.0.1"
	p.Name = "myPod"
	p.Namespace = "default"
	return p
}
