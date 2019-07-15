#!/usr/bin/env bash

DIR=$(dirname "$0")
NS=${1:-test}

function deleteResources() {
  result=$(kubectl api-resources --verbs=list --namespaced -o name)
  for i in $result[@]
  do
    echo "deleting resource type : $i ..."
    kubectl delete $i --ignore-not-found=true --all -n $1
  done
}

deleteResources $NS
kubectl delete ns ${NS} --wait=false
${DIR}/kill-ns.sh ${NS}
