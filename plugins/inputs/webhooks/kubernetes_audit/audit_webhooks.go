package kubernetes_audit

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/influxdata/telegraf"
	"k8s.io/apiserver/pkg/apis/audit"
)

type KubeAuditWebhook struct {
	Path string
	acc  telegraf.Accumulator
}

func (k *KubeAuditWebhook) Register(router *mux.Router, acc telegraf.Accumulator) {
	router.HandleFunc(k.Path, k.eventHandler).Methods("POST")
	log.Printf("I! Started the kubernetest_audit_webhook on %s", k.Path)
	k.acc = acc
}

func (k *KubeAuditWebhook) eventHandler(w http.ResponseWriter, r *http.Request) {
	var audits audit.EventList
	// Read body
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(b, &audits)
	if err != nil {
		http.Error(w, string(b), http.StatusInternalServerError)
		return
	}
	for ii := 0; ii < len(audits.Items); ii++ {
		tags := map[string]string{}
		fields := map[string]interface{}{}
		audit := audits.Items[ii]
		tags["level"] = string(audit.Level)
		tags["stage"] = string(audit.Stage)
		tags["verb"] = string(audit.Verb)
		tags["user_username"] = audit.User.Username
		tags["user_uid"] = audit.User.UID
		if audit.ObjectRef != nil {
			tags["resource_name"] = audit.ObjectRef.Name
			tags["namespace"] = audit.ObjectRef.Namespace
			tags["resource_type"] = audit.ObjectRef.Resource
			tags["resource_version"] = audit.ObjectRef.ResourceVersion
		}

		//"groups": ["system:authenticated"]
		//audit.User.Groups

		// "sourceIPs": ["192.168.99.100"],
		//audit.SourceIPs

		fields["audit_id"] = string(audit.AuditID)
		fields["request_uri"] = string(audit.RequestURI)
		k.acc.AddFields("kubernetes_audit", fields, tags, audit.RequestReceivedTimestamp.Time)
	}
	w.WriteHeader(http.StatusOK)
}
