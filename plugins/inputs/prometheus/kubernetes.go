package prometheus

import (
	"fmt"
	"log"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
)

func start(p *Prometheus) error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	watchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				registerPod(pod, p)
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				unregisterPod(pod, p)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				podPod := oldObj.(*v1.Pod)
				newPod := newObj.(*v1.Pod)
				unregisterPod(podPod, p)
				registerPod(newPod, p)
			},
		},
	)

	go controller.Run(wait.NeverStop)
	return nil
}

func registerPod(pod *v1.Pod, p *Prometheus) {
	url := scrapeURL(pod)
	if url != nil {
		log.Printf("Will scrape metrics from %v\n", *url)
		p.lock.Lock()
		// add annotation as metrics tags
		tags := pod.GetAnnotations()
		tags["pod_name"] = pod.Name
		tags["namespace"] = pod.Namespace
		// add labels as metrics tags
		for k, v := range pod.GetLabels() {
			tags[k] = v
		}
		p.KubernetesPods = append(p.KubernetesPods, Target{url: *url, tags: tags})
		p.lock.Unlock()
	}
}

func scrapeURL(pod *v1.Pod) *string {
	scrape := pod.ObjectMeta.Annotations["prometheus.io/scrape"]
	if pod.Status.PodIP == "" {
		// return as if scrape was disabled, we will be notified again once the pod
		// has an IP
		return nil
	}
	if scrape == "true" {
		path := pod.ObjectMeta.Annotations["prometheus.io/path"]
		port := pod.ObjectMeta.Annotations["prometheus.io/port"]
		if port == "" {
			port = "9102" // default
		}
		if path == "" {
			path = "/metrics"
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}

		ip := pod.Status.PodIP
		x := fmt.Sprintf("http://%v:%v%v", ip, port, path)
		return &x
	}
	return nil
}

func unregisterPod(pod *v1.Pod, p *Prometheus) {
	url := scrapeURL(pod)
	if url != nil {
		p.lock.Lock()
		defer p.lock.Unlock()
		log.Printf("Registred a delete request for %v in namespace '%v'\n", pod.Name, pod.Namespace)
		var result []Target
		for _, v := range p.KubernetesPods {
			if v.url != *url {
				result = append(result, v)
			} else {
				log.Printf("Will stop scraping for %v\n", *url)
			}

		}
		p.KubernetesPods = result
	}
}
