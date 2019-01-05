/*
 * Licensed to Echogogo under one or more contributor
 * license agreements. See the NOTICE file distributed with
 * this work for additional information regarding copyright
 * ownership. Echogogo licenses this file to you under
 * the Apache License, Version 2.0 (the "License"); you may
 * not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ConfigContent struct {
	ModuleRepositoryLocation string `json:"moduleRepositoryLocation" description:"location to find the module(s)"`
}


// method to load config content based on the given configuration file location
func LoadConfigContent(cfgFile string) (*ConfigContent, error) {
	content, err := os.Open(cfgFile)
	defer func() {
		content.Close()
	}()
	if err != nil {
		return nil, err
	}
	bArrContent, err := ioutil.ReadAll(content)
	if err != nil {
		return nil, err
	}

	var configContent ConfigContent
	err = json.Unmarshal(bArrContent, &configContent)
	if err != nil {
		return nil, err
	} else {
		return &configContent, nil
	}
}

