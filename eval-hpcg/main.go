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
	"os"
	"reflect"
)

/*
res_map = {
        "time.total": ['Benchmark Time Summary', 'Total'],
        "problem.dim.x": ['Global Problem Dimensions', 'Global nx'],
        "problem.dim.y": ['Global Problem Dimensions', 'Global ny'],
        "problem.dim.z": ['Global Problem Dimensions', 'Global nz'],
        "gflops": ['DDOT Timing Variations', 'HPCG result is VALID with a GFLOP/s rating of'],
        "local.dim.x": ['Local Domain Dimensions', 'nx'],
        "local.dim.y": ['Local Domain Dimensions', 'ny'],
        "local.dim.z": ['Local Domain Dimensions', 'nz'],
        "mach.num_proc": ['Machine Summary', 'Distributed Processes'],
        "mach.threads_per_proc": ['Machine Summary', 'Threads per processes'],
    }
*/

func nodeToMap(node yaml.Node) yaml.Map {
	m, ok := node.(yaml.Map)
	if !ok {
		panic(fmt.Sprintf("%v is not of type map", node))
	}
	return m
}

func nodeToList(node yaml.Node) yaml.List {
	m, ok := node.(yaml.List)
	if !ok {
		panic(fmt.Sprintf("%v is not of type list", node))
	}
	return m
}

func readLastLine(fname string) {
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		fmt.Println(err)
	}

	buf := make([]byte, 32)
	n, err := file.ReadAt(buf, fi.Size()-int64(len(buf)))
	if err != nil {
		fmt.Println(err)
	}
	buf = buf[:n]
	fmt.Printf("%s", buf)

}

func main() {
	usage := `evaluate HPCG output

Usage:
  eval-hpcg [options] <file>
  eval-hpcg -h | --help
  eval-hpcg --version

Options:
  -h --help     Show this screen.
  --version     Show version.
`

	arguments, _ := docopt.Parse(usage, nil, true, "Naval Fate 2.0", false)
	fmt.Println(arguments)
	file := arguments["<file>"].(string)
	file_descr, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer file_descr.Close()

	fi, err := file_descr.Stat()
	if err != nil {
		fmt.Println(err)
	}
	buf := make([]byte, fi.Size())
	n, err := file_descr.ReadAt(buf, 21)
	if err != nil {
		fmt.Println(err)
	}
	buf = buf[:n]
	fmt.Printf("%s", buf)
	config, err := yaml.ReadFile(file)
	if err != nil {
		log.Fatalf("readfile(%q): %s", file, err)
	}
	param := "Benchmark Time Summary"
	subparam := "Total"
	val := nodeToMap(config.Root)[param]
	subval := nodeToMap(val)[subparam]
	fmt.Println(reflect.TypeOf(val))
	if err != nil {
		log.Fatalf("read_param(%s): %s", param, err)
	}

	fmt.Printf("%s = %s\n", param, val)
	fmt.Printf("%s = %s\n", subparam, subval)
}
