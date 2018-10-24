package prometheus

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/ghodss/yaml"
)

// loadClient parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadClient(kubeconfigPath string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("read kubeconfig: %s", err.Error())
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("unmarshal kubeconfig: %s", err.Error())
	}
	return k8s.NewClient(&config)
}

func start(p *Prometheus) error {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		configLocation := fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
		if p.KubeConfig != "" {
			configLocation = p.KubeConfig
		}
		client, err = loadClient(configLocation)
		if err != nil {
			return err
		}
	}
	type payload struct {
		eventype string
		pod      *corev1.Pod
	}

	in := make(chan payload)
	go func() {
		var pod corev1.Pod
	rewatch:
		watcher, err := client.Watch(context.Background(), "", &pod)
		if err != nil {
			log.Printf("E! [inputs.prometheus] unable to watch resources: %s", err.Error())
		}
		defer watcher.Close()

		for {
			select {
			case <-p.done:
				log.Printf("I! [inputs.prometheus] shutting down\n")
				return
			default:
				cm := new(corev1.Pod)
				eventType, err := watcher.Next(cm)
				if err != nil {
					log.Printf("D! [inputs.prometheus] unable to watch next: %s", err.Error())
					goto rewatch
				}
				in <- payload{eventType, cm}
			}
		}
	}()

	go func() {
		for {
			select {
			case <-p.done:
				log.Printf("I! [inputs.prometheus] shutting down\n")
				return
			case payload := <-in:
				cm := payload.pod
				eventType := payload.eventype

				switch eventType {
				case k8s.EventAdded:
					registerPod(cm, p)
				case k8s.EventDeleted:
					unregisterPod(cm, p)
				case k8s.EventModified:
				}
			}
		}
	}()

	return nil
}

func registerPod(pod *corev1.Pod, p *Prometheus) {
	targetURL := getScrapeURL(pod)
	if targetURL == nil {
		return
	}

	log.Printf("I! [inputs.prometheus] will scrape metrics from %v\n", *targetURL)
	// add annotation as metrics tags
	tags := pod.GetMetadata().GetAnnotations()
	tags["pod_name"] = pod.GetMetadata().GetName()
	tags["namespace"] = pod.GetMetadata().GetNamespace()
	// add labels as metrics tags
	for k, v := range pod.GetMetadata().GetLabels() {
		tags[k] = v
	}
	URL, err := url.Parse(*targetURL)
	if err != nil {
		log.Printf("E! [inputs.prometheus] could not parse URL %q: %v", targetURL, err)
		return
	}
	podURL := p.AddressToURL(URL, URL.Hostname())
	p.lock.Lock()
	p.kubernetesPods = append(p.kubernetesPods,
		URLAndAddress{
			URL:         podURL,
			Address:     URL.Hostname(),
			OriginalURL: URL,
			Tags:        tags})
	p.lock.Unlock()
}

func getScrapeURL(pod *corev1.Pod) *string {
	scrape := pod.GetMetadata().GetAnnotations()["prometheus.io/scrape"]
	if scrape != "true" {
		return nil
	}
	ip := pod.Status.GetPodIP()
	if ip == "" {
		// return as if scrape was disabled, we will be notified again once the pod
		// has an IP
		return nil
	}

	path := pod.GetMetadata().GetAnnotations()["prometheus.io/path"]
	port := pod.GetMetadata().GetAnnotations()["prometheus.io/port"]
	if port == "" {
		port = "9102" // default
	}
	if path == "" {
		path = "/metrics"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	x := fmt.Sprintf("http://%s:%s%s", ip, port, path)

	return &x
}

func unregisterPod(pod *corev1.Pod, p *Prometheus) {
	url := getScrapeURL(pod)
	if url == nil {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	log.Printf("D! [inputs.prometheus] registered a delete request for %s in namespace %s\n",
		pod.GetMetadata().GetName(), pod.GetMetadata().GetNamespace())
	var result []URLAndAddress
	for _, v := range p.kubernetesPods {
		if v.URL.String() != *url {
			result = append(result, v)
		} else {
			log.Printf("D! [inputs.prometheus] will stop scraping for %v\n", *url)
		}

	}
	p.kubernetesPods = result
}
