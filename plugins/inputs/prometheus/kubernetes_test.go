package prometheus

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestScrapeURLNoAnnotations(t *testing.T) {
	p := &v1.Pod{}
	p.Annotations = map[string]string{}
	url := scrapeURL(p)
	assert.Nil(t, url)
}
func TestScrapeURLAnnotationsNoScrape(t *testing.T) {
	p := &v1.Pod{}
	p.Name = "myPod"
	p.Annotations = map[string]string{"prometheus.io/scrape": "false"}
	url := scrapeURL(p)
	assert.Nil(t, url)
}
func TestScrapeURLAnnotations(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	url := scrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9102/metrics", *url)
}
func TestScrapeURLAnnotationsCustomPort(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/port": "9000"}
	url := scrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9000/metrics", *url)
}
func TestScrapeURLAnnotationsCustomPath(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "mymetrics"}
	url := scrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9102/mymetrics", *url)
}

func TestScrapeURLAnnotationsCustomPathWithSep(t *testing.T) {
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true", "prometheus.io/path": "/mymetrics"}
	url := scrapeURL(p)
	assert.Equal(t, "http://127.0.0.1:9102/mymetrics", *url)
}

func TestAddPod(t *testing.T) {
	prom := &Prometheus{lock: &sync.Mutex{}}
	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	assert.Equal(t, 1, len(prom.KubernetesPods))
}
func TestAddMultiplePods(t *testing.T) {
	prom := &Prometheus{lock: &sync.Mutex{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	p.Name = "Pod2"
	registerPod(p, prom)
	assert.Equal(t, 2, len(prom.KubernetesPods))
}
func TestDeletePods(t *testing.T) {
	prom := &Prometheus{lock: &sync.Mutex{}}

	p := pod()
	p.Annotations = map[string]string{"prometheus.io/scrape": "true"}
	registerPod(p, prom)
	unregisterPod(p, prom)
	assert.Equal(t, 0, len(prom.KubernetesPods))
}

func pod() *v1.Pod {
	p := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}
	p.Status.PodIP = "127.0.0.1"
	p.Name = "myPod"
	return p
}
