#!/usr/bin/env bash

NS=demo
SERVICEACCOUNT=build
SECRETS=$(oc get secrets -n $NS -o=name)
TOKEN=$(oc sa get-token $SERVICEACCOUNT -n $NS)

function printTitle {
  r=$(typeset i=${#1} c="=" s="" ; while ((i)) ; do ((i=i-1)) ; s="$s$c" ; done ; echo  "$s" ;)
  printf "$r\n$1\n$r\n"
}

for item in $SECRETS;
do
   RES=$(oc get $item --output=json -n $NS | jq -r '.')
   #IS_DOCKERCFG=$(echo $RES | jq '.type == "kubernetes.io/dockercfg"')
   SECRET_NAME=$(echo $RES | jq '.metadata.name' | tr -d '"')

   if [[ $SECRET_NAME = build-dockercfg-* ]]; then
     DOCKER_CFG=$(echo $RES | jq '.data.".dockercfg"' | tr -d '"' | base64 -d)
     echo $DOCKER_CFG | jq '."172.30.1.1:5000".auth as $auth | . += {"http://registry-default.192.168.99.50.nip.io":{"auth": $auth}}' > auth.json
   fi
done

printTitle "## auth.json generated from the secret $SECRET_NAME for the serviceaccount $SERVICEACCOUNT"
cat auth.json

printTitle "## Serviceaccount: $SERVICEACCOUNT"
oc get sa/$SERVICEACCOUNT -o yaml -n $NS

printTitle "TEST1: podman login --authfile auth.json registry-default.192.168.99.50.nip.io -u $SERVICEACCOUNT -p TOKEN --tls-verify=false"
echo "User: $SERVICEACCOUNT"
echo "Password: $TOKEN"
podman login --authfile auth.json http://registry-default.192.168.99.50.nip.io -u $SERVICEACCOUNT -p $TOKEN --tls-verify=false


