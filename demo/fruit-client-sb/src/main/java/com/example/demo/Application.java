/*
 * Copyright 2016-2017 Red Hat, Inc, and individual contributors.
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package com.example.demo;

import io.ap4k.component.annotation.CompositeApplication;
import io.ap4k.component.annotation.Link;
import io.ap4k.component.model.Kind;
import io.ap4k.kubernetes.annotation.Env;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

/**
 * Entry point to the application.
 */
@SpringBootApplication
@CompositeApplication(
    name = "fruit-client-sb",
    exposeService = true
)
@Link(
    name = "link-to-fruit-backend",
    componentName = "fruit-client-sb",
    kind = Kind.Env,
    envs = @Env(
        name = "ENDPOINT_BACKEND",
        value = "http://fruit-backend-sb:8080/api/fruits"
    )
)
public class Application {
    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }
}

