# A Real Example

  * [Introduction](#introduction)
  * [Demo's time](#demos-time)
     * [Build the application](#build-the-application)
     * [Install the components on the cluster](#install-the-components-on-the-cluster)
     * [Check if the Component Client is replying](#check-if-the-component-client-is-replying)
     * [Using K8s](#using-k8s)
     * [Switch from Dev to Build mode](#switch-from-dev-to-build-mode)
     * [Nodejs deployment](#nodejs-deployment)
     * [Scaffold a project](#scaffold-a-project)
  * [Cleanup](#cleanup)


## Introduction

The purpose of this demo is to showcase how you can use the `Component`, `Link` and `Capability` CRs combined with the Kubernetes `operator` to help you to :
- Install your Microservices - Spring Boot applications,
- Instantiate a PostgreSQL database
- Inject the required information to the different Microservices to let a Spring Boot application to access a service which is a http endpoint or to consume a database

The real example consists, as depicted within the following diagram, of two Spring Boot applications and a PostgreSQL Database.

![Composition](component-operator-demo.png)

The application to be deployed can be described using a Fluent DSL syntax as :

`(from:componentA).(to:componentB).(to:serviceA)`

where the `ComponentA` and `ComponentB` correspond respectively to a Spring Boot application `fruit-client-sb` and `fruit-backend-sb`.

The relation `from -> to` indicates that we will `reference` the `ComponentA`  with the `ComponentB` using a `Link`.

The `link`'s purpose is to inject as `Env var(s)` the information required to by example configure the `HTTP client` of the `ComponentA` to access the 
`ComponentB` which exposes a `HTTP endpoint` that the client `ComponentA`  could use to access the `CRUD` operations.

To let the `ComponentB` to consume a database, we will also setup a `link` in order to pass using the database `Secret` the parameters which are needed to configure a Java Datasource's bean.

To avoid that you must manually generate such `CR`, we will use the project [`Dekorate`](https://dekorate.io) which supports to generate Kubernetes resources from Java Annotations or using parameters defined
within an `application.properties` file. 

## Demo's time

Build the `frontend` and the `Backend` using `maven` tool to generate their respective Spring Boot uber jar file
```bash
mvn clean package
``` 

As our different Spring Boot maven projects use the `dekorate` maven dependencies

```xml
 <dependency>
     <groupId>io.dekorate</groupId>
     <artifactId>component-annotations</artifactId>
     <scope>compile</scope>
 </dependency>
 <dependency>
     <groupId>io.dekorate</groupId>
     <artifactId>dekorate-spring-boot</artifactId>
     <scope>compile</scope>
 </dependency>
```

then during the `mvn package` phase, additional yaml resources will be created under the 
directory `dekorate`. They will be next used to deploy the different components on the cluster with the help of the `kubectl` client tool.

```bash
<maven_project>/target/classes/META-INF/dekorate/
```

A few explanation is certainly needed here in order to understand what we did within the different projects to configure them :-)

The `@ComponentApplication` has been added within the `Application.java` java classes and they will be scanned by `dekorate` to generate a `Component.yaml` resource
which contains the definition of the runtime, the name of the component, if a route is needed.

**client**
```java
@ComponentApplication(
    name = "fruit-client-sb",
    exposeService = true
)
```

The `@Link` annotation express using an `@Env` the name of the variable to be injected within the pod in order to let the Spring Boot Application
to configure its HTTP Client to access the HTTP endpoint exposed by the backend service.
 
**client**
```java
@Link(
    name = "link-to-fruit-backend",
    componentName = "fruit-client-sb",
    kind = Kind.Env,
    envs = @Env(
        name = "ENDPOINT_BACKEND",
        value = "http://fruit-backend-sb:8080/api/fruits"
    )
)
```

Like for the Client's component, we will define a `@ComponentApplication` and `@Link` annotations for the `Backend` component. The link will inject from the secret referenced, the parameters that the application
will use to create a `DataSource`'s java bean able to call the `PostgreSQL` instance.

**Backend**
```java
@ComponentApplication(
    name = "fruit-backend-sb",
    exposeService = true,
    envs = @Env(
        name = "SPRING_PROFILES_ACTIVE",
        value = "postgresql-kubedb")
)
```

```java
@Link(
    name = "link-to-database",
    componentName = "fruit-backend-sb",
    kind = Kind.Secret,
    ref = "postgresql-db")
```             
                
To configure the postgresql database to be used by the backend, we will use the `@Capability` annotation to 
configure the database, the user, password and database name.
                
```java
@Capability(
    name = "postgres-db",
    category = "database",
    kind = "postgres",
    version = "10",
    parameters = {
       @Parameter(name = "DB_USER", value = "admin"),
       @Parameter(name = "DB_PASSWORD", value = "admin"),
       @Parameter(name = "DB_NAME", value = "sample-db"),
    }
)
```  

Deploy the generated resource files
```bash
kubectl create ns demo
kubectl apply -f fruit-client-sb/target/classes/META-INF/dekorate/component.yml
kubectl apply -f fruit-backend-sb/target/classes/META-INF/dekorate/component.yml
``` 

Wait a few moment and verify if the status of the components deployed is ready
```bash
TODO - add resources deployed here
```

Then push the uber jar file within the pod using the following bash script 
```bash
./scripts/k8s_push_start.sh fruit-backend sb demo
./scripts/k8s_push_start.sh fruit-client sb demo
```

Check now if the `Component` Client is replying and calls its `HTTP Endpoint` exposed in order to fetch the `fruits` data from the database consumed by the 
other microservice.

```bash
# export FRONTEND_ROUTE_URL=<service_name>.<hostname_or_ip>.<domain_name>
export FRONTEND_ROUTE_URL=fruit-client-sb.10.0.76.186.nip.io
curl -H "Host: fruit-client-sb" ${FRONTEND_ROUTE_URL}/api/client
```

### Switch from Dev to Build mode

TODO: To be reviewed !

Decorate the Component with the following values in order to specify the git info needed to perform a Build, like the name of the component to be selected to switch from
the dev loop to the publish loop

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
  
Patch the component when it has been deployed to switch from `dev` to `build`
  
```bash
kubectl patch cp fruit-backend-sb -p '{"spec":{"deploymentMode":"build"}}' --type=merge
```   

### Nodejs deployment

TODO : To be reviewed !

Build node project locally
```bash
cd fruit-client-nodejs
nvm use v10.1.0
npm audit fix
npm install -s --only=production
```

Run locally
```bash
export ENDPOINT_BACKEND=http://fruit-backend-sb.my-spring-app.195.201.87.126.nip.io/api/fruits
npm run -d start      
```

Deploy the node's component and link it to the Spring Boot fruit backend
```bash
kubectl apply -f fruit-client-nodejs/component.yml
kubectl apply -f fruit-client-nodejs/env-backend-endpoint.yml
```

Push the code and start the nodejs application
```bash
./scripts/push_start.sh fruit-client nodejs
```

Test it locally or remotely
```bash
# locally
http :8080/api/client
http :8080/api/client/1 

#Remotely
route_address=$(kubectl get route/fruit-client-nodejs -o jsonpath='{.spec.host}')
curl http://$route_address/api/client
or 

using httpie client
http http://$route_address/api/client
http http://$route_address/api/client/1
http http://$route_address/api/client/2
http http://$route_address/api/client/3
```

### Scaffold a project

TODO: To be reviewed using `kreate` tool

```bash
git clone git@github.com:snowdrop/scaffold-command.git && cd scaffold-command
go build -o scaffold cmd/scaffold.go
export PATH=$PATH:/Users/dabou/Code/snowdrop/scaffold-command

 scaffold   
? Create from template Yes
? Available templates rest
? Group Id me.snowdrop
? Artifact Id myproject
? Version 1.0.0-SNAPSHOT
? Package name me.snowdrop.myproject
? Spring Boot version 2.1.2.RELEASE
? Use supported version No
? Where should we create your new project ./fruit-demo

cd demo

add to the pom.xml file

<properties>
  ...
  <dekorate.version>0.3.2</dekorate.version>
</properties>
  
and 

<dependencies>
  <dependency>
    <groupId>io.dekorate</groupId>
    <artifactId>component-annotations</artifactId>
    <version>${dekorate.version}</version>
  </dependency>
  <dependency>
    <groupId>io.dekorate</groupId>
    <artifactId>kubernetes-annotations</artifactId>
    <version>${dekorate.version}</version>
  </dependency>
  <dependency>
    <groupId>io.dekorate</groupId>
    <artifactId>dekorate-spring-boot</artifactId>
    <version>${dekorate.version}</version>
  </dependency>
  <!-- spring Boot -->
```

