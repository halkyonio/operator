package component

import (
	"fmt"
	securityv1 "github.com/openshift/api/security/v1"
	"github.com/snowdrop/component-operator/pkg/apis/component/v1alpha2"
	"github.com/snowdrop/component-operator/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type scc struct {
	base
	reconciler *ReconcileComponent // todo: remove
}

func (res scc) NewInstanceWith(owner v1alpha2.Resource) controller.DependentResource {
	return newOwnedScc(res.reconciler, owner)
}

func newScc(reconciler *ReconcileComponent) scc {
	return newOwnedScc(reconciler, nil)
}

func newOwnedScc(reconciler *ReconcileComponent, owner v1alpha2.Resource) scc {
	dependent := newBaseDependent(&securityv1.SecurityContextConstraints{}, owner)
	scc := scc{
		base:       dependent,
		reconciler: reconciler,
	}
	dependent.SetDelegate(scc)
	return scc
}

func (res scc) Name() string {
	return "privileged"
}

func (res scc) Build() (runtime.Object, error) {
	// oc adm policy add-scc-to-user privileged -z build-bot
	// TODO: Fetch SCC Privileged and add to the users's list a new sa if it does not exist. This resource is cluster-wide managed and should not be deleted or owned by us
	ser := &securityv1.SecurityContextConstraints{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "SecurityContextConstraints",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: res.Name(),
		},
		Users: []string{
			"system:serviceaccount:<NAMESPACE>:<Tekton -> SA = build-bot>",
		},
	}
	return ser, nil
}

func (res scc) ShouldWatch() bool {
	return false
}

//  allowHostDirVolumePlugin: true
//  allowHostIPC: true
//  allowHostNetwork: true
//  allowHostPID: true
//  allowHostPorts: true
//  allowPrivilegeEscalation: true
//  allowPrivilegedContainer: true
//  allowedCapabilities:
//    - '*'
//  apiVersion: v1
//  defaultAddCapabilities: null
//  fsGroup:
//    type: RunAsAny
//  groups:
//    - system:cluster-admins
//    - system:nodes
//    - system:masters
//  kind: SecurityContextConstraints
//  metadata:
//    annotations:
//      kubernetes.io/description: 'privileged allows access to all privileged and host
//      features and the ability to run as any user, any group, any fsGroup, and with
//      any SELinux context.  WARNING: this is the most relaxed SCC and should be used
//      only for cluster administration. Grant with caution.'
//    creationTimestamp: "2018-11-26T05:44:02Z"
//    name: privileged
//    resourceVersion: "46030068"
//    selfLink: /api/v1/securitycontextconstraints/privileged
//    uid: 473da8d3-f13e-11e8-9ad2-107b44b03540
//  priority: null
//  readOnlyRootFilesystem: false
//  requiredDropCapabilities: null
//  runAsUser:
//    type: RunAsAny
//  seLinuxContext:
//    type: RunAsAny
//  seccompProfiles:
//    - '*'
//  supplementalGroups:
//    type: RunAsAny
//  users:
//    - system:admin
//    - system:serviceaccount:openshift-infra:build-controller
//    - system:serviceaccount:openshift-node:sync
//    - system:serviceaccount:openshift-sdn:sdn
//    - system:serviceaccount:management-infra:management-admin
//    - system:serviceaccount:management-infra:inspector-admin
//    - system:serviceaccount:istio-system:jaeger
//    - system:serviceaccount:gandrian:default
//    - system:serviceaccount:istio-system:istio-pilot-service-account
//    - system:serviceaccount:gandrian:sa-greeting
//    - system:serviceaccount:openshift-logging:aggregated-logging-fluentd
//    - system:serviceaccount:cmoullia:default
//    - system:serviceaccount:cmoullia:build-bot
//    - system:serviceaccount:demo:build-bot
//    - system:serviceaccount:demo:my-postgres
//    - system:serviceaccount:demo:postgres-db
//    - system:serviceaccount:test1:postgres-db
//    - system:serviceaccount:claprun:build-bot
//    - system:serviceaccount:test2:postgres-db
//    - system:serviceaccount:testclaprun:build-bot
//    - system:serviceaccount:test1:build-bot
//    - system:serviceaccount:test2:build-bot
//    - system:serviceaccount:test:build-bot
//    - system:serviceaccount:test99:build-bot
//    - system:serviceaccount:test:postgres-db
//    - system:serviceaccount:gytis:postgres-db
//    - system:serviceaccount:gytis:build-bot
//    - system:serviceaccount:gytis-test:postgres-db
//    - system:serviceaccount:gytis-test:build-bot
//    - system:serviceaccount:claprun:postgres-db
//    - system:serviceaccount:claprun:built-bot
//  volumes:
//    - '*'
