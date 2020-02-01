package prometheus

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"

	v1 "github.com/ericchiang/k8s/apis/core/v1"
	metav1 "github.com/ericchiang/k8s/apis/meta/v1"
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

func TestAddPodOnlyonSameNode(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Spec.NodeName = str("ip-10-1-2-3.acme.com")
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	if podOnNode(p, "ip-10-1-2-3.acme.com") {
		registerPod(p, prom)
	}
	assert.Equal(t, 1, len(prom.kubernetesPods))
}

func TestDoNotAddPodIfNotonSameNode(t *testing.T) {
	prom := &Prometheus{Log: testutil.Logger{}}

	p := pod()
	p.Spec.NodeName = str("ip-10-1-2-3.acme.com")
	p.Metadata.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	if podOnNode(p, "ip-10-4-5-6.acme.com") {
		registerPod(p, prom)
	}
	assert.Equal(t, 0, len(prom.kubernetesPods))
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
