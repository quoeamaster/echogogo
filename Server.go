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
	"bytes"
	"errors"
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/quoeamaster/echogogo_plugin"
	"io/ioutil"
	"net/http"
	"os"
	"plugin"
	"reflect"
	"strings"
)

// structure for the Server instance's member variables
type Server struct {
	configFile        	string
	configContentJson	ConfigContent

	modules 			map[string]*EchoModule

	logConfig 			LogConfig
	logger 				Logger
}

// structure for a valid Echo-module
type EchoModule struct {
	ModulePtr 			*plugin.Plugin
	FxGetRestConfig 	plugin.Symbol
	FxDoAction 			plugin.Symbol
	ModulePath			string

	WebservicePath		string
}


// ctor. Create instance of *Server
func NewServer(configFile string) *Server {
	srv := new(Server)
	srv.configFile = configFile
	srv.modules = make(map[string]*EchoModule)

	srv.logConfig = *new(LogConfig)
	srv.logConfig.DefaultLevel = LogLevelInfo
	srv.logConfig.DefaultFuncName = "StartServer"
	srv.logConfig.Filename = "Server"

	srv.logger = NewLogger(LogLevelInfo)

	return srv
}

// ctor. Create instance of *EchoModule
func NewEchoModule(modulePtr *plugin.Plugin, symGetRestConfig plugin.Symbol, symDoAction plugin.Symbol, modulePath string) *EchoModule {
	modPtr := new(EchoModule)
	modPtr.ModulePtr = modulePtr
	modPtr.FxGetRestConfig = symGetRestConfig
	modPtr.FxDoAction = symDoAction
	modPtr.ModulePath = modulePath

	return modPtr
}

// TODO: test on running multiple "modules" e.g. echo + mock

// method to start the echo server
func (srv *Server) StartServer() error {
	srv.logger.LogWithFuncName("bootstrapping SERVER...", "StartServer", srv.logConfig)
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

	srv.logger.LogWithFuncName("SERVER started at port 8001", "", srv.logConfig)
	return http.ListenAndServe(":8001", nil)
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
	srv.logger.LogWithFuncName("searching MODULE(s) to bootstrap...", "loadModulesFromRepos", srv.logConfig)

	// load the modules through plugin api
	for _, matchedModule := range matchedModulesSlice {
		matchedModulePath := fmt.Sprintf("%v/%v", srv.configContentJson.ModuleRepositoryLocation, matchedModule.Name())
		modulePtr, err := srv._loadModule(matchedModulePath)
		if err != nil {
			/*	TODO: should ignore this unloaded module OR exit? (default is exit if any module can't be LOADED)  */
			return err
		}
		srv.modules[matchedModule.Name()] = modulePtr
		// setup the REST api
		err = srv._setupRestForModule(modulePtr)
		if err != nil {
			return err
		}
		srv.logger.LogWithFuncName(fmt.Sprintf("bootstrapped module - %v", matchedModule.Name()), "loadModulesFromRepos", srv.logConfig)
	}
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
	echoModPtr := NewEchoModule(modulePtr, symGetRestConfig, symDoAction, modulePath)

	return echoModPtr, nil
}

func (srv *Server) _setupRestForModule(echoModPtr *EchoModule) error {
	ws := new(restful.WebService)
	configMap := echoModPtr.FxGetRestConfig.(func() map[string]interface{})()
	// fmt.Printf("config returned => %v\n", configMap)

	webservicePath := configMap["path"].(string)
	echoModPtr.WebservicePath = webservicePath
	ws.Path(webservicePath)

	ws = srv._setWebserviceFormat(configMap["consumeFormat"].(string), ws, true)
	ws = srv._setWebserviceFormat(configMap["produceFormat"].(string), ws, false)
	// set endpoints too...
	ws, err := srv._setWebserviceEndPoints(configMap["endPoints"].([]string), ws)
	if err != nil {
		return err
	}
	srv.logger.Log(fmt.Sprintf("MODULE - %v mapped to %v successfully", echoModPtr.ModulePath, echoModPtr.WebservicePath), LogLevelDebug, "Server", "loadModulesFromRepos")
	return nil
}

// method to set the consume and produce format for this WebService
func (srv *Server) _setWebserviceFormat(format string, ws *restful.WebService, isConsume bool) *restful.WebService {
	switch format {
	case echogogo.FORMAT_JSON:
		if isConsume == true {
			ws.Consumes(restful.MIME_JSON)
		} else {
			ws.Produces(restful.MIME_JSON)
		}
	case echogogo.FORMAT_XML:
		if isConsume == true {
			ws.Consumes(restful.MIME_XML)
		} else {
			ws.Produces(restful.MIME_XML)
		}
	case echogogo.FORMAT_XML_JSON:
		if isConsume == true {
			ws.Consumes(restful.MIME_XML, restful.MIME_JSON)
		} else {
			ws.Produces(restful.MIME_XML, restful.MIME_JSON)
		}
	default:
		if isConsume == true {
			ws.Consumes(restful.MIME_XML, restful.MIME_JSON)
		} else {
			ws.Produces(restful.MIME_XML, restful.MIME_JSON)
		}
	}
	return ws
}

func (srv *Server) _setWebserviceEndPoints(endpoints []string, ws *restful.WebService) (ws1 *restful.WebService, err error) {
	ws1 = ws
	err = nil

	defer func() {
		if r:=recover(); r!=nil {
			// catch all unexpected panic(s) and transform it into an Error instead (keep execution of the routine)
			err = fmt.Errorf("exception in setting webservice endpoints: %v\n", r)
		}
	}()

	if endpoints == nil {
		err = errors.New("endpoints missing! ")
		return ws1, err
	}

	for _, endpoint := range endpoints {
		// extract http action / verb and the target path
		parts := strings.Split(endpoint, "::")
		if len(parts) == 2 {
			switch parts[0] {
			case "PUT":
				ws1 = ws1.Route(ws1.PUT(parts[1]).To(srv._webserviceActionRouter))
			case "POST":
				ws1 = ws1.Route(ws1.POST(parts[1]).To(srv._webserviceActionRouter))
			case "DELETE":
				ws1 = ws1.Route(ws1.DELETE(parts[1]).To(srv._webserviceActionRouter))
			case "GET":
				ws1 = ws1.Route(ws1.GET(parts[1]).To(srv._webserviceActionRouter))
			default:
				ws1 = ws1.Route(ws1.GET(parts[1]).To(srv._webserviceActionRouter))
			}
		} else {
			err = fmt.Errorf("invalid endpoint, format for a valid endpoint is [http_verb]::[target_path] (e.g. GET::/hobby ) => %v\n", endpoint)
		}
	}
	restful.Add(ws1)

	return ws1, err
}

// router-like method to intercept every module's DoAction method; prepare the
// request, response and endPoint value for the corresponding DoAction()
func (srv *Server) _webserviceActionRouter(request *restful.Request, response *restful.Response) {
	routePath := request.SelectedRoutePath()
	parts := strings.Split(routePath, "/")
	if len(parts) > 1 {
		// 1st element is "", 2nd element is the module-path (we need this to clarify which module's DoAction method to invoke
		targetModule := "/" + parts[1]
		for _, modulePtr := range srv.modules {
			if modulePtr.WebservicePath == targetModule {
				// invoke the DoAction()
				model := modulePtr.FxDoAction.(func(http.Request, string, ...map[string]interface{}) interface{})(
					*request.Request, targetModule, nil)
				// fmt.Printf("model => %v\n", model)

				switch model.(type) {
				case error:
					// panic or just printf?
					panic(model)
				}
				// based on the response... create the output in either json (default) or xml
				isHandled := false
				for idx := 1; idx < len(parts); idx++ {
					if parts[idx] == "json" {
						if err := response.WriteAsJson(model); err != nil {
							fmt.Printf("%v\n", err)
						}
						isHandled = true
						break
					} else if parts[idx] == "xml" {
						xml := srv._marshalInterface2XmlString(model)
						response.Header().Add("Content-Type", "application/xml")
						if _, err := response.Write([]byte(xml)); err != nil {
							fmt.Printf("%v\n", err)
						}
						isHandled = true
						break
					}
				}
				if !isHandled {
					if err := response.WriteAsJson(model); err != nil {
						fmt.Printf("%v\n", err)
					}
				}
				break	// end - break of (invoke a Matched echo module)
			}
		}	// end -- for (modules)
	} else {
		if err := response.WriteAsJson(fmt.Sprintf("unknown route path => %v\n", routePath)); err != nil {
			// logging
			fmt.Printf("%v\n", err)
		}
	}
}

// method to marshal interface{} into xml string
/* TODO: move to a util package later... */
func (srv *Server) _marshalInterface2XmlString(model interface{}) (xmlString string) {
	xmlString = ""
	var buffer bytes.Buffer
	if model == nil {
		return
	}

	switch model.(type) {
	case map[string]string:
		fModel := model.(map[string]string)
		buffer.WriteString("<response>")
		for key, value := range fModel {
			buffer.WriteString(fmt.Sprintf("<%v>%v</%v>", key, value, key))
		}
		buffer.WriteString("</response>")
	case map[string]interface{}:
		fModel := model.(map[string]interface{})
		buffer.WriteString("<response>")
		for key, value := range fModel {
			buffer.WriteString(fmt.Sprintf("<%v>%v</%v>", key, value, key))
		}
		buffer.WriteString("</response>")
	default:
		// treat as struct like object...
		modelType := reflect.TypeOf(model)
		modelValue := reflect.ValueOf(model)
		// fmt.Printf("$ %v \n", modelType)
		// fmt.Printf("* %v \n", modelValue)
		// fmt.Printf("%%% %v \n", modelType.NumField())

		buffer.WriteString("<response>")
		for idx := 0; idx < modelType.NumField(); idx++ {
			field := modelType.Field(idx)
			if field.Name[0:1] != strings.ToLower(field.Name[0:1]) {
				// public / exported field
				buffer.WriteString(fmt.Sprintf("<%v>%v</%v>", field.Name, modelValue.FieldByName(field.Name), field.Name))
			}	// end -- if (only public / exported field(s) should be shown)
		}
		buffer.WriteString("</response>")
	}
	if buffer.Len() > 0 {
		xmlString = buffer.String()
	}
	return
}