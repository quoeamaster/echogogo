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
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
)

/**
 *	main method to kick start the echogogo server
 */
func main()  {
	echoSrv := cli.NewApp()
	echoSrv.Name = "echogogo server"
	echoSrv.Usage = "main entry point of echogogo server"
	echoSrv.Author = "Jason.Wong"
	echoSrv.Version = "1.0.0"
	echoSrv.Flags = []cli.Flag {
		cli.StringFlag{
			Name: "config, C",
			EnvVar: "envVarEchogogoConfig",
			Usage: "provide a targeted configuration file to startup the server. Can also use the environment-variable: ",
		},
	}

	echoSrv.Action = func(ctx *cli.Context) error {
		srvPtr := NewServer(ctx.String("C"))
		/*
		srvPtr := new(Server)
		srvPtr.configFile = ctx.String("C")
		*/
		err := srvPtr.StartServer()

		return err
	}

	err := echoSrv.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}


