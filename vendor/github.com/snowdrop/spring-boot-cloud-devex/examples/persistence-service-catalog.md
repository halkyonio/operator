# Scaffold a Spring Boot application using JPA's module and connect it to a Database deployed as a service

## Prerequisites

- SB tool is installed (>= [0.14.0](https://github.com/snowdrop/spring-boot-cloud-devex/releases/tag/v0.14.0)). See README.md 
- Minishift (>= v1.23.0+91235ee) with Service Catalog feature enabled

## Install tools

- sb to scaffold the project and create the DeploymentConfig

```bash
sudo curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/v0.14.0/sb-darwin-amd64 -o /usr/local/bin/sb
or 
sudo curl -L https://github.com/snowdrop/spring-boot-cloud-devex/releases/download/v0.14.0/sb-linux-amd64 -o /usr/local/bin/sb
sudo chmod +x /usr/local/bin/sb
```

- odo's tool to create the service from the catalog

```bash
cd $GOPATH/src/github.com/redhat-developer
git clone https://github.com/redhat-developer/odo.git && cd odo
git fetch origin pull/622/head:pr-622
git checkout pr-622
make install && sudo cp $GOPATH/bin/odo /usr/local/bin
```

- Create a Minishift's vm running okd 3.10 and service catalog 

```bash
MINISHIFT_ENABLE_EXPERIMENTAL=y minishift start --extra-clusterup-flags="--enable=service-catalog,automation-service-broker" 
oc login -u admin -p admin
```

## Steps to follow to play the scenario

```bash
oc new-project odo

cd /Temp/my-spring-boot
# rm -rf {src,target} && rm -rf *.{iml,xml,zip}
# Scaffold a JPA Persistence Spring Boot Project
sb create -t crud -i my-spring-boot
mvn clean package

# Create the Development's Pod (= supervisord)
sb init

# Create a service's instance using the OABroker and postgresql DB + secret. Next bind/link the secret to the DC and restart it
odo service create dh-postgresql-apb/dev -p postgresql_user=luke -p postgresql_password=secret -p postgresql_database=my_data -p postgresql_version=9.6
odo service link dh-postgresql-apb my-spring-boot

# Push the uber jar file and start the supervisord's run command
sb push --mode binary
sb exec start
018-09-12 13:56:54.582  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/heapdump || /heapdump.json],methods=[GET],produces=[application/octet-stream]}" onto public void org.springframework.boot.actuate.endpoint.mvc.HeapdumpMvcEndpoint.invoke(boolean,javax.servlet.http.HttpServletRequest,javax.servlet.http.HttpServletResponse) throws java.io.IOException,javax.servlet.ServletException
2018-09-12 13:56:54.583  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/configprops || /configprops.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.585  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/auditevents || /auditevents.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public org.springframework.http.ResponseEntity<?> org.springframework.boot.actuate.endpoint.mvc.AuditEventsMvcEndpoint.findByPrincipalAndAfterAndType(java.lang.String,java.util.Date,java.lang.String)
2018-09-12 13:56:54.586  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/autoconfig || /autoconfig.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.590  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/loggers/{name:.*}],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.LoggersMvcEndpoint.get(java.lang.String)
2018-09-12 13:56:54.590  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/loggers/{name:.*}],methods=[POST],consumes=[application/vnd.spring-boot.actuator.v1+json || application/json],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.LoggersMvcEndpoint.set(java.lang.String,java.util.Map<java.lang.String, java.lang.String>)
2018-09-12 13:56:54.591  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/loggers || /loggers.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.591  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/mappings || /mappings.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.592  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/beans || /beans.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.593  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/info || /info.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.594  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/trace || /trace.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.595  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/metrics/{name:.*}],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.MetricsMvcEndpoint.value(java.lang.String)
2018-09-12 13:56:54.595  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/metrics || /metrics.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.596  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/dump || /dump.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.597  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/env/{name:.*}],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EnvironmentMvcEndpoint.value(java.lang.String)
2018-09-12 13:56:54.597  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/env || /env.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.EndpointMvcAdapter.invoke()
2018-09-12 13:56:54.598  INFO 32 --- [           main] o.s.b.a.e.mvc.EndpointHandlerMapping     : Mapped "{[/health || /health.json],methods=[GET],produces=[application/vnd.spring-boot.actuator.v1+json || application/json]}" onto public java.lang.Object org.springframework.boot.actuate.endpoint.mvc.HealthMvcEndpoint.invoke(javax.servlet.http.HttpServletRequest,java.security.Principal)
2018-09-12 13:56:55.233  INFO 32 --- [           main] o.s.j.e.a.AnnotationMBeanExporter        : Registering beans for JMX exposure on startup
2018-09-12 13:56:55.256  INFO 32 --- [           main] o.s.c.support.DefaultLifecycleProcessor  : Starting beans in phase 0
2018-09-12 13:56:55.437  INFO 32 --- [           main] s.b.c.e.t.TomcatEmbeddedServletContainer : Tomcat started on port(s): 8080 (http)
2018-09-12 13:56:55.448  INFO 32 --- [           main] com.example.demo.CrudApplication         : Started CrudApplication in 27.545 seconds (JVM running for 29.165)
```

- Query the service

```bash
http my-spring-boot-odo.192.168.99.50.nip.io/api/fruits
HTTP/1.1 200 
Cache-control: private
Content-Type: application/json;charset=UTF-8
Date: Wed, 12 Sep 2018 13:57:47 GMT
Set-Cookie: d63dbdcbd8cb849d5e746267aaa4ed8d=ab5897d9975f4716398ac33b10f3c13d; path=/; HttpOnly
Transfer-Encoding: chunked
X-Application-Context: application:openshift-catalog

[
    {
        "id": 1,
        "name": "Cherry"
    },
    {
        "id": 2,
        "name": "Apple"
    },
    {
        "id": 3,
        "name": "Banana"
    }
]
```


