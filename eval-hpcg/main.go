// Copyright 2013 Google, Inc.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/kylelemons/go-gypsy/yaml"
	"log"
	//"reflect"
)

func main() {
	usage := `evaluate HPCG output

Usage:
  eval-hpcg [options] <params>
  eval-hpcg -h | --help
  eval-hpcg --version

Options:
  -y <path>	   	"(Simple) YAML file to read"
  -h --help     Show this screen.
  --version     Show version.
`

	arguments, _ := docopt.Parse(usage, nil, true, "Naval Fate 2.0", false)
	fmt.Println(arguments)
	//file := arguments["-y"]
	file := "config.yaml"
	//param := arguments["<param>"]
	param := "key"
	config, err := yaml.ReadFile(file)
	if err != nil {
		log.Fatalf("readfile(%q): %s", file, err)
	}
	val, err := config.Get(param)
	if err != nil {
		log.Fatalf("read_param(%s): %s", param, err)
	}
	fmt.Printf("%s = %s\n", param, val)
}
