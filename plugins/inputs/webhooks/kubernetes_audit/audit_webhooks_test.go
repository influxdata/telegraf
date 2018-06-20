package kubernetes_audit

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

var fakeWebhook = `{
    "kind": "EventList",
    "apiVersion": "audit.k8s.io/v1beta1",
    "metadata": {},
    "items": [
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:17Z",
            "auditID": "bdffd78c-91e3-40f9-8226-912760d95a64",
            "stage": "RequestReceived",
            "requestURI": "/api/v1/namespaces/kube-system/endpoints/kube-scheduler",
            "verb": "get",
            "user": {
                "username": "system:kube-scheduler",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "endpoints",
                "namespace": "kube-system",
                "name": "kube-scheduler",
                "apiVersion": "v1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:17.415464Z",
            "stageTimestamp": "2018-06-20T11:16:17.415464Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:17Z",
            "auditID": "bdffd78c-91e3-40f9-8226-912760d95a64",
            "stage": "ResponseComplete",
            "requestURI": "/api/v1/namespaces/kube-system/endpoints/kube-scheduler",
            "verb": "get",
            "user": {
                "username": "system:kube-scheduler",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "endpoints",
                "namespace": "kube-system",
                "name": "kube-scheduler",
                "apiVersion": "v1"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:17.415464Z",
            "stageTimestamp": "2018-06-20T11:16:17.418944Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:17Z",
            "auditID": "f34fbe4c-ffc6-4074-94bf-961d3273fb5d",
            "stage": "RequestReceived",
            "requestURI": "/api/v1/namespaces/kube-system/endpoints/kube-scheduler",
            "verb": "update",
            "user": {
                "username": "system:kube-scheduler",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "endpoints",
                "namespace": "kube-system",
                "name": "kube-scheduler",
                "apiVersion": "v1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:17.421414Z",
            "stageTimestamp": "2018-06-20T11:16:17.421414Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:17Z",
            "auditID": "f34fbe4c-ffc6-4074-94bf-961d3273fb5d",
            "stage": "ResponseComplete",
            "requestURI": "/api/v1/namespaces/kube-system/endpoints/kube-scheduler",
            "verb": "update",
            "user": {
                "username": "system:kube-scheduler",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "endpoints",
                "namespace": "kube-system",
                "name": "kube-scheduler",
                "uid": "93eab84f-746c-11e8-a34b-080027191022",
                "apiVersion": "v1",
                "resourceVersion": "7661"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:17.421414Z",
            "stageTimestamp": "2018-06-20T11:16:17.429819Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:10:26Z",
            "auditID": "b9f2b2cc-49e7-471c-85fa-fdf87088ecb7",
            "stage": "ResponseComplete",
            "requestURI": "/apis/admissionregistration.k8s.io/v1beta1/mutatingwebhookconfigurations?resourceVersion=1256&timeoutSeconds=351&watch=true",
            "verb": "watch",
            "user": {
                "username": "system:kube-controller-manager",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "mutatingwebhookconfigurations",
                "apiGroup": "admissionregistration.k8s.io",
                "apiVersion": "v1beta1"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:10:26.948720Z",
            "stageTimestamp": "2018-06-20T11:16:17.949180Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:17Z",
            "auditID": "7d498301-476f-4569-ad0e-e908d22380b7",
            "stage": "RequestReceived",
            "requestURI": "/apis/admissionregistration.k8s.io/v1beta1/mutatingwebhookconfigurations?resourceVersion=1256&timeoutSeconds=423&watch=true",
            "verb": "watch",
            "user": {
                "username": "system:kube-controller-manager",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "mutatingwebhookconfigurations",
                "apiGroup": "admissionregistration.k8s.io",
                "apiVersion": "v1beta1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:17.949885Z",
            "stageTimestamp": "2018-06-20T11:16:17.949885Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:17Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:17Z",
            "auditID": "7d498301-476f-4569-ad0e-e908d22380b7",
            "stage": "ResponseStarted",
            "requestURI": "/apis/admissionregistration.k8s.io/v1beta1/mutatingwebhookconfigurations?resourceVersion=1256&timeoutSeconds=423&watch=true",
            "verb": "watch",
            "user": {
                "username": "system:kube-controller-manager",
                "groups": [
                    "system:authenticated"
                ]
            },
            "sourceIPs": [
                "192.168.99.100"
            ],
            "objectRef": {
                "resource": "mutatingwebhookconfigurations",
                "apiGroup": "admissionregistration.k8s.io",
                "apiVersion": "v1beta1"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:17.949885Z",
            "stageTimestamp": "2018-06-20T11:16:17.950127Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "a305ebd8-0705-42da-9130-7d39c44aeb8b",
            "stage": "RequestReceived",
            "requestURI": "/apis/admissionregistration.k8s.io/v1alpha1/initializerconfigurations",
            "verb": "list",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "initializerconfigurations",
                "apiGroup": "admissionregistration.k8s.io",
                "apiVersion": "v1alpha1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.094223Z",
            "stageTimestamp": "2018-06-20T11:16:18.094223Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "a305ebd8-0705-42da-9130-7d39c44aeb8b",
            "stage": "ResponseComplete",
            "requestURI": "/apis/admissionregistration.k8s.io/v1alpha1/initializerconfigurations",
            "verb": "list",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "initializerconfigurations",
                "apiGroup": "admissionregistration.k8s.io",
                "apiVersion": "v1alpha1"
            },
            "responseStatus": {
                "kind": "Status",
                "apiVersion": "v1",
                "metadata": {},
                "status": "Failure",
                "message": "the server could not find the requested resource",
                "reason": "NotFound",
                "details": {},
                "code": 404
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.094223Z",
            "stageTimestamp": "2018-06-20T11:16:18.094490Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "a9ebfc7b-3fc7-486b-b1ec-4dd0adebba6b",
            "stage": "RequestReceived",
            "requestURI": "/api/v1/namespaces/default",
            "verb": "get",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "namespaces",
                "namespace": "default",
                "name": "default",
                "apiVersion": "v1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.239925Z",
            "stageTimestamp": "2018-06-20T11:16:18.239925Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "a9ebfc7b-3fc7-486b-b1ec-4dd0adebba6b",
            "stage": "ResponseComplete",
            "requestURI": "/api/v1/namespaces/default",
            "verb": "get",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "namespaces",
                "namespace": "default",
                "name": "default",
                "apiVersion": "v1"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.239925Z",
            "stageTimestamp": "2018-06-20T11:16:18.245890Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "0e878cf4-2ec6-409b-9713-13090a7b460f",
            "stage": "RequestReceived",
            "requestURI": "/api/v1/namespaces/default/services/kubernetes",
            "verb": "get",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "services",
                "namespace": "default",
                "name": "kubernetes",
                "apiVersion": "v1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.247179Z",
            "stageTimestamp": "2018-06-20T11:16:18.247179Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "0e878cf4-2ec6-409b-9713-13090a7b460f",
            "stage": "ResponseComplete",
            "requestURI": "/api/v1/namespaces/default/services/kubernetes",
            "verb": "get",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "services",
                "namespace": "default",
                "name": "kubernetes",
                "apiVersion": "v1"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.247179Z",
            "stageTimestamp": "2018-06-20T11:16:18.251173Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "6bdff7de-3c9d-47fc-ba7a-41c5a6341836",
            "stage": "RequestReceived",
            "requestURI": "/api/v1/namespaces/default/endpoints/kubernetes",
            "verb": "get",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "endpoints",
                "namespace": "default",
                "name": "kubernetes",
                "apiVersion": "v1"
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.252898Z",
            "stageTimestamp": "2018-06-20T11:16:18.252898Z"
        },
        {
            "metadata": {
                "creationTimestamp": "2018-06-20T11:16:18Z"
            },
            "level": "Metadata",
            "timestamp": "2018-06-20T11:16:18Z",
            "auditID": "6bdff7de-3c9d-47fc-ba7a-41c5a6341836",
            "stage": "ResponseComplete",
            "requestURI": "/api/v1/namespaces/default/endpoints/kubernetes",
            "verb": "get",
            "user": {
                "username": "system:apiserver",
                "uid": "069ab130-1661-4682-a231-b6e0c11b10ea",
                "groups": [
                    "system:masters"
                ]
            },
            "sourceIPs": [
                "127.0.0.1"
            ],
            "objectRef": {
                "resource": "endpoints",
                "namespace": "default",
                "name": "kubernetes",
                "apiVersion": "v1"
            },
            "responseStatus": {
                "metadata": {},
                "code": 200
            },
            "requestReceivedTimestamp": "2018-06-20T11:16:18.252898Z",
            "stageTimestamp": "2018-06-20T11:16:18.258116Z"
        }
	]
}`

func postWebhooks(md *KubeAuditWebhook, eventBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", "/kubernestes-audit", strings.NewReader(eventBody))
	w := httptest.NewRecorder()

	md.eventHandler(w, req)

	return w
}

func TestBadWebHookFormat(t *testing.T) {
	fixture := map[string]string{
		"random-string": "agrer",
		"empty-string":  "",
	}
	for k, f := range fixture {
		t.Run(k, func(t *testing.T) {
			k := &KubeAuditWebhook{Path: "/kubernetes-audit"}
			resp := postWebhooks(k, f)
			if resp.Code != http.StatusInternalServerError {
				t.Errorf("Expected internal server error but we got status code %d.", resp.Code)
			}
		})
	}
}

func TestBorderlineWebHookFormat(t *testing.T) {
	fixture := map[string]string{
		"empty-json":     "{}",
		"empity-webhook": "{\"items\":[]}",
	}
	for k, f := range fixture {
		t.Run(k, func(t *testing.T) {
			k := &KubeAuditWebhook{Path: "/kubernetes-audit"}
			resp := postWebhooks(k, f)
			if resp.Code != http.StatusOK {
				t.Errorf("Expected status code %d error but we got status code %d.", http.StatusOK, resp.Code)
			}
		})
	}
}

func TestFakeWebhook(t *testing.T) {
	var acc testutil.Accumulator
	k := &KubeAuditWebhook{
		Path: "/kubernetes-audit",
		acc:  &acc,
	}
	resp := postWebhooks(k, fakeWebhook)
	b, _ := ioutil.ReadAll(resp.Body)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Err: %s", http.StatusOK, resp.Code, b)
	}
	if len(acc.Metrics) != 15 {
		t.Errorf("Expected 15 pints inside the accumulator but we got %d.", len(acc.Metrics))
	}
}
