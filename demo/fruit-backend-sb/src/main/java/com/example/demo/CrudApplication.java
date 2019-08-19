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

import io.dekorate.halkyon.annotation.HalkyonCapability;
import io.dekorate.halkyon.annotation.HalkyonComponent;
import io.dekorate.halkyon.annotation.HalkyonLink;
import io.dekorate.halkyon.annotation.Parameter;
import io.dekorate.halkyon.model.Kind;
import io.dekorate.kubernetes.annotation.Env;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@HalkyonComponent(
    name = "fruit-backend-sb",
    exposeService = true,
    port = 8080
)
@HalkyonLink(
    name = "link-to-database",
    componentName = "fruit-backend-sb",
    kind = Kind.Secret,
    ref = "postgres-db-config")
@HalkyonCapability(
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
