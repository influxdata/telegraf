package containerapp

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf/internal"
	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type K8s struct {
	nodename       string
	kubeconfig     string
	pods           map[string]*api_v1.Pod
	syncinterval   internal.Duration
	startinterval  internal.Duration
	kubeClient     kubernetes.Interface
	informer       cache.SharedIndexInformer
	stopInformerCh chan struct{}
	Add            func(id string, conf *Config) error
	Del            func(id string)
	Error          func(err error)
}

func NewK8s(
	nodename string,
	kubeconfig string,
	startinterval internal.Duration,
	syncinterval internal.Duration,
	Add func(id string, conf *Config) error,
	Del func(id string),
	Error func(err error),
) (*K8s, error) {

	var kubeClient kubernetes.Interface
	log.Printf("I! k8s get InClusterConfig config")
	_, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("I! k8s get ClientOutOfCluster config")
		kubeClient = GetClientOutOfCluster(kubeconfig)
	} else {
		log.Printf("I! k8s get standart Client config")
		kubeClient = GetClient()
	}
	log.Printf("I! k8s get Client finished")

	pods := map[string]*api_v1.Pod{}

	return &K8s{
		nodename:      nodename,
		kubeconfig:    kubeconfig,
		kubeClient:    kubeClient,
		pods:          pods,
		syncinterval:  syncinterval,
		startinterval: startinterval,
		Add:           Add,
		Del:           Del,
		Error:         Error,
	}, nil
}

// GetClient returns a k8s clientset to the request from inside of cluster
func GetClient() kubernetes.Interface {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Can not create kubernetes client: %v", err)
	}

	return clientset
}

// GetClientOutOfCluster returns a k8s clientset to the request from outside of cluster
func GetClientOutOfCluster(kubeconfig string) kubernetes.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("Can not get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	return clientset
}

func (k8s *K8s) stopInformer() {
	if k8s.stopInformerCh != nil {
		close(k8s.stopInformerCh)
	}
	k8s.stopInformerCh = nil
	k8s.informer = nil
}

func (k8s *K8s) runInformer() {
	fieldSelector := fmt.Sprintf("spec.nodeName=%s", k8s.nodename)

	k8s.informer = cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				options.FieldSelector = fieldSelector
				return k8s.kubeClient.CoreV1().Pods(meta_v1.NamespaceAll).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				options.FieldSelector = fieldSelector
				return k8s.kubeClient.CoreV1().Pods(meta_v1.NamespaceAll).Watch(options)
			},
		},
		&api_v1.Pod{},
		k8s.syncinterval.Duration,
		cache.Indexers{},
	)

	k8s.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*api_v1.Pod)
			if ok {
				if pod.Spec.NodeName == k8s.nodename {
					k8s.add(pod)
				}
			}
		},
		UpdateFunc: func(old, new interface{}) {
			pod, ok := new.(*api_v1.Pod)
			if ok {
				if pod.Spec.NodeName == k8s.nodename {
					k8s.add(pod)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*api_v1.Pod)
			if ok {
				k8s.del(pod)
			}
		},
	})

	log.Printf("I! kubernetes connected, node: %s", k8s.nodename)
	k8s.stopInformerCh = make(chan struct{})
	k8s.informer.Run(k8s.stopInformerCh)
	defer k8s.stopInformer()
	log.Printf("I! kubernetes disconnected")
}

func (k8s *K8s) сreateConf(
	pod *api_v1.Pod,
) (*Config, error) {

	conf := &Config{
		Name: pod.ObjectMeta.Name,
		IP:   pod.Status.PodIP,
		Values: []map[string]string{
			pod.ObjectMeta.Labels, pod.ObjectMeta.Annotations,
		},
		SystemTags: map[string]string{
			"namespace": pod.ObjectMeta.Namespace,
			"pod_name":  pod.ObjectMeta.Name,
			"pod_id":    string(pod.ObjectMeta.UID),
		},
	}

	return conf, nil
}

func (k8s *K8s) add(pod *api_v1.Pod) {
	// skip unready pod
	if len(pod.Status.PodIP) == 0 {
		return
	}

	id := string(pod.ObjectMeta.UID)
	if _, ok := k8s.pods[id]; ok {
		return
	}

	conf, err := k8s.сreateConf(pod)
	if err != nil {
		return
	}

	conf.IP = pod.Status.PodIP
	err = k8s.Add(id, conf)
	if err != nil {
		return
	}

	k8s.pods[id] = pod
}
func (k8s *K8s) del(pod *api_v1.Pod) {
	id := string(pod.ObjectMeta.UID)
	if _, ok := k8s.pods[id]; !ok {
		return
	}
	k8s.Del(id)
	delete(k8s.pods, id)
}
func (k8s *K8s) error(err error) {
	k8s.Error(err)
}

func checkPodID(pods []api_v1.Pod, id string) bool {
	for i := range pods {
		podID := string(pods[i].ObjectMeta.UID)
		if podID == id {
			return true
		}
	}
	return false
}

func (k8s *K8s) syncClinets() error {
	log.Printf("D! syncing clients, node: %s", k8s.nodename)
	options := meta_v1.ListOptions{}
	options.FieldSelector = fmt.Sprintf("spec.nodeName=%s", k8s.nodename)
	pods, err := k8s.kubeClient.CoreV1().Pods(meta_v1.NamespaceAll).List(options)
	if err != nil {
		log.Printf("E! ContainerApp input: %s", err.Error())
		return err
	}

	for _, pod := range pods.Items {
		id := string(pod.ObjectMeta.UID)
		if _, ok := k8s.pods[id]; !ok {
			time.Sleep(k8s.startinterval.Duration)
			k8s.add(&pod)
		}
	}

	for _, pod := range k8s.pods {
		id := string(pod.ObjectMeta.UID)
		if !checkPodID(pods.Items, id) {
			k8s.del(pod)
		}
	}

	return nil
}

func (k8s *K8s) sync() error {
	log.Printf("I! k8s start sync, node: %s", k8s.nodename)
	err := k8s.syncClinets()
	if err != nil {
		log.Printf("E! k8s sync errors: %s", err.Error())
		for _, pod := range k8s.pods {
			k8s.del(pod)
		}
		k8s.stopInformer()
		return err
	}
	log.Printf("I! k8s finish sync, node: %s", k8s.nodename)
	return nil
}

func (k8s *K8s) checkInformer() {
	// start Informer if it was closed or did not exist
	if k8s.stopInformerCh == nil {
		go k8s.runInformer()
		log.Printf("E! k8s informer started")
	}
}

func (k8s *K8s) Run() {

	defer func() {
		log.Printf("E! ContainerApp: k8s connect error, restart")
		time.Sleep(5 * time.Second)
		go k8s.Run()
	}()

	err := k8s.sync()
	if err != nil {
		return
	}

	k8s.checkInformer()

	ticker := time.NewTicker(k8s.syncinterval.Duration)

	for {
		select {
		case <-ticker.C:
			k8s.checkInformer()
		}
	}
}
