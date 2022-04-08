package prometheus

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type podMetadata struct {
	ResourceVersion string `json:"resourceVersion"`
	SelfLink        string `json:"selfLink"`
}

type podResponse struct {
	Kind       string        `json:"kind"`
	APIVersion string        `json:"apiVersion"`
	Metadata   podMetadata   `json:"metadata"`
	Items      []*corev1.Pod `json:"items,string,omitempty"`
}

const cAdvisorPodListDefaultInterval = 60

// loadClient parses a kubeconfig from a file and returns a Kubernetes
// client. It does not support extensions or client auth providers.
func loadClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading '%s': %v", kubeconfigPath, err)
	}

	// Unmarshal YAML into a Kubernetes config object.
	var config rest.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(&config)
}

func (p *Prometheus) startK8s(ctx context.Context) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get InClusterConfig - %v", err)
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		u, err := user.Current()
		if err != nil {
			return fmt.Errorf("failed to get current user - %v", err)
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

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				if p.isNodeScrapeScope {
					err = p.cAdvisor(ctx, config.BearerToken)
					if err != nil {
						p.Log.Errorf("Unable to monitor pods with node scrape scope: %s", err.Error())
					}
				} else {
					p.watchPod(ctx, client)
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
func (p *Prometheus) watchPod(ctx context.Context, clientset *kubernetes.Clientset) {
	var resyncinterval time.Duration

	if p.CacheRefreshInterval != 0 {
		resyncinterval = time.Duration(p.CacheRefreshInterval) * time.Minute
	} else {
		resyncinterval = 60 * time.Minute
	}

	informerfactory := informers.NewSharedInformerFactory(clientset, resyncinterval)

	podinformer := informerfactory.Core().V1().Pods()
	podinformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				p.Log.Errorf("getting key from cache %s\n", err.Error())
			}

			namespace, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				p.Log.Errorf("splitting key into namespace and name %s\n", err.Error())
			}

			pod, _ := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})

			if pod.Annotations["prometheus.io/scrape"] == "true" &&
				podReady(pod.Status.ContainerStatuses) &&
				podHasMatchingNamespace(pod, p) &&
				podHasMatchingLabelSelector(pod, p.podLabelSelector) &&
				podHasMatchingFieldSelector(pod, p.podFieldSelector) {
				registerPod(pod, p)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newKey, err := cache.MetaNamespaceKeyFunc(newObj)
			if err != nil {
				p.Log.Errorf("getting key from cache %s\n", err.Error())
			}

			newNamespace, newName, err := cache.SplitMetaNamespaceKey(newKey)
			if err != nil {
				p.Log.Errorf("splitting key into namespace and name %s\n", err.Error())
			}

			newPod, _ := clientset.CoreV1().Pods(newNamespace).Get(ctx, newName, metav1.GetOptions{})

			if newPod.Annotations["prometheus.io/scrape"] == "true" &&
				podReady(newPod.Status.ContainerStatuses) &&
				podHasMatchingNamespace(newPod, p) &&
				podHasMatchingLabelSelector(newPod, p.podLabelSelector) &&
				podHasMatchingFieldSelector(newPod, p.podFieldSelector) {
				if newPod.GetDeletionTimestamp() == nil {
					registerPod(newPod, p)
				}
			}

			oldKey, err := cache.MetaNamespaceKeyFunc(oldObj)
			if err != nil {
				p.Log.Errorf("getting key from cache %s\n", err.Error())
			}

			oldNamespace, oldName, err := cache.SplitMetaNamespaceKey(oldKey)
			if err != nil {
				p.Log.Errorf("splitting key into namespace and name %s\n", err.Error())
			}

			oldPod, _ := clientset.CoreV1().Pods(oldNamespace).Get(ctx, oldName, metav1.GetOptions{})

			if oldPod.Annotations["prometheus.io/scrape"] == "true" &&
				podReady(oldPod.Status.ContainerStatuses) &&
				podHasMatchingNamespace(oldPod, p) &&
				podHasMatchingLabelSelector(oldPod, p.podLabelSelector) &&
				podHasMatchingFieldSelector(oldPod, p.podFieldSelector) {
				if oldPod.GetDeletionTimestamp() != nil {
					unregisterPod(oldPod, p)
				}
			}
		},
		DeleteFunc: func(oldObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(oldObj)
			if err != nil {
				p.Log.Errorf("getting key from cache %s", err.Error())
			}

			namespace, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				p.Log.Errorf("splitting key into namespace and name %s\n", err.Error())
			}

			pod, _ := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})

			if pod.Annotations["prometheus.io/scrape"] == "true" &&
				podReady(pod.Status.ContainerStatuses) &&
				podHasMatchingNamespace(pod, p) &&
				podHasMatchingLabelSelector(pod, p.podLabelSelector) &&
				podHasMatchingFieldSelector(pod, p.podFieldSelector) {
				if pod.GetDeletionTimestamp() != nil {
					unregisterPod(pod, p)
				}
			}
		},
	})

	informerfactory.Start(wait.NeverStop)
	informerfactory.WaitForCacheSync(wait.NeverStop)

	<-ctx.Done()
}

func (p *Prometheus) cAdvisor(ctx context.Context, bearerToken string) error {
	// The request will be the same each time
	podsURL := fmt.Sprintf("https://%s:10250/pods", p.NodeIP)
	req, err := http.NewRequest("GET", podsURL, nil)
	if err != nil {
		return fmt.Errorf("error when creating request to %s to get pod list: %w", podsURL, err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Add("Accept", "application/json")

	// Update right away so code is not waiting the length of the specified scrape interval initially
	err = updateCadvisorPodList(p, req)
	if err != nil {
		return fmt.Errorf("error initially updating pod list: %w", err)
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
			err := updateCadvisorPodList(p, req)
			if err != nil {
				return fmt.Errorf("error updating pod list: %w", err)
			}
		}
	}
}

func updateCadvisorPodList(p *Prometheus, req *http.Request) error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := http.Client{}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error when making request for pod list: %w", err)
	}

	// If err is nil, still check response code
	if resp.StatusCode != 200 {
		return fmt.Errorf("error when making request for pod list with status %s", resp.Status)
	}

	defer resp.Body.Close()

	cadvisorPodsResponse := podResponse{}

	// Will have expected type errors for some parts of corev1.Pod struct for some unused fields
	// Instead have nil checks for every used field in case of incorrect decoding
	if err := json.NewDecoder(resp.Body).Decode(&cadvisorPodsResponse); err != nil {
		return fmt.Errorf("decoding response failed: %v", err)
	}
	pods := cadvisorPodsResponse.Items

	// Updating pod list to be latest cadvisor response
	p.lock.Lock()
	p.kubernetesPods = make(map[string]URLAndAddress)

	// Register pod only if it has an annotation to scrape, if it is ready,
	// and if namespace and selectors are specified and match
	for _, pod := range pods {
		if necessaryPodFieldsArePresent(pod) &&
			pod.Annotations["prometheus.io/scrape"] == "true" &&
			podReady(pod.Status.ContainerStatuses) &&
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
	return pod.Annotations != nil &&
		pod.Labels != nil &&
		pod.Status.ContainerStatuses != nil
}

/* See the docs on kubernetes label selectors:
 * https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
 */
func podHasMatchingLabelSelector(pod *corev1.Pod, labelSelector labels.Selector) bool {
	if labelSelector == nil {
		return true
	}

	var labelsSet labels.Set = pod.Labels
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

	fieldsSet := make(fields.Set)
	fieldsSet["spec.nodeName"] = pod.Spec.NodeName
	fieldsSet["spec.restartPolicy"] = string(pod.Spec.RestartPolicy)
	fieldsSet["spec.schedulerName"] = pod.Spec.SchedulerName
	fieldsSet["spec.serviceAccountName"] = pod.Spec.ServiceAccountName
	fieldsSet["status.phase"] = string(pod.Status.Phase)
	fieldsSet["status.podIP"] = pod.Status.PodIP
	fieldsSet["status.nominatedNodeName"] = pod.Status.NominatedNodeName

	return fieldSelector.Matches(fieldsSet)
}

/*
 * If a namespace is specified and the pod doesn't have that namespace, return false
 * Else return true
 */
func podHasMatchingNamespace(pod *corev1.Pod, p *Prometheus) bool {
	return !(p.PodNamespace != "" && pod.Namespace != p.PodNamespace)
}

func podReady(statuss []corev1.ContainerStatus) bool {
	if len(statuss) == 0 {
		return false
	}
	for _, cs := range statuss {
		if !cs.Ready {
			return false
		}
	}
	return true
}

func registerPod(pod *corev1.Pod, p *Prometheus) {
	if p.kubernetesPods == nil {
		p.kubernetesPods = map[string]URLAndAddress{}
	}
	targetURL, err := getScrapeURL(pod)
	if err != nil {
		p.Log.Errorf("could not parse URL: %s", err)
		return
	} else if targetURL == nil {
		return
	}

	p.Log.Debugf("will scrape metrics from %q", targetURL.String())
	// add annotation as metrics tags
	tags := pod.Annotations
	if tags == nil {
		tags = map[string]string{}
	}
	tags["pod_name"] = pod.Name
	tags["namespace"] = pod.Namespace
	// add labels as metrics tags
	for k, v := range pod.Labels {
		tags[k] = v
	}
	podURL := p.AddressToURL(targetURL, targetURL.Hostname())

	// Locks earlier if using cAdvisor calls - makes a new list each time
	// rather than updating and removing from the same list
	if !p.isNodeScrapeScope {
		p.lock.Lock()
		defer p.lock.Unlock()
	}
	p.kubernetesPods[podURL.String()] = URLAndAddress{
		URL:         podURL,
		Address:     targetURL.Hostname(),
		OriginalURL: targetURL,
		Tags:        tags,
	}
}

func getScrapeURL(pod *corev1.Pod) (*url.URL, error) {
	ip := pod.Status.PodIP
	if ip == "" {
		// return as if scrape was disabled, we will be notified again once the pod
		// has an IP
		return nil, nil
	}

	scheme := pod.Annotations["prometheus.io/scheme"]
	pathAndQuery := pod.Annotations["prometheus.io/path"]
	port := pod.Annotations["prometheus.io/port"]

	if scheme == "" {
		scheme = "http"
	}
	if port == "" {
		port = "9102"
	}
	if pathAndQuery == "" {
		pathAndQuery = "/metrics"
	}

	base, err := url.Parse(pathAndQuery)
	if err != nil {
		return nil, err
	}

	base.Scheme = scheme
	base.Host = net.JoinHostPort(ip, port)

	return base, nil
}

func unregisterPod(pod *corev1.Pod, p *Prometheus) {
	targetURL, err := getScrapeURL(pod)
	if err != nil {
		p.Log.Errorf("failed to parse url: %s", err)
		return
	} else if targetURL == nil {
		return
	}

	p.lock.Lock()
	defer p.lock.Unlock()
	if _, ok := p.kubernetesPods[targetURL.String()]; ok {
		p.Log.Debugf("registered a delete request for %q in namespace %q", pod.Name, pod.Namespace)
		delete(p.kubernetesPods, targetURL.String())
		p.Log.Debugf("will stop scraping for %q", targetURL.String())
	}
}
