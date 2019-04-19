// Copyright ©2017 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/simple"
)

type node struct {
	id   int
	name string
	desc string
}

func (n node) ID() int64     { return int64(n.id) }
func (n node) DOTID() string { return `"` + n.name + `"` }
func (n node) DOTAttributes() []encoding.Attribute {
	return []encoding.Attribute{{Key: "desc", Value: `"` + n.desc + `"`}}
}

const (
	network = iota + 1
	vertices
	edges
)

func parsenet(r io.Reader, g graph.UndirectedBuilder) (graph.Undirected, string, error) {
	var name string
	var n int
	nodes := make(map[int]graph.Node)
	sc := bufio.NewScanner(r)
	var state int
	for sc.Scan() {
		if len(sc.Bytes()) == 0 || sc.Bytes()[0] == '%' {
			continue
		}
		text := sc.Text()
		if text[0] == '*' {
			switch t := strings.ToLower(text); {
			case strings.HasPrefix(t, "*network "):
				state = network
				name = strings.SplitN(text, " ", 2)[1]
			case strings.HasPrefix(t, "*vertices "):
				state = vertices
				var err error
				n, err = strconv.Atoi(strings.Fields(t)[1])
				if err != nil {
					return nil, "", err
				}
			case strings.HasPrefix(t, "*edges"):
				state = edges
			}
			continue
		}

		switch state {
		case network:
			// Do nothing.
		case vertices:
			var name, desc string
			text = strings.TrimSpace(text)
			f := strings.SplitN(text, " ", 2)
			id, err := strconv.Atoi(f[0])
			if err != nil {
				return nil, "", err
			}
			attr, err := strconv.Unquote(f[1])
			if err != nil {
				return nil, "", err
			}
			if f[1] == `"Unknow protein !!!"` {
				name = fmt.Sprintf("Unknown%04d", id)
				desc = attr
			} else {
				attrs := strings.SplitN(attr, " ", 2)
				if len(attrs) == 2 {
					desc = attrs[1]
				}
				name = attrs[0]
			}
			n := node{id: id, name: name, desc: desc}
			nodes[id] = n
			g.AddNode(n)
		case edges:
			text = strings.TrimSpace(text)
			f := strings.Fields(text)
			if len(f) < 2 {
				return nil, "", fmt.Errorf("too few parameters for edge: %q", text)
			}
			from, err := strconv.Atoi(f[0])
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse from node id for %q: %v", text, err)
			}
			to, err := strconv.Atoi(f[1])
			if err != nil {
				return nil, "", fmt.Errorf("failed to parse to node id for %q: %v", text, err)
			}
			if from == to {
				continue
			}
			g.SetEdge(simple.Edge{F: nodes[from], T: nodes[to]})
		default:
			panic("cannot reach")
		}
	}

	if len(nodes) != n {
		return nil, "", fmt.Errorf("unexpected number of nodes: got=%d want=%d", len(nodes), n)
	}
	return g, name, nil
}

func main() {
	g, name, err := parsenet(os.Stdin, simple.NewUndirectedGraph())
	if err != nil {
		log.Fatalf("failed parse net file: %v", err)
	}
	b, err := dot.Marshal(g, name, "", "  ")
	if err != nil {
		log.Fatalf("failed marshal graph: %v", err)
	}
	fmt.Printf("%s\n", b)
}
