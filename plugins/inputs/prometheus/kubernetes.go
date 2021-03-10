package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/ericchiang/k8s"
	corev1 "github.com/ericchiang/k8s/apis/core/v1"
	"github.com/ghodss/yaml"
	"github.com/kubernetes/apimachinery/pkg/fields"
	"github.com/kubernetes/apimachinery/pkg/labels"
)

type payload struct {
	eventype string
	pod      *corev1.Pod
}

type podMetadata struct {
	ResourceVersion string `json:"resourceVersion"`
	SelfLink        string `json:"selfLink"`
}

type podResponse struct {
	Kind       string        `json:"kind"`
	ApiVersion string        `json:"apiVersion"`
	Metadata   podMetadata   `json:"metadata"`
	Items      []*corev1.Pod `json:"items,string,omitempty"`
}

const cAdvisorPodListDefaultInterval = 60

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
				if p.isNodeScrapeScope {
					err = p.cAdvisor(ctx, client)
					if err != nil {
						p.Log.Errorf("Unable to monitor pods with node scrape scope: %s", err.Error())
					}
				} else {
					err = p.watch(ctx, client)
					if err != nil {
						p.Log.Errorf("Unable to watch resources: %s", err.Error())
					}
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

func (p *Prometheus) cAdvisor(ctx context.Context, client *k8s.Client) error {
	// Set InsecureSkipVerify for cAdvisor client since Node IP will not be a SAN for the CA cert
	tlsConfig := client.Client.Transport.(*http.Transport).TLSClientConfig
	tlsConfig.InsecureSkipVerify = true

	// The request will be the same each time
	podsUrl := fmt.Sprintf("https://%s:10250/pods", p.NodeIP)
	req, err := http.NewRequest("GET", podsUrl, nil)
	if err != nil {
		return fmt.Errorf("Error when creating request to %s to get pod list: %w", podsUrl, err)
	}
	client.SetHeaders(req.Header)

	// Update right away so code is not waiting the length of the specified scrape interval initially
	err = updateCadvisorPodList(ctx, p, client, req)
	if err != nil {
		return fmt.Errorf("Error initially updating pod list: %w", err)
	}

	scrapeInterval := cAdvisorPodListDefaultInterval
	if p.PodScrapeInterval != 0 {
		scrapeInterval = p.PodScrapeInterval
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(scrapeInterval) * time.Second):
			err := updateCadvisorPodList(ctx, p, client, req)
			if err != nil {
				return fmt.Errorf("Error updating pod list: %w", err)
			}
		}
	}
}

func updateCadvisorPodList(ctx context.Context, p *Prometheus, client *k8s.Client, req *http.Request) error {

	resp, err := client.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Error when making request for pod list: %w", err)
	}

	// If err is nil, still check response code
	if resp.StatusCode != 200 {
		return fmt.Errorf("Error when making request for pod list with status %s", resp.Status)
	}

	defer resp.Body.Close()

	cadvisorPodsResponse := podResponse{}

	// Will have expected type errors for some parts of corev1.Pod struct for some unused fields
	// Instead have nil checks for every used field in case of incorrect decoding
	json.NewDecoder(resp.Body).Decode(&cadvisorPodsResponse)
	pods := cadvisorPodsResponse.Items

	// Updating pod list to be latest cadvisor response
	p.lock.Lock()
	p.kubernetesPods = make(map[string]URLAndAddress)

	// Register pod only if it has an annotation to scrape, if it is ready,
	// and if namespace and selectors are specified and match
	for _, pod := range pods {
		if necessaryPodFieldsArePresent(pod) &&
			pod.GetMetadata().GetAnnotations()["prometheus.io/scrape"] == "true" &&
			podReady(pod.GetStatus().GetContainerStatuses()) &&
			podHasMatchingNamespace(pod, p) &&
			podHasMatchingLabelSelector(pod, p.podLabelSelector) &&
			podHasMatchingFieldSelector(pod, p.podFieldSelector) {
			registerPod(pod, p)
		}

	}
	p.lock.Unlock()

	// No errors
	return nil
}

func necessaryPodFieldsArePresent(pod *corev1.Pod) bool {
	return pod.GetMetadata() != nil &&
		pod.GetMetadata().GetAnnotations() != nil &&
		pod.GetMetadata().GetLabels() != nil &&
		pod.GetSpec() != nil &&
		pod.GetStatus() != nil &&
		pod.GetStatus().GetContainerStatuses() != nil
}

/* See the docs on kubernetes label selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
 */
func podHasMatchingLabelSelector(pod *corev1.Pod, labelSelector labels.Selector) bool {
	if labelSelector == nil {
		return true
	}

	var labelsSet labels.Set = pod.GetMetadata().GetLabels()
	return labelSelector.Matches(labelsSet)
}

/* See ToSelectableFields() for list of fields that are selectable:
 * https://github.com/kubernetes/kubernetes/release-1.20/pkg/registry/core/pod/strategy.go
 * See docs on kubernetes field selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
 */
func podHasMatchingFieldSelector(pod *corev1.Pod, fieldSelector fields.Selector) bool {
	if fieldSelector == nil {
		return true
	}

	podSpec := pod.GetSpec()
	podStatus := pod.GetStatus()

	// Spec and Status shouldn't be nil.
	// Error handling just in case something goes wrong but won't crash telegraf
	if podSpec == nil || podStatus == nil {
		return false
	}

	fieldsSet := make(fields.Set)
	fieldsSet["spec.nodeName"] = podSpec.GetNodeName()
	fieldsSet["spec.restartPolicy"] = podSpec.GetRestartPolicy()
	fieldsSet["spec.schedulerName"] = podSpec.GetSchedulerName()
	fieldsSet["spec.serviceAccountName"] = podSpec.GetServiceAccountName()
	fieldsSet["status.phase"] = podStatus.GetPhase()
	fieldsSet["status.podIP"] = podStatus.GetPodIP()
	fieldsSet["status.nominatedNodeName"] = podStatus.GetNominatedNodeName()

	return fieldSelector.Matches(fieldsSet)
}

/*
 * If a namespace is specified and the pod doesn't have that namespace, return false
 * Else return true
 */
func podHasMatchingNamespace(pod *corev1.Pod, p *Prometheus) bool {
	return !(p.PodNamespace != "" && pod.GetMetadata().GetNamespace() != p.PodNamespace)
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

	// Locks earlier if using cAdvisor calls - makes a new list each time
	// rather than updating and removing from the same list
	if !p.isNodeScrapeScope {
		p.lock.Lock()
		defer p.lock.Unlock()
	}
	p.kubernetesPods[podURL.String()] = URLAndAddress{
		URL:         podURL,
		Address:     URL.Hostname(),
		OriginalURL: URL,
		Tags:        tags,
	}
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
