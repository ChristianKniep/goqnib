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

/***************
 * Slidly modified version of John Graham-Cumming script
 * presented at dotGo 2014 (Paris)
 * source: https://github.com/jgrahamc/dotgo
 */

package partasker

import (
	"bufio"
	"fmt"
	"log"
	//"net"
	"os"
	//"strings"
	"sync"
	//"github.com/docopt/docopt-go"
)

type factory interface {
	make(line string) tasks
}

type tasks interface {
	process()
	print()
}

type mytask struct {
	name       string
	err        error
	cloudflare bool
}

func (l *mytask) print() {
	fmt.Println(l.name)
}
func (l *mytask) process() {
	fmt.Println("Inside process: %s", l.name)
	/*
		nss, err := net.LookupNS(l.name)
		if err != nil {
			l.err = err
		} else {
			for _, ns := range nss {
				if strings.HasSuffix(ns.Host, ".ns.cludflare.com.") {
					l.cloudflare = true
					break
				}
			}
		}
	*/
}

type taskFactory struct {
}

func (f *taskFactory) make(line string) tasks {
	return &mytask{name: line}
}

func run(f factory) {
	var wg sync.WaitGroup

	in := make(chan tasks)

	wg.Add(1)
	go func() {
		s := bufio.NewScanner(os.Stdin)
		for s.Scan() {
			in <- f.make(s.Text())
		}
		if s.Err() != nil {
			log.Fatalf("Error reading STDIN: %s", s.Err())
		}
		close(in)
		wg.Done()
	}()

	out := make(chan tasks)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for t := range in {
				t.process()
				out <- t
			}
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	for t := range out {
		t.print()
	}
}
