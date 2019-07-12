/*
 * Copyright 2016-2017 Red Hat, Inc, and individual contributors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package com.example.demo;

import io.dekorate.component.annotation.Capability;
import io.dekorate.component.annotation.ComponentApplication;
import io.dekorate.component.annotation.Link;
import io.dekorate.component.annotation.Parameter;
import io.dekorate.component.model.Kind;
import io.dekorate.kubernetes.annotation.Env;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@ComponentApplication(
    name = "fruit-backend-sb",
    exposeService = true,
    envs = @Env(
        name = "SPRING_PROFILES_ACTIVE",
        value = "postgresql-kubedb")
)
@Link(
    name = "link-to-database",
    componentName = "fruit-backend-sb",
    kind = Kind.Secret,
    ref = "postgres-db-config")
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
@SpringBootApplication
public class CrudApplication {
    public static void main(String[] args) {
        SpringApplication.run(CrudApplication.class, args);
    }
}
