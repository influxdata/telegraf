package v1beta1

import "github.com/ericchiang/k8s"

func init() {
	k8s.Register("extensions", "v1beta1", "daemonsets", true, &DaemonSet{})
	k8s.Register("extensions", "v1beta1", "deployments", true, &Deployment{})
	k8s.Register("extensions", "v1beta1", "ingresses", true, &Ingress{})
	k8s.Register("extensions", "v1beta1", "networkpolicies", true, &NetworkPolicy{})
	k8s.Register("extensions", "v1beta1", "podsecuritypolicies", false, &PodSecurityPolicy{})
	k8s.Register("extensions", "v1beta1", "replicasets", true, &ReplicaSet{})

	k8s.RegisterList("extensions", "v1beta1", "daemonsets", true, &DaemonSetList{})
	k8s.RegisterList("extensions", "v1beta1", "deployments", true, &DeploymentList{})
	k8s.RegisterList("extensions", "v1beta1", "ingresses", true, &IngressList{})
	k8s.RegisterList("extensions", "v1beta1", "networkpolicies", true, &NetworkPolicyList{})
	k8s.RegisterList("extensions", "v1beta1", "podsecuritypolicies", false, &PodSecurityPolicyList{})
	k8s.RegisterList("extensions", "v1beta1", "replicasets", true, &ReplicaSetList{})
}
