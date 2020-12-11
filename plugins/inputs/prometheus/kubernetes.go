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
	"github.com/ghodss/yaml"
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
					p.Log.Errorf("Unable to watch resources: %s", err.Error())
				}
			}
		}
	}()

	return nil
}

// An edge case exists if a pod goes offline at the same time a new pod is created
// (without the scrape annotations). K8s may re-assign the old pod ip to the non-scrape
// pod, causing errors in the logs. This is only true if the pod going offline is not
// directed to do so by K8s.
func (p *Prometheus) watch(ctx context.Context, client *k8s.Client) error {

	selectors := podSelector(p)

	pod := &corev1.Pod{}
	watcher, err := client.Watch(ctx, p.PodNamespace, &corev1.Pod{}, selectors...)
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

			// If the pod is not "ready", there will be no ip associated with it.
			if pod.GetMetadata().GetAnnotations()["prometheus.io/scrape"] != "true" ||
				!podReady(pod.Status.GetContainerStatuses()) {
				continue
			}

			switch eventType {
			case k8s.EventAdded:
				registerPod(pod, p)
			case k8s.EventModified:
				// To avoid multiple actions for each event, unregister on the first event
				// in the delete sequence, when the containers are still "ready".
				if pod.Metadata.GetDeletionTimestamp() != nil {
					unregisterPod(pod, p)
				} else {
					registerPod(pod, p)
				}
			}
		}
	}
}

func podReady(statuss []*corev1.ContainerStatus) bool {
	if len(statuss) == 0 {
		return false
	}
	for _, cs := range statuss {
		if !cs.GetReady() {
			return false
		}
	}
	return true
}

func podSelector(p *Prometheus) []k8s.Option {
	options := []k8s.Option{}

	if len(p.KubernetesLabelSelector) > 0 {
		options = append(options, k8s.QueryParam("labelSelector", p.KubernetesLabelSelector))
	}

	if len(p.KubernetesFieldSelector) > 0 {
		options = append(options, k8s.QueryParam("fieldSelector", p.KubernetesFieldSelector))
	}

	return options

}

func registerPod(pod *corev1.Pod, p *Prometheus) {
	if p.kubernetesPods == nil {
		p.kubernetesPods = map[string]URLAndAddress{}
	}
	targetURL := getScrapeURL(pod)
	if targetURL == nil {
		return
	}

	log.Printf("D! [inputs.prometheus] will scrape metrics from %q", *targetURL)
	// add annotation as metrics tags
	tags := pod.GetMetadata().GetAnnotations()
	if tags == nil {
		tags = map[string]string{}
	}
	tags["pod_name"] = pod.GetMetadata().GetName()
	tags["namespace"] = pod.GetMetadata().GetNamespace()
	// add labels as metrics tags
	for k, v := range pod.GetMetadata().GetLabels() {
		tags[k] = v
	}
	URL, err := url.Parse(*targetURL)
	if err != nil {
		log.Printf("E! [inputs.prometheus] could not parse URL %q: %s", *targetURL, err.Error())
		return
	}
	podURL := p.AddressToURL(URL, URL.Hostname())
	p.lock.Lock()
	p.kubernetesPods[podURL.String()] = URLAndAddress{
		URL:         podURL,
		Address:     URL.Hostname(),
		OriginalURL: URL,
		Tags:        tags,
	}
	p.lock.Unlock()
}

func getScrapeURL(pod *corev1.Pod) *string {
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

	log.Printf("D! [inputs.prometheus] registered a delete request for %q in namespace %q",
		pod.GetMetadata().GetName(), pod.GetMetadata().GetNamespace())

	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.kubernetesPods[*url]; ok {
		delete(p.kubernetesPods, *url)
		log.Printf("D! [inputs.prometheus] will stop scraping for %q", *url)
	}
}
