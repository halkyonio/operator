## Component's kubernetes operator

[![CircleCI](https://circleci.com/gh/snowdrop/component-operator/tree/master.svg?style=shield)](https://circleci.com/gh/snowdrop/component-operator/tree/master)

Table of Contents
=================

  * [For the users](#for-the-users)
    * [How to play with the Component operator locally](#how-to-play-with-the-component-operator-locally)
    * [A more complex scenario](#a-more-complex-scenario)
    * [Switch fromm inner to outer mode](#switch-fromm-inner-to-outer)
    * [How to install the operator on an cluster](#how-to-install-the-operator-on-the-cluster)
    * [Cleanup the operator](#cleanup)
  * [For the developers only](#for-the-developers-only)
      * [How To create the operator, crd](#how-to-create-the-operator-crd)
  

## For the users

### How to play with the Component operator locally

- Log on to an OpenShift cluster >=3.10 with cluster-admin rights
- Create a namespace `component-operator`
  ```bash
  $ oc new-project component-operator
  ```

- Deploy the resources : service account, rbac and crd definition
  ```bash
  $ oc create -f deploy/sa.yaml
  $ oc create -f deploy/rbac.yaml
  $ oc create -f deploy/crd.yaml
  ```

- Start the Operator locally using the `Main` go file
  ```bash
  $ oc new-project my-spring-app
  $ OPERATOR_NAME=component-operator WATCH_NAMESPACE=my-spring-app KUBERNETES_CONFIG=$HOME/.kube/config go run cmd/manager/main.go
  
- In a separate terminal create a component's yaml file with the following information
```bash
echo "
apiVersion: component.k8s.io/v1alpha1
kind: Component
metadata:
  name: my-spring-boot
spec:
  runtime: spring-boot
  deploymentMode: innerloop" | oc apply -f -
```

- Check if the `operator` has created the following kubernetes resources, part of the `innerloop` deployment mode
  ```bash
  oc get all,pvc,component
  NAME                         READY     STATUS    RESTARTS   AGE
  pod/my-spring-boot-1-hrzcv   1/1       Running   0          11s
  
  NAME                                     DESIRED   CURRENT   READY     AGE
  replicationcontroller/my-spring-boot-1   1         1         1         12s
  
  NAME                         TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)     AGE
  service/component-operator   ClusterIP   172.30.54.114    <none>        60000/TCP   3m
  service/my-spring-boot       ClusterIP   172.30.189.247   <none>        8080/TCP    15s
  
  NAME                                                REVISION   DESIRED   CURRENT   TRIGGERED BY
  deploymentconfig.apps.openshift.io/my-spring-boot   1          1         1         image(copy-supervisord:latest),image(dev-runtime:latest)
  
  NAME                                              DOCKER REPO                                      TAGS      UPDATED
  imagestream.image.openshift.io/copy-supervisord   172.30.1.1:5000/my-spring-app/copy-supervisord   latest    13 seconds ago
  imagestream.image.openshift.io/dev-runtime            172.30.1.1:5000/my-spring-app/dev-runtime            latest    12 seconds ago
  
  NAME                                      HOST/PORT                                           PATH      SERVICES         PORT      TERMINATION   WILDCARD
  route.route.openshift.io/my-spring-boot   my-spring-boot-my-spring-app.192.168.99.50.nip.io             my-spring-boot   <all>                   None
  
  NAME                            STATUS    VOLUME    CAPACITY   ACCESS MODES   STORAGECLASS   AGE
  persistentvolumeclaim/m2-data   Bound     pv0065    100Gi      RWO,ROX,RWX                   15s
  
  NAME                                        AGE
  component.component.k8s.io/my-spring-boot   15s
  ```

- To cleanup the project installed (component)
  ```bash  
  $ oc delete components,route,svc,is,pvc,dc --all=true && 
  ``` 
  
### A more complex scenario   

In order to play with a more complex scenario where we would like to install 2 components: `frontend`, `backend` and database's service from the Ansible Broker's catalog
like also the `links` needed to update the `DeploymentConfig`, then you should execute the following command at the root of the github project within a terminal

  ```bash
  oc apply -f examples/demo/component-client.yml
  oc apply -f examples/demo/component-link-env.yml
  oc apply -f examples/demo/component-crud.yml
  oc apply -f examples/demo/component-service.yml
  oc apply -f examples/demo/component-link.yml
  ```  
  
### Switch from inner to outer

The existing operator supports to switch from the `inner` or development mode (where code must be pushed to the development's pod) to the `outer` mode (responsible to perform a `s2i` build 
deployment using a SCM project). In this case, a container image will be created from the project compiled and next a new deploymentConfig will be created in order to launch the 
runtime.

In order to switch, execute the following operations: 

- Decorate the `Component CRD yaml` file with the following values in order to specify the git info needed to perform a Build, like the name of the component to be selected to switch from
  the dev loop to the outer loop

  ```bash
   annotations:
     app.openshift.io/git-uri: https://github.com/snowdrop/component-operator-demo.git
     app.openshift.io/git-ref: master
     app.openshift.io/git-dir: fruit-backend-sb
     app.openshift.io/artifact-copy-args: "*.jar"
     app.openshift.io/runtime-image: "fruit-backend-sb"
     app.openshift.io/component-name: "fruit-backend-sb"
     app.openshift.io/java-app-jar: "fruit-backend-sb-0.0.1-SNAPSHOT.jar"
  ``` 
  
  **Remark** : When the maven project does not contain multi modules, then replace the name of the folder / module with `.` using the annotation `app.openshift.io/git-dir`
  
- Patch the component when it has been deployed to switch from the `inner` to the `outer` deployment mode
  
  ```bash
  oc patch cp fruit-backend-sb -p '{"spec":{"deploymentMode":"outerloop"}}' --type=merge
  ```   
  
### How to install the operator on a cluster

In the previous section, the operator was launched locally using a `go SDK`. If you would like to install it as a `container` on an OpenShift cluster, then it is required to build 
the container image using the `operator-sdk` tool [1], to push the image and next to install the operator using a `Deployment` kubernetes resource as defined hereafter

  ```bash
  operator-sdk build quay.io/snowdrop/component-operator
  docker push quay.io/snowdrop/component-operator
  oc create -f deploy/operator.yaml
  ``` 
  
[1] https://github.com/operator-framework/operator-sdk   

### Cleanup

  ```bash
  oc delete -f deploy/cr.yaml
  oc delete -f deploy/crd.yaml
  oc delete -f deploy/operator.yaml
  oc delete -f deploy/rbac.yaml
  oc delete -f deploy/sa.yaml
  ```   
  
## For the developers only
  
### How To create the operator, crd

Instructions followed to create the Component's CRD, operator using the `operator-sdk`'s kit

- Execute this command within the `$GOPATH/github.com/$ORG/` folder is a terminal
  ```bash
  operator-sdk new component-operator --api-version=component.k8s.io/v1alpha1 --kind=Component --skip-git-init
  operator-sdk add api --api-version=component.k8s.io/v1alpha1 --kind=Component 
  ```
  using the following parameters 

  Name of the folder to be created : `component-operator`
  Api Group Name   : `component.k8s.io`
  Api Version      : `v1alpha1`
  Kind of Resource : `Component` 

- Build and push the `component-operator` image to `quai.io`s registry
  ```bash
  $ operator-sdk build quay.io/snowdrop/component-operator
  $ docker push quay.io/snowdrop/component-operator
  ```
  
- Update the operator's manifest to use the built image name
  ```bash
  sed -i 's|REPLACE_IMAGE|quay.io/snowdrop/component-operator|g' deploy/operator.yaml
  ```
- Create a namespace `component-operator`

- Deploy the component-operator
  ```bash
  oc new-project component-operator
  oc create -f deploy/sa.yaml
  oc create -f deploy/cluster-rbac.yaml
  oc create -f deploy/crd.yaml
  oc create -f deploy/operator.yaml
  ```

- By default, creating a custom resource triggers the `component-operator` to deploy a busybox pod
  ```bash
  oc create -f deploy/component/cr.yaml
  ```

- Verify that the busybox pod is created
  ```bash
  oc get pod -l app=busy-box
  NAME            READY     STATUS    RESTARTS   AGE
  busy-box   1/1       Running   0          50s
  ```

- Cleanup
  ```bash
  oc delete -f deploy/cr.yaml
  oc delete -f deploy/crd.yaml
  oc delete -f deploy/operator.yaml
  oc delete -f deploy/cluster-rbac.yaml
  oc delete -f deploy/sa.yaml
  ```

- Start operator locally

  ```bash
  $ oc new-project my-spring-app
  $ OPERATOR_NAME=component-operator WATCH_NAMESPACE=my-spring-app KUBERNETES_CONFIG=/Users/dabou/.kube/config go run cmd/manager/main.go
  
  $ oc delete components,route,svc,is,pvc,dc --all=true && go run cmd/sd/sd.go create my-spring-boot
  OR
  $ oc apply -f deploy/component1.yml
  $ oc get all,pvc,components,dc
  
  oc delete components,route,svc,is,pvc,dc --all=true
  ```  
