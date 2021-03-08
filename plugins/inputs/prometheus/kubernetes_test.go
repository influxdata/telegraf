package prometheus

import (
	"github.com/ericchiang/k8s"
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"

	v1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"

	"github.com/kubernetes/apimachinery/pkg/fields"
	"github.com/kubernetes/apimachinery/pkg/labels"
)

func TestScrapeURLNoAnnotations(t *testing.T) {
	p := &v1.Pod{Metadata: &metav1.ObjectMeta{}}
	p.GetMetadata().Annotations = map[string]string{}
	url := getScrapeURL(p)
	assert.Nil(t, url)
}

func TestScrapeURLAnnotationsNoScrape(t *testing.T) {
	p := &v1.Pod{Metadata: &metav1.ObjectMeta{}}
	p.Metadata.Name = str("myPod")
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "false"}
	url := getScrapeURL(p)
	assert.Nil(t, url)
}

func TestScrapeURLAnnotations(t *testing.T) {
	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	url := getScrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9102/metrics", *url)
}

func TestScrapeURLAnnotationsCustomPort(t *testing.T) {
	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/port": "9000"}
	url := getScrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9000/metrics", *url)
}

func TestScrapeURLAnnotationsCustomPath(t *testing.T) {
	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "mymetrics"}
	url := getScrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9102/mymetrics", *url)
}

func TestScrapeURLAnnotationsCustomPathWithSep(t *testing.T) {
	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "/mymetrics"}
	url := getScrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9102/mymetrics", *url)
}

func TestAddPod(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	assert.Equal(t, 1, len(prom.kubernetesPods))
}

func TestAddMultipleDuplicatePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Metadata.Name = str("Pod2")
	registerPod(p, prom)
	assert.Equal(t, 1, len(prom.kubernetesPods))
}

func TestAddMultiplePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Metadata.Name = str("Pod2")
	p.Status.PodIP = str("127.0.0.2")
	registerPod(p, prom)
	assert.Equal(t, 2, len(prom.kubernetesPods))
}

func TestDeletePods(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	unregisterPod(p, prom)
	assert.Equal(t, 0, len(prom.kubernetesPods))
}

func TestPodSelector(t *testing.T) {

	cases := []struct {
		expected      []k8s.Option
		labelselector string
		fieldselector string
	}{
		{
			expected: []k8s.Option{
				k8s.QueryParam("labelSelector", "key1=val1,key2=val2,key3"),
				k8s.QueryParam("fieldSelector", "spec.nodeName=ip-1-2-3-4.acme.com"),
			},
			labelselector: "key1=val1,key2=val2,key3",
			fieldselector: "spec.nodeName=ip-1-2-3-4.acme.com",
		},
		{
			expected: []k8s.Option{
				k8s.QueryParam("labelSelector", "key1"),
				k8s.QueryParam("fieldSelector", "spec.nodeName=ip-1-2-3-4.acme.com"),
			},
			labelselector: "key1",
			fieldselector: "spec.nodeName=ip-1-2-3-4.acme.com",
		},
		{
			expected: []k8s.Option{
				k8s.QueryParam("labelSelector", "key1"),
				k8s.QueryParam("fieldSelector", "somefield"),
			},
			labelselector: "key1",
			fieldselector: "somefield",
		},
	}

	for _, c := range cases {
		prom := &Prometheus{
			Log:                     testutil.Logger{},
			KubernetesLabelSelector: c.labelselector,
			KubernetesFieldSelector: c.fieldselector,
		}

		output := podSelector(prom)

		assert.Equal(t, len(output), len(c.expected))
	}
}

func TestPodHasMatchingNamespace(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}, PodNamespace: "default"}

	pod := pod()
	pod.Metadata.Name = str("Pod1")
	pod.Metadata.Namespace = str("default")
	shouldMatch := podHasMatchingNamespace(pod, prom)
	assert.Equal(t, true, shouldMatch)

	pod.Metadata.Name = str("Pod2")
	pod.Metadata.Namespace = str("namespace")
	shouldNotMatch := podHasMatchingNamespace(pod, prom)
	assert.Equal(t, false, shouldNotMatch)
}

func TestPodHasMatchingLabelSelector(t *testing.T) {
	labelSelectorString := "label0==label0,label1=label1,label2!=label,label3 in (label1,label2, label3),label4 notin (label1, label2,label3),label5,!label6"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesLabelSelector: labelSelectorString}

	pod := pod()
	pod.Metadata.Labels = make(map[string]string)
	pod.Metadata.Labels["label0"] = "label0"
	pod.Metadata.Labels["label1"] = "label1"
	pod.Metadata.Labels["label2"] = "label2"
	pod.Metadata.Labels["label3"] = "label3"
	pod.Metadata.Labels["label4"] = "label4"
	pod.Metadata.Labels["label5"] = "label5"

	labelSelector, err := labels.Parse(prom.KubernetesLabelSelector)
	assert.Equal(t, err, nil)
	assert.Equal(t, true, podHasMatchingLabelSelector(pod, labelSelector))
}

func TestPodHasMatchingFieldSelector(t *testing.T) {
	fieldSelectorString := "status.podIP=127.0.0.1,spec.restartPolicy=Always,spec.NodeName!=nodeName"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}
	pod := pod()
	pod.Spec.RestartPolicy = str("Always")
	pod.Spec.NodeName = str("node1000")

	fieldSelector, err := fields.ParseSelector(prom.KubernetesFieldSelector)
	assert.Equal(t, err, nil)
	assert.Equal(t, true, podHasMatchingFieldSelector(pod, fieldSelector))
}

func TestInvalidFieldSelector(t *testing.T) {
	fieldSelectorString := "status.podIP=127.0.0.1,spec.restartPolicy=Always,spec.NodeName!=nodeName,spec.nodeName"
	prom := &Prometheus{Log: testutil.Logger{}, KubernetesFieldSelector: fieldSelectorString}
	pod := pod()
	pod.Spec.RestartPolicy = str("Always")
	pod.Spec.NodeName = str("node1000")

	_, err := fields.ParseSelector(prom.KubernetesFieldSelector)
	assert.NotEqual(t, err, nil)
}

func pod() *v1.Pod {
	p := &v1.Pod{Metadata: &metav1.ObjectMeta{}, Status: &v1.PodStatus{}, Spec: &v1.PodSpec{}}
	p.Status.PodIP = str("127.0.0.1")
	p.Metadata.Name = str("myPod")
	p.Metadata.Namespace = str("default")
	return p
}

func str(x string) *string {
	return &x
}
