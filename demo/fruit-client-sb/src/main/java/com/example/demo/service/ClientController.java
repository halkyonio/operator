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
package com.example.demo.service;

import com.example.demo.Fruit;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestTemplate;

import java.util.List;

@RestController
@RequestMapping("/api")
public class ClientController {

    @Value("${endpoint.fruit:}")
    private String endPoint;
    private String suffix = "/{id}";

    @GetMapping("/client")
    public List<Fruit> getFruits() {
        RestTemplate restTemplate = new RestTemplate();
        return restTemplate.getForObject(endPoint, List.class);
    }

    @GetMapping("/client/{id}")
    public com.example.demo.Fruit getFruitById(@PathVariable String id) {
        RestTemplate restTemplate = new RestTemplate();
        return restTemplate.getForObject(endPoint + suffix, Fruit.class, id);
    }
}
