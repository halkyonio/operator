

oc delete project my-spring-app && oc project default
oc new-project my-spring-app
OPERATOR_NAME=component-operator WATCH_NAMESPACE=my-spring-app KUBERNETES_CONFIG=$HOME/.kube/config go run cmd/component-operator/main.go

rm -rf {.idea,*}
sd create
sd catalog create

