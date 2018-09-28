# Cloud Native Developer's experience for Spring Boot

The prototype developed within this project aims to resolve the following user's stories for a Spring Boot developer

"As a developer, I want to install a pod running my runtime Java - Spring Boot application where my endpoint is exposed as a Service accessible via a route, where I can instruct the pod to start or stop a command such as "compile", "run java", ..." within the pod"

"As a developer, I would like to generate/scaffold a Spring Boot application supporting different technologies such as JAX-RS - REST, JPA, ...."

"As a developer, I want to customize the application deployed using a MANIFEST yaml file where I can specify, the name of the application, s2i image to be used, maven tool, port of the service, cpu, memory, ...."

"As a developer, I would like to know according to the OpenShift platform, which version of the template/builder and which resources are processed when it will be installed/deployed"

# Table of Contents

   * [Cloud Native Developer's experience for Spring Boot](#cloud-native-developers-experience-for-spring-boot)
   * [Table of Contents](#table-of-contents)
   * [Instructions](#instructions)
      * [Prerequisites](#prerequisites)
      * [Download and validate the Spring boot's go client](#download-and-validate-the-spring-boots-go-client)
      * [Create a project](#create-a-project)
      * [Scaffold a Spring Boot's project](#scaffold-a-spring-boots-project)
      * [Create the development's pod](#create-the-developments-pod)
      * [Push the code](#push-the-code)
      * [Start the java application](#start-the-java-application)
      * [Test the endpoint of the Spring Boot application](#test-the-endpoint-of-the-spring-boot-application)
      * [Remote debug the Java Application](#remote-debug-the-java-application)
      * [Stop/start or restart the spring boot application](#stopstart-or-restart-the-spring-boot-application)
      * [Compile the maven project within the pod (optional)](#compile-the-maven-project-within-the-pod-optional)
      * [Clean up](#clean-up)
      * [More examples](#more-examples)
   * [Technical ideas](#technical-ideas) 

# Instructions

## Prerequisites

- [minishift](https://docs.okd.io/latest/minishift/)
- [oc client](https://www.openshift.org/download.html)

## Download and validate the Spring boot's go client

- Execute within a terminal this curl command in order to download our Spring Boot go client 

  ```bash
  sudo curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/vv0.14.0/sb-darwin-amd64 -o /usr/local/bin/sb
  or 
  sudo curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/vv0.14.0/sb-linux-amd64 -o /usr/local/bin/sb
  sudo chmod +x /usr/local/bin/sb
  ```

- Test it to check that you can access the commands proposed within your terminal 

  ```bash
  $sb -h
  
  sb (ODO's Spring Boot prototype) is a prototype project experimenting supervisord and MANIFEST concepts
  
  Usage:
    sb [command]
  
  Examples:
      # Creating and deploying a spring Boot application
      git clone https://github.com/snowdrop/spring-boot-cloud-devex.git && cd spring-boot-cloud-devex/spring-boot
      sb init -n namespace
      sb push
      sb compile
      sb run
  
  Available Commands:
    build       Build an image of the application
    clean       Remove development pod for the component
    compile     Compile local project within the development pod
    create      Create a Spring Boot maven project
    debug       Debug your SpringBoot application
    exec        Stop, start or restart your SpringBoot application.
    help        Help about any command
    init        Create a development's pod for the component
    push        Push local code to the development pod
    version     Show sb  client version
  
  Flags:
    -a, --application string   Application name (defaults to current directory name)
    -h, --help                 help for sb
    -k, --kubeconfig string    Path to a kubeconfig ($HOME/.kube/config). Only required if out-of-cluster.
        --masterurl string     The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
    -n, --namespace string     Namespace/project (defaults to current project)
  
  Use "sb [command] --help" for more information about a command.
  ```

## Create a project

Create top of your OpenShift's cluster (minishift, ocp, ...) a project where you will deploy and run your spring boot application

  ```bash
  oc new-project <your-namespace>
  ```  
  
  **WARNING**: Check/verify that you have access to an OpenShift cluster before to create your project. 
  
## Scaffold a Spring Boot's project

Within your terminal, move to the folder where you would like to create a Spring REST application

  ```bash
  $ cd ~/my-spring-boot
  ```   

Execute the following command to create a Maven project containing the Spring code exposing a REST Endpoint `/api/greeting`

  ```bash
  sb create -t rest -i my-spring-boot
  ```

  **REMARK** : The parameter `-i` will configure the `artifactId` to use the same name as the folder of your project. The command will populate a project using default values for the GAV, package name, Spring Boot version. That could be of course tailored using
  the different parameters proposed by the command `sb create -h`  

## Create the development's pod

In order to run the Spring Boot application on OpenShift, we will install a Development's pod that our tool will use to interact in order to 
install the application, compile it (optional), debug or start/stop it. Execute then this `init` command.

  ```bash
  sb init
  ```
  
- After a few moment, verify that development's pod is up and running and logs the following messages

  ```bash
  $oc logs $(oc get pod -l app=my-spring-boot -o name)
  time="2018-08-24T09:45:02Z" level=info msg="create process:run-java"
  time="2018-08-24T09:45:02Z" level=info msg="create process:compile-java"
  time="2018-08-24T09:45:02Z" level=info msg="create process:build"
  ``` 

## Push the code
  
As the development's pod has been created and is running the `supervisord` server, we will now compile the maven project and push the binary code.

  ```bash
  mvn clean package
  ```
  
To use the created `uberjar` file located under `/target/<application-name-version>.jar`, then run this command :
  
  ```bash
  sb push --mode binary
  ```
  
## Start the java application

- Launch the Spring Boot Application

  ```bash
  sb exec start
  ime="2018-07-13T11:06:26Z" level=debug msg="succeed to find process:run-java"
  time="2018-07-13T11:06:26Z" level=info msg="try to start program" program=run-java
  time="2018-07-13T11:06:26Z" level=info msg="success to start program" program=run-java
  Starting the Java application using /opt/run-java/run-java.sh ...
  exec java -javaagent:/opt/jolokia/jolokia.jar=config=/opt/jolokia/etc/jolokia.properties -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:+UseParallelOldGC -XX:MinHeapFreeRatio=10 -XX:MaxHeapFreeRatio=20 -XX:GCTimeRatio=4 -XX:AdaptiveSizePolicyWeight=90 -XX:MaxMetaspaceSize=100m -XX:+ExitOnOutOfMemoryError -cp . -jar /deployments/spring-boot-http-1.0.jar
  time="2018-07-13T11:06:27Z" level=debug msg="wait program exit" program=run-java
  I> No access restrictor found, access to any MBean is allowed
  Jolokia: Agent started with URL https://172.17.0.7:8778/jolokia/
    .   ____          _            __ _ _
   /\\ / ___'_ __ _ _(_)_ __  __ _ \ \ \ \
  ( ( )\___ | '_ | '_| | '_ \/ _` | \ \ \ \
   \\/  ___)| |_)| | | | | || (_| |  ) ) ) )
    '  |____| .__|_| |_|_| |_\__, | / / / /
   =========|_|==============|___/=/_/_/_/
   :: Spring Boot ::       (v1.5.14.RELEASE)
  ... 
  2018-07-13 11:06:34.293  INFO 222 --- [           main] o.s.j.e.a.AnnotationMBeanExporter        : Registering beans for JMX exposure on startup
  2018-07-13 11:06:34.304  INFO 222 --- [           main] o.s.c.support.DefaultLifecycleProcessor  : Starting beans in phase 0
  2018-07-13 11:06:34.427  INFO 222 --- [           main] s.b.c.e.t.TomcatEmbeddedServletContainer : Tomcat started on port(s): 8080 (http)
  2018-07-13 11:06:34.436  INFO 222 --- [           main] io.openshift.booster.BoosterApplication  : Started BoosterApplication in 6.412 seconds (JVM running for 7.32) 
  ```
  
## Test the endpoint of the Spring Boot application

Access the endpoint of the Spring Boot application using curl and the route exposed on the cloud platform

  ```bash
  URL="http://$(oc get routes/my-spring-boot -o jsonpath='{.spec.host}')"
  curl $URL/api/greeting
  {"content":"Hello, World!"}% 
  ```   
  
## Start, stop, debug or restart the spring boot application

- The Spring Boot Application can be stopped, started or restarted using respectively these commands:
  ```bash
  sb exec stop
  sb exec start
  sb exec restart
  ```
- You can also debug your application by forwarding the traffic between the pod and your machine using the following command : 
  ```bash
  sb exec debug
  INFO[0000] sb exec start command called                        
  INFO[0000] [Step 1] - Parse MANIFEST of the project if it exists 
  INFO[0000] [Step 2] - Get K8s config file               
  INFO[0000] [Step 3] - Create kube Rest config client using config's file of the developer's machine 
  INFO[0000] [Step 4] - Wait till the dev's pod is available 
  INFO[0000] [Step 5] - Restart Java Application          
  run-java: stopped
  run-java: started
  INFO[0003] [Step 6] - Remote Debug the spring Boot Application ... 
  Forwarding from 127.0.0.1:5005 -> 5005
  ```
  
  **Remark** : You can change the local/remote ports to be used by passing the parameter `-p`. E.g `sb exec debug -p 9009:9009`


## Compile the maven project within the pod (optional)

if you want or prefer to compile the project using the development's pod which contains the maven tool, then execute this command 
responsible to copy the following resources within the pod : `pom.xml, src/ folder`
             
  ```bash
  sb push --mode source
  ```
    
And next execute the compilation using this command
    
  ```bash
  sb compile
  INFO[0000] sb Compile command called                    
  INFO[0000] [Step 1] - Parse MANIFEST of the project if it exists 
  INFO[0000] [Step 2] - Get K8s config file               
  INFO[0000] [Step 3] - Create kube Rest config client using config's file of the developer's machine 
  INFO[0000] [Step 4] - Wait till the dev's pod is available 
  INFO[0000] [Step 5] - Compile ...                       
  compile-java: started
  time="2018-07-13T10:59:05Z" level=info msg="create process:run-java"
  time="2018-07-13T10:59:05Z" level=info msg="create process:compile-java"
  time="2018-07-13T10:59:05Z" level=info msg="create process:build"
  time="2018-07-13T10:59:05Z" level=info msg="create process:echo"
  time="2018-07-13T11:03:39Z" level=debug msg="no auth required"
  time="2018-07-13T11:03:39Z" level=debug msg="succeed to find process:compile-java"
  time="2018-07-13T11:03:39Z" level=info msg="try to start program" program=compile-java
  time="2018-07-13T11:03:39Z" level=info msg="success to start program" program=compile-java
  ==================================================================
  Starting S2I Java Build .....
  Maven build detected
  Initialising default settings /tmp/artifacts/configuration/settings.xml
  Setting MAVEN_OPTS to -XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:+UseParallelOldGC -XX:MinHeapFreeRatio=10 -XX:MaxHeapFreeRatio=20 -XX:GCTimeRatio=4 -XX:AdaptiveSizePolicyWeight=90 -XX:MaxMetaspaceSize=100m -XX:+ExitOnOutOfMemoryError
  Found pom.xml ... 
  Running 'mvn -Dmaven.repo.local=/tmp/artifacts/m2 -s /tmp/artifacts/configuration/settings.xml -e -Popenshift -DskipTests -Dcom.redhat.xpaas.repo.redhatga -Dfabric8.skip=true package --batch-mode -Djava.net.preferIPv4Stack=true '
  Apache Maven 3.5.0 (Red Hat 3.5.0-4.3)
  Maven home: /opt/rh/rh-maven35/root/usr/share/maven
  Java version: 1.8.0_171, vendor: Oracle Corporation
  Java home: /usr/lib/jvm/java-1.8.0-openjdk-1.8.0.171-8.b10.el7_5.x86_64/jre
  Default locale: en_US, platform encoding: ANSI_X3.4-1968
  OS name: "linux", version: "3.10.0-693.21.1.el7.x86_64", arch: "amd64", family: "unix"
  time="2018-07-13T11:03:40Z" level=debug msg="wait program exit" program=compile-java
  [INFO] Error stacktraces are turned on.
  [INFO] Scanning for projects...
  [INFO] Downloading: https://repo1.maven.org/maven2/io/openshift/booster-parent/23/booster-parent-23.pom
  ...
  [INFO] ------------------------------------------------------------------------
  [INFO] BUILD SUCCESS
  [INFO] ------------------------------------------------------------------------
  [INFO] Total time: 01:07 min
  [INFO] Finished at: 2018-07-13T11:04:49Z
  [INFO] Final Memory: 26M/40M
  [INFO] ------------------------------------------------------------------------
  [WARNING] The requested profile "openshift" could not be activated because it does not exist.
  Copying Maven artifacts from /tmp/src/target to /deployments ...
  Running: cp *.jar /deployments
  ... done
  time="2018-07-13T11:04:49Z" level=info msg="program stopped with status:exit status 0" program=compile-java
  time="2018-07-13T11:04:49Z" level=info msg="Don't start the stopped program because its autorestart flag is false" program=compile-java
  ```
  
  **Remark** Before to launch the compilation's command using supervisord, the program will wait till the development's pod is alive !
  
  **Trick**: When the command finishes, you can verify the existence of the uberjar by executing:

  ```bash
  oc rsh $(oc get pod -l app=my-spring-boot -o name) ls -l /deployments
  ```
  
## Clean up
  
  ```bash
  sb clean
  ``` 

## More examples

Additional use cases are developed under the `examples` directory  
  
# Technical ideas

The technically consideration that we have investigated to implement the user stories are described hereafter :

- pod of the application designed with a :
  - initContainer : supervisord [1] where different commands are registered from ENV vars. E.g. start/stop the java runtime, debug or compile (= maven), ... 
  - container : created using Java S2I image
  - shared volume mounted to by example keep maven .m2 repository
- commands can be executed remotely to trigger and action within the developer's pod -> supervisord ctl start|stop program1,....,programN
- OpenShift Template -> converted into individual yaml files (= builder concept) and containing "{{.key}} to be processed by the go template engine
- Developer's user preferences are stored into a MANIFEST yaml (as Cloudfoundry proposes too) which is parsed at bootstrap to create an "Application" struct object used next to process the template and replace the keys with their values

- [1] https://github.com/redhat-developer/odo/issues/556  

