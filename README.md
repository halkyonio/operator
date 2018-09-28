## Spring Boot's operator

Instructions followed to create the Spring Boot's CRD, operator using the `operator-sdk`'s kit

- Execute this command within the `$GOPATH/github.com/$ORG/` folder is a terminal
  ```bash
  operator-sdk new spring-boot-oeprator --api-version=springboot.snowdrop.me/v1alpha1 --kind=SpringBoot
  ```
  using the following parameters 

  Name of the folder to be created : `spring-boot-oeprator`
  Api Group Name   : `springboot.snowdrop.me`
  Api Version      : `v1alpha1`
  Kind of Resource : `SpringBoot` 

- Build and push the `spring-boot-oeprator` image to `quai.io`s registry
  ```bash
  $ operator-sdk build quay.io/snowdrop/spring-boot-oeprator
  $ docker push quay.io/snowdrop/spring-boot-oeprator
  ```
  
- Update the operator's manifest to use the built image name
  ```bash
  sed -i 's|REPLACE_IMAGE|quay.io/snowdrop/spring-boot-oeprator|g' deploy/operator.yaml
  ```
- Create a namespace `spring-boot-oeprator`

- Deploy the spring-boot-oeprator
  ```bash
  oc create -f deploy/sa.yaml
  oc create -f deploy/rbac.yaml
  oc create -f deploy/crd.yaml
  oc create -f deploy/operator.yaml
  ```

- By default, creating a custom resource (Spring Boot App) triggers the `spring-boot-oeprator` to deploy a busybox pod
  ```bash
  oc create -f deploy/cr.yaml
  ```

- Verify that the busybox pod is created
  ```bash
  kubectl get pod -l app=busy-box
  NAME            READY     STATUS    RESTARTS   AGE
  busy-box   1/1       Running   0          50s
  ```

- Cleanup
  ```bash
  oc delete -f deploy/cr.yaml
  oc delete -f deploy/crd.yaml
  oc delete -f deploy/operator.yaml
  oc delete -f deploy/rbac.yaml
  oc delete -f deploy/sa.yaml
  ```