#!/usr/bin/env bash

dir=$(dirname "$0")
component=$1
runtime=$2
project=$component-$runtime
namespace=${3:-my-spring-app}

pod_id=$(kubectl get pod -lapp=${project} -n ${namespace} | grep "Running" | awk '{print $1}')

echo "## Pushing $runtime file ${project} to the pod $pod_id ..."

if [ $runtime = "nodejs" ]; then
  kubectl rsync ${dir}/../$project/ $pod_id:/opt/app-root/src/ --no-perms=true -n ${namespace}
else
  kubectl cp ${dir}/../${project}/pom.xml $pod_id:/usr/src -n ${namespace}
  kubectl cp ${dir}/../${project}/src $pod_id:/usr/src -n ${namespace}
fi

kubectl exec $pod_id -n ${namespace} /var/lib/supervisord/bin/supervisord ctl start build

echo "## component ${component} (re)started"
