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
	Value    string `json:"value,omitempty"`
}

func (selectorInfo *selectorInfo) String() string {
	return fmt.Sprintf("{ %s %s}", selectorInfo.Operator, selectorInfo.Value)
}

// "Enum" for Operator of selectorInfo struct
const (
	EqualsOperator      = "="
	EqualEqualsOperator = "=="
	NotEqualsOperator   = "!="
	ExistsOperator      = ""
	NotExistsOperator   = "!"
	InOperator          = "in"
	NotInOperator       = "notin"
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

	// Set InsecureSkipVerify for cAdvisor client since Node IP will not be a SAN for the CA cert
	if p.MonitorPodsVersion == 2 {
		p.Log.Infof("Using monitor pods version 2 to get pod list using cAdvisor.")
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

	// Parse label and field selectors - will be used to filter pods after cAdvisor call
	labelSelectorMap = createSelectorMap(p.KubernetesLabelSelector, p)
	log.Printf("[cAdvisor Changes Log] labelSelectorMap: %v", labelSelectorMap)
	fieldSelectorMap = createSelectorMap(p.KubernetesFieldSelector, p)
	log.Printf("[cAdvisor Changes Log] fieldSelectorMap: %v", fieldSelectorMap)

	// The request will be the same each time
	nodeIP := os.Getenv("NODE_IP")
	podsUrl := fmt.Sprintf("https://%s:10250/pods", nodeIP)
	req, err := http.NewRequest("GET", podsUrl, nil)
	if err != nil {
		return err
	}
	client.SetHeaders(req.Header)

	// Update right away so code is not waiting 60s initially
	err = updateCadvisorPodList(ctx, p, client, req)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(updatePodScrapeListInterval * time.Second):
			err := updateCadvisorPodList(ctx, p, client, req)
			if err != nil {
				return err
			}
		}
	}
}

func updateCadvisorPodList(ctx context.Context, p *Prometheus, client *k8s.Client, req *http.Request) error {
	log.Printf("[cAdvisor Changes Log] Making request for updated pod list")
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

	// Updating pod list to be cadvisor response
	p.lock.Lock()
	log.Printf("[cAdvisor Changes Log] locking for new kubernetes pods list")
	p.kubernetesPods = nil
	for _, pod := range pods {
		if pod.GetMetadata().GetAnnotations()["prometheus.io/scrape"] == "true" {
			log.Printf("[cAdvisor Changes Log] pod %s has scrape annotation. Checking namespace, field, and label selectors.", pod.GetMetadata().GetName())
		}
		// Register pod only if it has an annotation to scrape, if it is ready,
		// and if namespace/selectors are specified and match
		if pod.GetMetadata().GetAnnotations()["prometheus.io/scrape"] == "true" &&
			podReady(pod.Status.GetContainerStatuses()) &&
			podHasMatchingNamespace(pod, p) &&
			podHasMatchingLabelSelector(pod) &&
			podHasMatchingFieldSelector(pod, p) {
				registerPod(pod, p)
		}

	}
	log.Printf("[cAdvisor Changes Log] unlocking for new kubernetes pods list")
	p.lock.Unlock()

	// No errors
	return nil
}

/* See the docs on kubernetes label and field selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
 */
func createSelectorMap(selectors string, p *Prometheus) map[string]*selectorInfo {
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
			if strings.Contains(selectorKeyValue, NotEqualsOperator) {
				addEqualityBasedSelector(NotEqualsOperator, selectorKeyValue, selectorMap, p)
			} else if strings.Contains(selectorKeyValue, EqualEqualsOperator) {
				addEqualityBasedSelector(EqualEqualsOperator, selectorKeyValue, selectorMap, p)
			} else if strings.Contains(selectorKeyValue, EqualsOperator) {
				addEqualityBasedSelector(EqualsOperator, selectorKeyValue, selectorMap, p)

			// Set-based label selectors
			} else if strings.Contains(selectorKeyValue, fmt.Sprintf(" %s", NotInOperator)) {
				labelName := strings.Split(selectorKeyValue, fmt.Sprintf(" %s", NotInOperator))
				if len(labelName) > 0 && len(wantedSets) >= wantedSetsIndex {
					selectorMap[strings.TrimSpace(labelName[0])] = &selectorInfo{
						Operator: NotInOperator,
						Value:    strings.TrimSpace(wantedSets[wantedSetsIndex]),
					}
					wantedSetsIndex++
				} else {
					p.Log.Errorf("Unable to parse label selector.")
				}
			} else if strings.Contains(selectorKeyValue, fmt.Sprintf(" %s", InOperator)) {
				labelName := strings.Split(selectorKeyValue, fmt.Sprintf(" %s", InOperator))
				if len(labelName) > 0 && len(wantedSets) >= wantedSetsIndex {
					selectorMap[strings.TrimSpace(labelName[0])] = &selectorInfo{
						Operator: InOperator,
						Value:    strings.TrimSpace(wantedSets[wantedSetsIndex]),
					}
					wantedSetsIndex++
				} else {
					p.Log.Errorf("Unable to parse label selector.")
				}

			// Existence set-based label selectors
			} else if strings.HasPrefix(selectorKeyValue, NotExistsOperator) {
				selectorMap[strings.TrimSpace(selectorKeyValue[1:])] = &selectorInfo{
					Operator: NotExistsOperator,
				}
			} else if selectorKeyValue != "" {
				selectorMap[selectorKeyValue] = &selectorInfo{
					Operator: ExistsOperator,
				}
			}
		}
	}

	return selectorMap
}

func addEqualityBasedSelector(operator string, selectorKeyValue string, selectorMap map[string]*selectorInfo, p *Prometheus) {
	strings.ReplaceAll(selectorKeyValue, " ", "")
	keyValueParts := strings.Split(selectorKeyValue, operator)
	if len(keyValueParts) > 1 {
		selectorMap[keyValueParts[0]] = &selectorInfo{
			Operator: operator,
			Value:    keyValueParts[1],
		}
	} else {
		p.Log.Errorf("Unable to parse selector %s", selectorKeyValue)
	}
}

/* See the docs on kubernetes label selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
 */
func podHasMatchingLabelSelector(pod *corev1.Pod) bool {
	if len(labelSelectorMap) > 0 {
		actualLabels := pod.GetMetadata().GetLabels()

		for name, selectorInfo := range labelSelectorMap {
			operator := selectorInfo.Operator
			wantedValue := selectorInfo.Value
			actualValue := actualLabels[name]

			if (operator == NotEqualsOperator && strings.EqualFold(actualValue, wantedValue)) ||
				((operator == EqualsOperator || operator == EqualEqualsOperator) && !strings.EqualFold(actualValue, wantedValue)) ||
				(operator == ExistsOperator && actualValue == "") ||
				(operator == NotExistsOperator && actualValue != "") {
				log.Printf("[cAdvisor Changes Log] pod %s does not equality/existence label selector", pod.GetMetadata().GetName())
				return false

			// Wanted value is a set that actual value should or shouldn't be in
			} else if operator == InOperator || operator == NotInOperator {
				charReplacer := strings.NewReplacer("(", "", ")", "", " ", "")
				setValues := strings.Split(charReplacer.Replace(wantedValue), ",")

				set := make(map[string]bool)
				for _, value := range setValues {
					set[value] = true
				}

				if (operator == InOperator && !set[actualValue]) ||
					(operator == NotInOperator && set[actualValue]) {
					log.Printf("[cAdvisor Changes Log] pod %s does not match set-based label selector", pod.GetMetadata().GetName())
					return false
				}
			}
		}
	}

	log.Printf("[cAdvisor Changes Log] pod %s matches label selectors", pod.GetMetadata().GetName())
	return true
}

/* See PodToSelectableFields() for list of fields that are selectable:
 * https://github.com/kubernetes/kubernetes/blob/v1.16.1/pkg/registry/core/pod/strategy.go
 * See docs on kubernetes field selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/
 */
func podHasMatchingFieldSelector(pod *corev1.Pod, p *Prometheus) bool {
	if len(fieldSelectorMap) > 0 {
		for name, selectorInfo := range fieldSelectorMap {
			getField := func() string { return "" }

			// Not all fields can be selected. See link above for the list
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
			default:
				p.Log.Errorf("Unrecognized field selector specified.")
				return false
			}

			operator := selectorInfo.Operator
			wantedValue := selectorInfo.Value
			actualValue := string(getField())

			if (operator == EqualsOperator && !strings.EqualFold(actualValue, wantedValue)) ||
				(operator == NotEqualsOperator && strings.EqualFold(actualValue, wantedValue)) {
				log.Printf("[cAdvisor Changes Log] pod %s does not match field selector %s", pod.GetMetadata().GetName(), name)
				return false
			}

		}
	}

	log.Printf("[cAdvisor Changes Log] pod %s matches field selectors", pod.GetMetadata().GetName())
	return true
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
	if p.MonitorPodsVersion != 2 {
		p.lock.Lock()
	}
	p.kubernetesPods[podURL.String()] = URLAndAddress{
		URL:         podURL,
		Address:     URL.Hostname(),
		OriginalURL: URL,
		Tags:        tags,
	}
	if p.MonitorPodsVersion != 2 {
		p.lock.Unlock()
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
