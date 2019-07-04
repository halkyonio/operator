#!/usr/bin/env bash

set -x

component=$1
runtime=$2
project=$component-$runtime
namespace=${3:-my-spring-app}

pod_id=$(kubectl get pod -lapp=${project} -n ${namespace} | grep "Running" | awk '{print $1}')

echo "## $runtime files ${project} pushed ..."

if [ $runtime = "nodejs" ]; then
  cmd="run-node"
  kubectl rsync demo/$project/ $pod_id:/opt/app-root/src/ --no-perms=true -n ${namespace}
else
  cmd="run-java"
  kubectl cp demo/${project}/target/${project}-0.0.1-SNAPSHOT.jar $pod_id:/deployments/${project}-0.0.1-SNAPSHOT.jar -n ${namespace}
fi

kubectl exec $pod_id -n ${namespace} /var/lib/supervisord/bin/supervisord ctl stop $cmd
kubectl exec $pod_id -n ${namespace} /var/lib/supervisord/bin/supervisord ctl start $cmd

echo "## component ${component} (re)started"