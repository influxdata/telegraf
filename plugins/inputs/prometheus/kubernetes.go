package prometheus

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"gopkg.in/yaml.v2"
)

type payload struct {
	eventype string
	pod      *corev1.Pod
}

// loadClient parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadClient(kubeconfigPath string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s': %v", kubeconfigPath, err)
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config k8s.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return k8s.NewClient(&config)
}

func (p *Prometheus) start(ctx context.Context) error {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("Failed to get current user - %v", err)
		}
		configLocation := filepath.Join(u.HomeDir, ".kube/config")
		if p.KubeConfig != "" {
			configLocation = p.KubeConfig
		}
		client, err = loadClient(configLocation)
		if err != nil {
			return err
		}
	}

	p.wg = sync.WaitGroup{}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				err := p.watch(ctx, client)
				if err != nil {
					log.Printf("E! [inputs.prometheus] unable to watch resources: %v", err)
				}
			}
		}
	}()

	return nil
}

func (p *Prometheus) watch(ctx context.Context, client *k8s.Client) error {
	pod := &corev1.Pod{}
	watcher, err := client.Watch(ctx, "", &corev1.Pod{})
	if err != nil {
		return err
	}
	defer watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			pod = &corev1.Pod{}
			// An error here means we need to reconnect the watcher.
			eventType, err := watcher.Next(pod)
			if err != nil {
				return err
			}

			switch eventType {
			case k8s.EventAdded:
				registerPod(pod, p)
			case k8s.EventDeleted:
				unregisterPod(pod, p)
			case k8s.EventModified:
			}
		}
	}
}

func registerPod(pod *corev1.Pod, p *Prometheus) {
	targetURL := getScrapeURL(pod)
	if targetURL == nil {
		return
	}

	log.Printf("D! [inputs.prometheus] will scrape metrics from %s", *targetURL)
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
		log.Printf("E! [inputs.prometheus] could not parse URL %s: %v", *targetURL, err)
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

	scheme := pod.GetMetadata().GetAnnotations()["prometheus.io/scheme"]
	path := pod.GetMetadata().GetAnnotations()["prometheus.io/path"]
	port := pod.GetMetadata().GetAnnotations()["prometheus.io/port"]

	if scheme == "" {
		scheme = "http"
	}
	if port == "" {
		port = "9102"
	}
	if path == "" {
		path = "/metrics"
	}

	u := &url.URL{
		Scheme: scheme,
		Host:   net.JoinHostPort(ip, port),
		Path:   path,
	}

	x := u.String()

	return &x
}

func unregisterPod(pod *corev1.Pod, p *Prometheus) {
	url := getScrapeURL(pod)
	if url == nil {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	log.Printf("D! [inputs.prometheus] registered a delete request for %s in namespace %s",
		pod.GetMetadata().GetName(), pod.GetMetadata().GetNamespace())
	var result []URLAndAddress
	for _, v := range p.kubernetesPods {
		if v.URL.String() != *url {
			result = append(result, v)
		} else {
			log.Printf("D! [inputs.prometheus] will stop scraping for %s", *url)
		}

	}
	p.kubernetesPods = result
}
