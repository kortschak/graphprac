// Copyright ©2017 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package graphprac provides helper routines for short practical in
// graph analysis.
package graphprac

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/community"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/network"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

// PageRank performs a PageRank analysis on g using the provided damping
// and tolerance parameters.
//
// The PageRank value is written into the "rank" attribute of each node.
func PageRank(g *Graph, damp, tol float64) {
	rank := network.PageRank(directed{g}, damp, tol)

	nodes := g.NodeMap()
	for id, w := range rank {
		nodes[id].Attributes.Set("rank", fmt.Sprint(w))
	}
}

type directed struct {
	*Graph
}

func (g directed) HasEdgeFromTo(u, v graph.Node) bool { return g.HasEdgeBetween(u, v) }
func (g directed) To(n graph.Node) []graph.Node       { return g.From(n) }

// Closeness performs a closeness centrality analysis on g.
//
// The closeness centrality value is written into the "closeness" attribute of each node.
func Closeness(g *Graph) {
	p := path.DijkstraAllPaths(g)
	rank := network.Closeness(g, p)

	nodes := g.NodeMap()
	for id, w := range rank {
		nodes[id].Attributes.Set("closeness", fmt.Sprint(w))
	}
}

// Farness performs a farness centrality analysis on g.
//
// The farness centrality value is written into the "farness" attribute of each node.
func Farness(g *Graph) {
	p := path.DijkstraAllPaths(g)
	rank := network.Farness(g, p)

	nodes := g.NodeMap()
	for id, w := range rank {
		nodes[id].Attributes.Set("farness", fmt.Sprint(w))
	}
}

// Betweenness performs a betweenness centrality analysis on g.
//
// The betweenness centrality value is written into the "betweenness" attribute of each node.
func Betweenness(g *Graph) {
	rank := network.Betweenness(g)
	nodes := g.NodeMap()
	// network.Betweenness does not retain zero
	// betweenness values, so fill them in.
	for _, n := range nodes {
		if _, ok := rank[n.ID()]; !ok {
			rank[n.ID()] = 0
		}
	}

	for id, w := range rank {
		nodes[id].Attributes.Set("betweenness", fmt.Sprint(w))
	}
}

// EdgeBetweenness performs an edge betweenness centrality analysis on g.
//
// The edge betweenness centrality value is written into the "edge_betweenness" attribute of each edge.
func EdgeBetweenness(g *Graph) {
	rank := network.EdgeBetweenness(g)

	for ids, w := range rank {
		e := g.EdgeBetween(simple.Node(ids[0]), simple.Node(ids[1]))
		e.(*Edge).Attributes.Set("edge_betweenness", fmt.Sprint(w))
	}
}

// Communities performs a community modularisation of the graph g at the
// specified resolution.
//
// The community identity value is written into the "community" attribute of each node.
func Communities(g *Graph, resolution float64) {
	r := community.Modularize(g, resolution, nil)

	nodes := g.NodeMap()
	for i, c := range r.Communities() {
		for _, n := range c {
			nodes[n.ID()].Attributes.Set("community", fmt.Sprint(i))
		}
	}
}

// Clique performs a maximal clique analysis on g where cliques must be
// k-cliques or larger.
//
// The clique membership values are written as a comma-separated list into the
// "clique" attribute of each node and the number of cliques a node is a member of
// is written into "clique_count".
func Clique(g *Graph, k int) {
	mc := topo.BronKerbosch(g)
	var ck int
	for _, c := range mc {
		if len(c) >= k {
			ck++
		}
	}
	nodes := g.NodeMap()
	var i int
	for _, c := range mc {
		if len(c) < k {
			continue
		}
		for _, n := range c {
			found := false
			attrs := nodes[n.ID()].Attributes
			if attrs.Get("clique_count") == "" {
				for j, a := range attrs {
					if a.Key == "clique" {
						attrs[j] = dot.Attribute{"clique", fmt.Sprintf("%s,%d", a.Value, i)}
						found = true
						break
					}
				}
			}
			if !found {
				nodes[n.ID()].Attributes.Set("clique", fmt.Sprint(i))
				nodes[n.ID()].Attributes.Set("clique_count", "")
			}
		}
		i++
	}
	for _, n := range nodes {
		for _, a := range n.Attributes {
			if a.Key == "clique" {
				n.Attributes.Set("clique_count", fmt.Sprint(len(strings.Split(a.Value, ","))))
				break
			}
		}
	}
}

// NodesByAttribute return a slice of nodes sorted descending by the
// given attribute.
func NodesByAttribute(attr string, g *Graph) ([]*Node, error) {
	var nodes []*Node
	var vals []float64
	for _, n := range g.Nodes() {
		n := n.(*Node)
		var v float64
		var err error
		for _, a := range n.Attributes {
			if a.Key == attr {
				v, err = strconv.ParseFloat(a.Value, 64)
				if err != nil {
					return nil, err
				}
				break
			}
		}
		nodes = append(nodes, n)
		vals = append(vals, v)
	}
	sort.Sort(nodesByAttr{vals: vals, nodes: nodes})
	return nodes, nil
}

type nodesByAttr struct {
	nodes []*Node
	vals  []float64
}

func (a nodesByAttr) Len() int           { return len(a.nodes) }
func (a nodesByAttr) Less(i, j int) bool { return a.vals[i] > a.vals[j] }
func (a nodesByAttr) Swap(i, j int) {
	a.nodes[i], a.nodes[j] = a.nodes[j], a.nodes[i]
	a.vals[i], a.vals[j] = a.vals[j], a.vals[i]
}

// EdgesByAttribute return a slice of edges sorted descending by the
// given attribute.
func EdgesByAttribute(attr string, g *Graph) ([]*Edge, error) {
	var edges []*Edge
	var vals []float64
	for _, n := range g.Edges() {
		n := n.(*Edge)
		var v float64
		var err error
		for _, a := range n.Attributes {
			if a.Key == attr {
				v, err = strconv.ParseFloat(a.Value, 64)
				if err != nil {
					return nil, err
				}
				break
			}
		}
		edges = append(edges, n)
		vals = append(vals, v)
	}
	sort.Sort(edgesByAttr{vals: vals, edges: edges})
	return edges, nil
}

type edgesByAttr struct {
	edges []*Edge
	vals  []float64
}

func (a edgesByAttr) Len() int           { return len(a.edges) }
func (a edgesByAttr) Less(i, j int) bool { return a.vals[i] > a.vals[j] }
func (a edgesByAttr) Swap(i, j int) {
	a.edges[i], a.edges[j] = a.edges[j], a.edges[i]
	a.vals[i], a.vals[j] = a.vals[j], a.vals[i]
}
