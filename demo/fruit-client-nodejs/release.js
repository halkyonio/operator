#!/usr/bin/env node

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

// This file is run during the "postbump" lifecyle of standard-version
// We need to be able to update the metadata.label.verion of the resource objects in the openshift template in the .openshiftio folder

const {promisify} = require('util');
const fs = require('fs');
const jsyaml = require('js-yaml');
const packagejson = require('./package.json');

const writeFile = promisify(fs.writeFile);
const readFile = promisify(fs.readFile);

async function updateApplicationYaml () {
  const applicationyaml = jsyaml.safeLoad(await readFile(`${__dirname}/.openshiftio/application.yaml`, {encoding: 'utf8'}));
  // Loop through and update the metadata.label.version to the version from the package.json
  applicationyaml.objects = applicationyaml.objects.map(object => {
    if (object.metadata && object.metadata.labels && object.metadata.labels.version) {
      object.metadata.labels.version = packagejson.version;
    }

    if (object.kind === 'DeploymentConfig') {
      object.spec.template.metadata.labels.version = packagejson.version; // Probably should do a better check here
    }

    return object;
  });

  // Now write the file back out
  await writeFile(`${__dirname}/.openshiftio/application.yaml`, jsyaml.safeDump(applicationyaml, {skipInvalid: true}), {encoding: 'utf8'});
}

updateApplicationYaml();
