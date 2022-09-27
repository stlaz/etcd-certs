module github.com/stlaz/etcd-certs

go 1.19

require (
	github.com/openshift/microshift v0.0.0-20220926090322-0dfd0bbc62ea
	k8s.io/apiserver v0.25.2
)

// for the version that contains the combined certs generation
replace github.com/openshift/microshift => github.com/stlaz/microshift v0.0.0-20220927101038-bdd95484d0a8

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/openshift/library-go v0.0.0-20220615161831-8b2df431789c // indirect
	k8s.io/apimachinery v0.25.2 // indirect
	k8s.io/client-go v0.25.2 // indirect
	k8s.io/klog/v2 v2.70.1 // indirect
	k8s.io/utils v0.0.0-20220728103510-ee6ede2d64ed // indirect
)
