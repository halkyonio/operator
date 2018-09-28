# Scaffold a Spring Boot project using a CRUD template

## Prerequisites

 - SB tool is installed (>= [0.12.0](https://github.com/snowdrop/spring-boot-cloud-devex/releases/tag/vv0.14.0)). See README.md 

## Step by step instructions

- Create a `my-spring-boot` project directory and move under this directory

```bash
cd /path/to/my-spring-boot
```

- Log to your OpenShift's cluster and create a crud namespace

```bash
oc login -u admin -p admin
oc new-project crud
```

- Install the `Postgresql` database on OpenShift

```bash
oc new-app \
  -e POSTGRESQL_USER=luke \
  -e POSTGRESQL_PASSWORD=secret \
  -e POSTGRESQL_DATABASE=my_data \
  openshift/postgresql-92-centos7 \
  --name=my-database
```

- Create a `MANIFEST`'s project file within the current folder containing the ENV vars used by the spring's boot application to be authenticated with the Database and to use the Openshift's profile 
 which supports `Postgresql`'s database

```yaml
name: my-spring-boot
env:
  - name: DB_USERNAME
    value: luke
  - name: DB_PASSWORD
    value: secret
  - name: SPRING_PROFILES_ACTIVE
    value: openshift
```

- Initialize the Development's pod 

```bash
sb init -n crud
```

- Scaffold the CRUD project using as artifactId - `my-spring-boot` name

```bash
sb create -t crud -i my-spring-boot
```

- Generate the binary uber jar file and push it

```bash
mvn clean package
sb push --mode binary
```

- Start the Spring Boot application

```bash
sb exec start
```

- Use `curl` or `httpie` tool to fetch the records using the Spring Boot CRUD endpoint exposed

```bash
http http://MY_APP_NAME-MY_PROJECT_NAME.OPENSHIFT_HOSTNAME/api/fruits
HTTP/1.1 200 
Cache-control: private
Content-Type: application/json;charset=UTF-8
Date: Mon, 27 Aug 2018 07:54:21 GMT
Set-Cookie: 23678da2d4b6649bf39522f45e3064f1=45e36b4277a6cd39f15bb8efaa87c882; path=/; HttpOnly
Transfer-Encoding: chunked
X-Application-Context: application:openshift

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
        "id": 3,git pull
        "name": "Banana"
    }
]
```

- Add a `fruit`

```bash
curl -H "Content-Type: application/json" -X POST -d '{"name":"pear"}' http://MY_APP_NAME-MY_PROJECT_NAME.OPENSHIFT_HOSTNAME/api/fruits
```