package v1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("", "v1", "componentstatuses", false, &ComponentStatus{})
	k8s.Register("", "v1", "configmaps", true, &ConfigMap{})
	k8s.Register("", "v1", "endpoints", true, &Endpoints{})
	k8s.Register("", "v1", "limitranges", true, &LimitRange{})
	k8s.Register("", "v1", "namespaces", false, &Namespace{})
	k8s.Register("", "v1", "nodes", false, &Node{})
	k8s.Register("", "v1", "persistentvolumeclaims", true, &PersistentVolumeClaim{})
	k8s.Register("", "v1", "persistentvolumes", false, &PersistentVolume{})
	k8s.Register("", "v1", "pods", true, &Pod{})
	k8s.Register("", "v1", "replicationcontrollers", true, &ReplicationController{})
	k8s.Register("", "v1", "resourcequotas", true, &ResourceQuota{})
	k8s.Register("", "v1", "secrets", true, &Secret{})
	k8s.Register("", "v1", "services", true, &Service{})
	k8s.Register("", "v1", "serviceaccounts", true, &ServiceAccount{})

	k8s.RegisterList("", "v1", "componentstatuses", false, &ComponentStatusList{})
	k8s.RegisterList("", "v1", "configmaps", true, &ConfigMapList{})
	k8s.RegisterList("", "v1", "endpoints", true, &EndpointsList{})
	k8s.RegisterList("", "v1", "limitranges", true, &LimitRangeList{})
	k8s.RegisterList("", "v1", "namespaces", false, &NamespaceList{})
	k8s.RegisterList("", "v1", "nodes", false, &NodeList{})
	k8s.RegisterList("", "v1", "persistentvolumeclaims", true, &PersistentVolumeClaimList{})
	k8s.RegisterList("", "v1", "persistentvolumes", false, &PersistentVolumeList{})
	k8s.RegisterList("", "v1", "pods", true, &PodList{})
	k8s.RegisterList("", "v1", "replicationcontrollers", true, &ReplicationControllerList{})
	k8s.RegisterList("", "v1", "resourcequotas", true, &ResourceQuotaList{})
	k8s.RegisterList("", "v1", "secrets", true, &SecretList{})
	k8s.RegisterList("", "v1", "services", true, &ServiceList{})
	k8s.RegisterList("", "v1", "serviceaccounts", true, &ServiceAccountList{})
}
