package v1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("apps", "v1", "controllerrevisions", true, &ControllerRevision{})
	k8s.Register("apps", "v1", "daemonsets", true, &DaemonSet{})
	k8s.Register("apps", "v1", "deployments", true, &Deployment{})
	k8s.Register("apps", "v1", "replicasets", true, &ReplicaSet{})
	k8s.Register("apps", "v1", "statefulsets", true, &StatefulSet{})

	k8s.RegisterList("apps", "v1", "controllerrevisions", true, &ControllerRevisionList{})
	k8s.RegisterList("apps", "v1", "daemonsets", true, &DaemonSetList{})
	k8s.RegisterList("apps", "v1", "deployments", true, &DeploymentList{})
	k8s.RegisterList("apps", "v1", "replicasets", true, &ReplicaSetList{})
	k8s.RegisterList("apps", "v1", "statefulsets", true, &StatefulSetList{})
}
