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
	"bufio"
	"container/list"
	"fmt"
	"github.com/ChristianKniep/qnib.go/qcfg"
	"github.com/docopt/docopt-go"
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

func evalResultDir(path string) map[string]string {
	files, _ := ioutil.ReadDir(path)
	job_res := make(map[string]string)
	const layout = "2006-01-02_15:04:05"
	for _, f := range files {
		re := regexp.MustCompile("osu_(.*).out")
		sub := re.FindStringSubmatch(f.Name())
		if sub != nil {
			subtest := sub[1]
			job_res = updateMap(job_res, parseJobOut(fmt.Sprintf("%s/%s", path, f.Name()), subtest))
		}
		if strings.HasSuffix(f.Name(), ".cfg") {
			job_res = updateMap(job_res, parseJobCfg(fmt.Sprintf("%s/%s", path, f.Name())))
		}
	}
	return job_res

}

func parseJobOut(file_path string, subtest string) map[string]string {
	job_res := make(map[string]string)
	file_descr, err := os.Open(file_path)
	if err != nil {
		panic(err)
	}
	defer file_descr.Close()

	var res list.List
	if subtest == "alltoall" {
		// define pattern to look for
		re := regexp.MustCompile("([0-9]+)[\t ]+([0-9.]+)")
		scanner := bufio.NewScanner(file_descr)
		for scanner.Scan() {
			mat := re.FindStringSubmatch(scanner.Text())
			if mat != nil {
				msg_size, err := strconv.Atoi(mat[1])
				if err != nil {
					// handle error
					fmt.Println(err)
					os.Exit(2)
				}
				usec, err := strconv.ParseFloat(mat[2], 64)
				if err != nil {
					// handle error
					fmt.Println(err)
					os.Exit(2)
				}
				if msg_size <= 64 {
					res.PushBack(usec)
				}
			}
		}

		// Iterate through list and print its contents.
		var total float64 = 0
		for e := res.Front(); e != nil; e = e.Next() {
			total += e.Value.(float64)
		}
		avg := total / float64(res.Len())
		job_res["a2a_avg"] = fmt.Sprintf("%f", avg)
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}
	return job_res
}

func printDir(path string, params map[string]string) {
	job_cfg := evalResultDir(path)
	wtime, _ := strconv.ParseFloat(job_cfg["wall_clock"], 64)
	a2a_avg, _ := strconv.ParseFloat(job_cfg["a2a_avg"], 64)
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
	fmt.Printf(" A2A_AVG:%6.2f |", a2a_avg)
	fmt.Printf("\n")
}

func walkDir(path string, params map[string]string) {
	/* Walks down the path and spawns evaluation if the
	*  directory contains a result file (yaml file)
	 */
	items, _ := ioutil.ReadDir(path)
	re := regexp.MustCompile("osu_.*.out")
	for _, item := range items {
		if item.IsDir() {
			walkDir(fmt.Sprintf("%s/%s", path, item.Name()), params)
		} else if re.FindStringSubmatch(item.Name()) != nil {
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
	var params map[string]string
	params = make(map[string]string)
	arguments, _ := docopt.Parse(usage, nil, true, "0.1", false)
	path := arguments["<path>"].(string)
	// enter directory
	walkDir(path, params)

}
