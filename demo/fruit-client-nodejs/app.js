'use strict';

/*
 *
 *  Copyright 2016-2017 Red Hat, Inc, and individual contributors.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */

const path = require('path');
const express = require('express');
const bodyParser = require('body-parser');
const request = require('request');

const app = express();

const endpoint = process.env.ENDPOINT_BACKEND;

app.use(bodyParser.json());
app.use(bodyParser.urlencoded({extended: false}));
app.use('/', express.static(path.join(__dirname, 'public')));

app.use('/api/client/', (req, resp) => {
    var x = request(endpoint)
    req.pipe(x)
    x.pipe(resp)
});

app.use('/api/client/:id', (req, resp) => {
    const id = req.params ? req.params.id : undefined;
    console.log("params", req.params);
    const params = {id: id};
    var x = request({url:endpoint, qs:params})
    req.pipe(x)
    x.pipe(resp)
});

module.exports = app;
