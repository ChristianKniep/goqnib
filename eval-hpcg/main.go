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
	"./cfg"
	"bytes"
	"fmt"
	"github.com/docopt/docopt-go"
	"github.com/kylelemons/go-gypsy/yaml"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func getFileInfo(file_path string) os.FileInfo {
	file_descr, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file_descr.Close()

	fi, err := file_descr.Stat()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return fi
}

func parseJobCfg(file_path string) map[string]string {
	mymap := make(map[string]string)
	err := cfg.Load(file_path, mymap)
	if err != nil {
		log.Fatal(err)
	}
	fi := getFileInfo(file_path)
	const layout = "2006-01-02_15:04:05"
	mymap["mod_time"] = fi.ModTime().Format(layout)
	return mymap
}

func updateMap(old map[string]string, new map[string]string) map[string]string {
	for key, value := range new {
		old[key] = value
	}
	return old
}

func evalResultDir(path string) (yaml.Node, map[string]string) {
	files, _ := ioutil.ReadDir(path)
	job_cfg := make(map[string]string)
	job_res := *new(yaml.Node)
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".yaml") {
			//2014.10.08.22.12.51
			re := regexp.MustCompile("20[0-9]+.[0-9]+.[0-9]+.[0-9]+.[0-9]+.[0-9]+")
			ydate := re.FindStringSubmatch(f.Name())
			// TODO: []string -> string?
			if ydate != nil {
				job_cfg["yaml_date"] = strings.Join(ydate, "")
			}
			job_res = parseHPCG(fmt.Sprintf("%s/%s", path, f.Name()))
		}
		if strings.HasSuffix(f.Name(), ".cfg") {
			job_cfg = updateMap(job_cfg, parseJobCfg(fmt.Sprintf("%s/%s", path, f.Name())))
		}
	}
	return job_res, job_cfg

}

func parseHPCG(file_path string) yaml.Node {
	file_descr, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file_descr.Close()

	fi, err := file_descr.Stat()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	buf := make([]byte, fi.Size())
	n, err := file_descr.ReadAt(buf, 19)
	buf = buf[:n]
	reader := bytes.NewReader(buf)
	config, err := yaml.Parse(reader)
	return config
}

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

func fetch2D(node yaml.Node, param string) yaml.Node {
	s := strings.Split(param, "##")
	val := nodeToMap(node)[strings.Trim(s[0], " ")]
	if len(s) > 1 {
		val = nodeToMap(val)[strings.Trim(s[1], " ")]
	}
	return val
}
func printDir(path string, params map[string]string) {
	job_res, job_cfg := evalResultDir(path)
	for key, value := range params {
		job_cfg[key] = fmt.Sprint(fetch2D(job_res, value))
	}
	gflops, _ := strconv.ParseFloat(job_cfg["GFLOPs"], 64)
	ctime, _ := strconv.ParseFloat(job_cfg["CTIME"], 64)
	wtime, _ := strconv.ParseFloat(job_cfg["wall_clock"], 64)
	_, present := job_cfg["yaml_date"]
	if present {
		fmt.Printf("| YTIME:%s |", job_cfg["yaml_date"])
	} else {
		fmt.Printf("| MTIME:%s |", job_cfg["mod_time"])
	}
	fmt.Printf(" #PROC:%2s |", job_cfg["#PROC"])
	fmt.Printf(" NODES:%-27s |", job_cfg["slurm_nodelist"])
	fmt.Printf(" MPI:%-10s |", job_cfg["mpi_ver"])
	fmt.Printf(" WTIME:%6.1f |", wtime)
	fmt.Printf(" CTIME:%6.1f |", ctime)
	fmt.Printf(" GFLOPS:%8.5f |", gflops)
	fmt.Printf("\n")
}

func walkDir(path string, params map[string]string) {
	/* Walks down the path and spawns evaluation if the
	*  directory contains a result file (yaml file)
	 */
	items, _ := ioutil.ReadDir(path)
	for _, item := range items {
		if item.IsDir() {
			walkDir(fmt.Sprintf("%s/%s", path, item.Name()), params)
		} else if strings.HasSuffix(item.Name(), ".yaml") {
			printDir(path, params)
		}
	}
}

func main() {
	usage := `evaluate HPCG output

Usage:
  eval-hpcg [options] <path>
  eval-hpcg -h | --help
  eval-hpcg --version

Options:
  -h --help     Show this screen.
  --version     Show version.
`
	var pwidth map[string]int
	pwidth = make(map[string]int)
	var params map[string]string
	params = make(map[string]string)
	params["CTIME"] = "Benchmark Time Summary ## Total"
	pwidth["CTIME"] = 7
	params["GFLOPs"] = "GFLOP/s Summary ## Raw Total"
	pwidth["GFLOPs"] = 7
	params["#PROC"] = "Machine Summary##Distributed Processes"
	pwidth["#PROC"] = 2
	/*
		params["PNx"] = "Processor Dimensions##npx"
		pwidth["PNx"] = 3
		params["PNy"] = "Processor Dimensions##npy"
		pwidth["PNy"] = 3
		params["PNz"] = "Processor Dimensions##npz"
		pwidth["PNz"] = 3
	*/
	arguments, _ := docopt.Parse(usage, nil, true, "0.1", false)
	path := arguments["<path>"].(string)
	// enter directory
	walkDir(path, params)

}
