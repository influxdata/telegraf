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
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
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

type selectorInfo struct {
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// "Enum" for Operator of selectorInfo struct
const (
	Equals      = "="
	EqualEquals = "=="
	NotEquals   = "!="
	Exists      = ""
	NotExists   = "!"
	In          = "in"
	NotIn       = "notin"
)

const updatePodScrapeListInterval = 60

var labelSelectorMap map[string]*selectorInfo
var fieldSelectorMap map[string]*selectorInfo

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

	if p.MonitorPodsVersion == 2 {
		p.Log.Infof("Using monitor pods version 2 to get pod list using cAdvisor.")
		// Set InsecureSkipVerify for cAdvisor client: Node IP will not be a SAN for the CA cert
		tlsConfig := client.Client.Transport.(*http.Transport).TLSClientConfig
		tlsConfig.InsecureSkipVerify = true
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
				if p.MonitorPodsVersion == 2 {
					err = p.cAdvisor(ctx, client)
				} else {
					err = p.watch(ctx, client)
				}
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

func (p *Prometheus) cAdvisor(ctx context.Context, client *k8s.Client) error {
	log.Printf("Grace-log: in p.watch")
	// HTTP request is the same each time
	nodeIP := os.Getenv("NODE_IP")
	podsUrl := fmt.Sprintf("https://%s:10250/pods", nodeIP)
	req, err := http.NewRequest("GET", podsUrl, nil)
	if err != nil {
		return err
	}
	client.SetHeaders(req.Header)

	// Parse label and field selectors
	labelSelectorMap = createSelectorMap(p.KubernetesLabelSelector)
	fieldSelectorMap = createSelectorMap(p.KubernetesFieldSelector)
	log.Printf("label selector map: %v", labelSelectorMap)
	log.Printf("field selector map: %v", fieldSelectorMap)

	// Update at beginning of watch() so code is not waiting 60s initially
	err = updatePodList(ctx, p, client, req)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(updatePodScrapeListInterval * time.Second):
			err := updatePodList(ctx, p, client, req)
			if err != nil {
				return err
			}
		}
	}
}

func updatePodList(ctx context.Context, p *Prometheus, client *k8s.Client, req *http.Request) error {
	log.Printf("Grace-log: after 60s attempting to update url list")

	resp, err := client.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	cadvisorPodsResponse := podResponse{}
	json.Unmarshal([]byte(responseBody), &cadvisorPodsResponse)
	pods := cadvisorPodsResponse.Items

	p.lock.Lock()
	log.Printf("Grace-log: In watch - locking to update the URLAndAddress map")

	p.kubernetesPods = nil
	for _, pod := range pods {
		log.Printf("pod: %s", pod.GetMetadata().GetName())

		if pod.GetMetadata().GetAnnotations()["prometheus.io/scrape"] != "true" ||
			!podReady(pod.Status.GetContainerStatuses()) ||
			(p.PodNamespace != "" && pod.GetMetadata().GetNamespace() != p.PodNamespace) ||
			pod.Metadata.GetDeletionTimestamp() != nil ||
			!podHasMatchingLabelSelector(pod) ||
			!podHasMatchingFieldSelector(pod) {
			continue
		}

		registerPod(pod, p)
	}

	log.Printf("Grace-log: In watch - Unlocking after updating the URLAndAddress map")
	p.lock.Unlock()

	return nil
}

/* See the docs on kubernetes label and field selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
 */
func createSelectorMap(selectors string) map[string]*selectorInfo {
	selectorMap := make(map[string]*selectorInfo)
	if selectors != "" {
		selectors = strings.TrimSpace(selectors)

		re := regexp.MustCompile("\\(.*?\\)|(,)")
		wantedSelectors := re.Split(selectors, -1)

		re = regexp.MustCompile("\\(.*?\\)")
		wantedSets := re.FindAllString(selectors, -1)
		wantedSetsIndex := 0

		for _, selectorKeyValue := range wantedSelectors {
			selectorKeyValue = strings.TrimSpace(selectorKeyValue)

			// Equality-based label and field selectors
			if strings.Contains(selectorKeyValue, NotEquals) {
				strings.ReplaceAll(selectorKeyValue, " ", "")
				parts := strings.Split(selectorKeyValue, NotEquals)
				if len(parts) > 1 {
					selectorMap[strings.TrimSpace(parts[0])] = &selectorInfo{
						Operator: NotEquals,
						Value:    strings.TrimSpace(parts[1]),
					}
				}
			} else if strings.Contains(selectorKeyValue, EqualEquals) {
				strings.ReplaceAll(selectorKeyValue, " ", "")
				parts := strings.Split(selectorKeyValue, EqualEquals)
				if len(parts) > 1 {
					selectorMap[strings.TrimSpace(parts[0])] = &selectorInfo{
						Operator: Equals,
						Value:    strings.TrimSpace(parts[1]),
					}
				}
			} else if strings.Contains(selectorKeyValue, Equals) {
				strings.ReplaceAll(selectorKeyValue, " ", "")
				parts := strings.Split(selectorKeyValue, Equals)
				if len(parts) > 1 {
					selectorMap[strings.TrimSpace(parts[0])] = &selectorInfo{
						Operator: Equals,
						Value:    strings.TrimSpace(parts[1]),
					}
				}

				// Set-based label selectors
			} else if strings.Contains(selectorKeyValue, fmt.Sprintf(" %s", NotIn)) {
				parts := strings.Split(selectorKeyValue, fmt.Sprintf(" %s", NotIn))
				if len(parts) > 0 && len(wantedSets) >= wantedSetsIndex {
					selectorMap[strings.TrimSpace(parts[0])] = &selectorInfo{
						Operator: NotIn,
						Value:    strings.TrimSpace(wantedSets[wantedSetsIndex]),
					}
					wantedSetsIndex++
				}
			} else if strings.Contains(selectorKeyValue, fmt.Sprintf(" %s", In)) {
				parts := strings.Split(selectorKeyValue, fmt.Sprintf(" %s", In))
				if len(parts) > 0 && len(wantedSets) >= wantedSetsIndex {
					selectorMap[strings.TrimSpace(parts[0])] = &selectorInfo{
						Operator: In,
						Value:    strings.TrimSpace(wantedSets[wantedSetsIndex]),
					}
					wantedSetsIndex++
				}
			} else if strings.HasPrefix(selectorKeyValue, NotExists) {
				selectorMap[strings.TrimSpace(selectorKeyValue[1:])] = &selectorInfo{
					Operator: NotExists,
					Value:    "",
				}
			} else if selectorKeyValue != "" {
				selectorMap[selectorKeyValue] = &selectorInfo{
					Operator: Exists,
					Value:    "",
				}
			}
		}
	}

	return selectorMap
}

/* See the docs on kubernetes label selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
 */
func podHasMatchingLabelSelector(pod *corev1.Pod) bool {
	if len(labelSelectorMap) > 0 {
		actualLabels := pod.GetMetadata().GetLabels()
		log.Printf("actual Labels: %v", actualLabels)
		for name, selectorInfo := range labelSelectorMap {
			Operator := selectorInfo.Operator
			value := selectorInfo.Value
			actualValue := actualLabels[name]
			if (Operator == Equals && !strings.EqualFold(actualValue, value)) ||
				(Operator == NotEquals && strings.EqualFold(actualValue, value)) ||
				(Operator == Exists && actualValue == "") ||
				(Operator == NotExists && actualValue != "") {
				return false
			} else if Operator == In || Operator == NotIn {
				setParts := strings.Split(strings.ReplaceAll(strings.ReplaceAll(value, "(", ""), ")", ""), ",")
				set := make(map[string]bool)
				for _, setPart := range setParts {
					setPart := strings.TrimSpace(setPart)
					set[setPart] = true
				}
				if (Operator == In && !set[actualValue]) ||
					(Operator == NotIn && set[actualValue]) {
					return false
				}
			}
		}
	}

	return true
}

/* See PodToSelectableFields() for list of fields that are selectable:
 * https://github.com/kubernetes/kubernetes/blob/v1.16.1/pkg/registry/core/pod/strategy.go
 * See docs on kubernetes field selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
 */
func podHasMatchingFieldSelector(pod *corev1.Pod) bool {
	if len(fieldSelectorMap) > 0 {
		for name, selectorInfo := range fieldSelectorMap {
			getField := func() string { return "" }
			switch strings.ToLower(name) {
			case "spec.nodename":
				getField = pod.GetSpec().GetNodeName
			case "spec.restartpolicy":
				getField = pod.GetSpec().GetRestartPolicy
			case "spec.schedulername":
				getField = pod.GetSpec().GetSchedulerName
			case "spec.serviceaccountname":
				getField = pod.GetSpec().GetServiceAccountName
			case "status.phase":
				getField = pod.GetStatus().GetPhase
			case "status.podip":
				getField = pod.GetStatus().GetPodIP
			case "status.nominatednodename":
				getField = pod.GetStatus().GetNominatedNodeName
			}
			Operator := selectorInfo.Operator
			value := selectorInfo.Value
			actualValue := string(getField())
			if (Operator == Equals && !strings.EqualFold(actualValue, value)) ||
				(Operator == NotEquals && strings.EqualFold(actualValue, value)) {
				return false
			}
		}
	}

	return true
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
