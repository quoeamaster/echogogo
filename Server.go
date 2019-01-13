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
	"fmt"
	"io/ioutil"
	"os"
	"plugin"
	"strings"
)

// structure for the Server instance's member variables
type Server struct {
	configFile        	string
	configContentJson	ConfigContent

	modules 			map[string]*EchoModule
}

// structure for a valid Echo-module
type EchoModule struct {
	ModulePtr 			*plugin.Plugin
	FxGetRestConfig 	plugin.Symbol
	FxDoAction 			plugin.Symbol
}


// ctor. Create instance of *Server
func NewServer(configFile string) *Server {
	srv := new(Server)
	srv.configFile = configFile
	srv.modules = make(map[string]*EchoModule)

	return srv
}

// ctor. Create instance of *EchoModule
func NewEchoModule(modulePtr *plugin.Plugin, symGetRestConfig plugin.Symbol, symDoAction plugin.Symbol) *EchoModule {
	modPtr := new(EchoModule)
	modPtr.ModulePtr = modulePtr
	modPtr.FxGetRestConfig = symGetRestConfig
	modPtr.FxDoAction = symDoAction

	return modPtr
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
	// load the module(s) available in the folder (load all files with suffix .so)
	err := srv.loadModulesFromRepos()
	if err != nil {
		return err
	}

	return nil
}

func (srv *Server) StopServer() error {
	return nil
}

// load the files / modules within the given repo; modules have a suffix of "so"
func (srv *Server) loadModulesFromRepos() error {
	matchedModulesSlice, err := srv._getModuleFileInfosFromRepos()
	if err != nil {
		return err
	}
	// load the modules through plugin api
	for _, matchedModule := range matchedModulesSlice {
		matchedModulePath := fmt.Sprintf("%v/%v", srv.configContentJson.ModuleRepositoryLocation, matchedModule.Name())
		modulePtr, err := srv._loadModule(matchedModulePath)
		if err != nil {
			/*	TODO: should ignore this unloaded module OR exit? (default is exit if any module can't be LOADED)  */
			return err
		}
		srv.modules[matchedModule.Name()] = modulePtr
	}
	fmt.Printf("map content => %v\n", srv.modules)
	return nil
}

// method to get all files in the repository and then filter valid module files out (suffix of .so)
func (srv *Server) _getModuleFileInfosFromRepos() ([]os.FileInfo, error) {
	matchedModulesPtr := make([]os.FileInfo, 0)
	/*
	 *	read files from the repos dir
	 */
	fileInfos, err := ioutil.ReadDir(srv.configContentJson.ModuleRepositoryLocation)
	if err != nil {
		return nil, err
	}
	/*	only suffix matches ".so" should be treated as a match */
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), ".so") {
			matchedModulesPtr = append(matchedModulesPtr, fileInfo)
		}
	}
	return matchedModulesPtr, nil
}

// method to load modules / plugins; returning a pointer to the actual running ".so" module / plugin / library
func (srv *Server) _loadModule(modulePath string) (*EchoModule, error) {
	modulePtr, err := plugin.Open(modulePath)
	if err != nil {
		return nil, err
	}
	// validation - if the plugin matches the interface method(s)
	symGetRestConfig, err := modulePtr.Lookup("GetRestConfig")
	if err != nil {
		return nil, err
	}
	symDoAction, err := modulePtr.Lookup("DoAction")
	if err != nil {
		return nil, err
	}
	// everything is good, setup the REST module now
	echoModPtr := NewEchoModule(modulePtr, symGetRestConfig, symDoAction)
	err = srv._setupRestForModule(echoModPtr)
	if err != nil {
		return nil, err
	}
	return echoModPtr, nil
}

func (srv *Server) _setupRestForModule(echoModPtr *EchoModule) error {
	// TODO: add back the logic to update the rest api module
	return nil
}
