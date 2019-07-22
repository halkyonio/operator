#!/usr/bin/env bash

#
# Prerequisite : install tool jq
# This script assumes that KubeDB & Component operators are installed
#
# End to end scenario to be executed on minikube, minishift or k8s cluster
# Example: ./scripts/end-to-end.sh <CLUSTER_IP> <NAMESPACE> <DEPLOYMENT_MODE>
# where CLUSTER_IP represents the external IP address exposed top of the VM
#
DIR=$(dirname "$0")
CLUSTER_IP=${1:-192.168.99.50}
NS=${2:-test}
MODE=${3:-dev}

SLEEP_TIME=30s
TIME=$(date +"%Y-%m-%d_%H-%M")
REPORT_FILE="report.txt"
EXPECTED_RESPONSE='{"status":"UP"}'
EXPECTED_FRUITS='[{"id":1,"name":"Cherry"},{"id":2,"name":"Apple"},{"id":3,"name":"Banana"}]'

# Test if we run on plain k8s or openshift
res=$(kubectl api-versions | grep user.openshift.io/v1)
if [ "$res" == "" ]; then
  isOpenShift="false"
else
  isOpenShift="true"
fi

if [ "$MODE" == "build" ]; then
    COMPONENT_FRUIT_BACKEND_NAME="fruit-backend-sb-build"
    COMPONENT_FRUIT_CLIENT_NAME="fruit-client-sb-build"
else
    COMPONENT_FRUIT_BACKEND_NAME="fruit-backend-sb"
    COMPONENT_FRUIT_CLIENT_NAME="fruit-client-sb"
fi

function printTitle {
  r=$(typeset i=${#1} c="=" s="" ; while ((i)) ; do ((i=i-1)) ; s="$s$c" ; done ; echo  "$s" ;)
  printf "$r\n$1\n$r\n"
}

printTitle "Creating the namespace"
kubectl create ns ${NS}

printTitle "Add privileged SCC to the serviceaccount postgres-db and buildbot. Required for the operators Tekton and KubeDB"
if [ "$isOpenShift" == "true" ]; then
  echo "We run on Openshift. So we will apply the SCC rule"
  oc adm policy add-scc-to-user privileged system:serviceaccount:${NS}:postgres-db
  oc adm policy add-scc-to-user privileged system:serviceaccount:${NS}:build-bot
  oc adm policy add-role-to-user edit system:serviceaccount:${NS}:build-bot
else
  echo "We DON'T run on OpenShift. So need to change SCC"
fi

printTitle "Deploy the component for the fruit-backend, link and capability"
kubectl apply -n ${NS} -f ${DIR}/../fruit-backend-sb/target/classes/META-INF/dekorate/component.yml
echo "Sleep ${SLEEP_TIME}"
sleep ${SLEEP_TIME}

printTitle "Deploy the component for the fruit-client, link"
kubectl apply -n ${NS} -f ${DIR}/../fruit-client-sb/target/classes/META-INF/dekorate/component.yml
echo "Sleep ${SLEEP_TIME}"
sleep ${SLEEP_TIME}

printTitle "Report status : ${TIME}" > ${REPORT_FILE}

printTitle "1. Status of the resources created using the CRDs : Component, Link or Capability" >> ${REPORT_FILE}
if [ "$isOpenShift" == "true" ]; then
  for i in components links capabilities pods deployments deploymentconfigs services routes pvc postgreses secret/postgres-db-config
  do
    printTitle "$(echo $i | tr a-z A-Z)" >> ${REPORT_FILE}
    kubectl get $i -n ${NS} >> ${REPORT_FILE}
    printf "\n" >> ${REPORT_FILE}
  done
else
  for i in components links capabilities pods deployments services ingresses pvc postgreses secret/postgres-db-config
  do
    printTitle "$(echo $i | tr a-z A-Z)" >> ${REPORT_FILE}
    kubectl get $i -n ${NS} >> ${REPORT_FILE}
    printf "\n" >> ${REPORT_FILE}
  done
fi

printTitle "2. ENV injected to the fruit backend component"
printTitle "2. ENV injected to the fruit backend component" >> ${REPORT_FILE}
until kubectl get pods -n $NS -l app=$COMPONENT_FRUIT_BACKEND_NAME | grep "Running"; do sleep 5; done
kubectl exec -n ${NS} $(kubectl get pod -n ${NS} -lapp=$COMPONENT_FRUIT_BACKEND_NAME | grep "Running" | awk '{print $1}') env | grep DB >> ${REPORT_FILE}
printf "\n" >> ${REPORT_FILE}

printTitle "3. ENV var defined for the fruit client component"
printTitle "3. ENV var defined for the fruit client component" >> ${REPORT_FILE}
until kubectl get pods -n $NS -l app=$COMPONENT_FRUIT_CLIENT_NAME | grep "Running"; do sleep 5; done
for item in $(kubectl get pod -n ${NS} -lapp=$COMPONENT_FRUIT_CLIENT_NAME --output=name); do printf "Envs for %s\n" "$item" | grep --color -E '[^/]+$' && kubectl get "$item" -n ${NS} --output=json | jq -r -S '.spec.containers[0].env[] | " \(.name)=\(.value)"' 2>/dev/null; printf "\n"; done >> ${REPORT_FILE}
printf "\n" >> ${REPORT_FILE}

if [ "$MODE" == "dev" ]; then
  printTitle "Push fruit client and backend"
  ${DIR}/k8s_push_start.sh fruit-backend sb ${NS}
  ${DIR}/k8s_push_start.sh fruit-client sb ${NS}
fi

printTitle "Wait until Spring Boot actuator health replies UP for both microservices"
for i in $COMPONENT_FRUIT_BACKEND_NAME $COMPONENT_FRUIT_CLIENT_NAME
do
  HTTP_BODY=""
  until [ "$HTTP_BODY" == "$EXPECTED_RESPONSE" ]; do
    HTTP_RESPONSE=$(kubectl exec -n $NS $(kubectl get pod -n $NS -lapp=$i | grep "Running" | awk '{print $1}') -- curl -L -w "HTTPSTATUS:%{http_code}" -s localhost:8080/actuator/health 2>&1)
    HTTP_BODY=$(echo $HTTP_RESPONSE | sed -e 's/HTTPSTATUS\:.*//g')
    echo "$i: Response is : $HTTP_BODY, expected is : $EXPECTED_RESPONSE"
    sleep 10s
  done
done

printTitle "Curl Fruit service"
printTitle "4. Curl Fruit Endpoint service"  >> ${REPORT_FILE}

if [ "$isOpenShift" == "true" ]; then
    echo "No ingress resources found. We run on OpenShift" >> ${REPORT_FILE}
    FRONTEND_ROUTE_URL=$(kubectl get route/fruit-client-sb -o jsonpath='{.spec.host}' -n ${NS})
    CURL_RESPONSE=$(curl http://$FRONTEND_ROUTE_URL/api/client)
    echo $CURL_RESPONSE >> ${REPORT_FILE}
else
    CURL_RESPONSE=$(curl -H "Host: fruit-client-sb" http://${CLUSTER_IP}/api/client)
    echo $CURL_RESPONSE >> ${REPORT_FILE}
fi

if [ "$CURL_RESPONSE" == "$EXPECTED_FRUITS" ]; then
   printTitle "CALLING FRUIT ENDPOINT SUCCEEDED USING MODE : $MODE :-)"
else
   printTitle "FAILED TO CALL FRUIT ENDPOINT USING MODE : $MODE :-("
   exit 1
fi

printTitle "Delete the resources components, links and capabilities"
if [ "$isOpenShift" == "true" ]; then
  kubectl delete components,links,capabilities,imagestreams --all -n ${NS}
else
  kubectl delete components,links,capabilities --all -n ${NS}
fi
