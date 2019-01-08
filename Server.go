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

type Server struct {
	configFile        string
	configContentJson ConfigContent
}

// method to start the echo server
func (srv *Server) StartServer() error {
	// load the config file contents if valid
	if srv.configFile == "" {
		cModelPtr := new(ConfigContent)
		cModelPtr.ModuleRepositoryLocation = "modules"
		srv.configContentJson = *cModelPtr

	} else {
		val, err := LoadConfigContent(srv.configFile)
		if err != nil {
			return err
		}
		srv.configContentJson = ConfigContent(*val)
		// fmt.Printf("%v\n", srv.configContentJson.ModuleRepositoryLocation)
	}
	// TODO: load the module(s) available in the folder (load all files with suffix .so)

	return nil
}

func (srv *Server) StopServer() error {
	return nil
}


