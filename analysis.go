// Copyright ©2017 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
func PageRank(g *Graph, damp, tol float64) {
	rank := network.PageRank(directed{g}, damp, tol)

	nodes := g.NodeMap()
	for id, w := range rank {
		nodes[id].Attributes = []dot.Attribute{{"rank", fmt.Sprint(w)}}
	}
}

type directed struct {
	*Graph
}

func (g directed) HasEdgeFromTo(u, v graph.Node) bool { return g.HasEdgeBetween(u, v) }
func (g directed) To(n graph.Node) []graph.Node       { return g.From(n) }

// Closeness performs a closeness centrality analysis on g.
func Closeness(g *Graph) {
	p := path.DijkstraAllPaths(g)
	rank := network.Closeness(g, p)

	nodes := g.NodeMap()
	for id, w := range rank {
		nodes[id].Attributes = []dot.Attribute{{"closeness", fmt.Sprint(w)}}
	}
}

// Farness performs a farness centrality analysis on g.
func Farness(g *Graph) {
	p := path.DijkstraAllPaths(g)
	rank := network.Farness(g, p)

	nodes := g.NodeMap()
	for id, w := range rank {
		nodes[id].Attributes = []dot.Attribute{{"farness", fmt.Sprint(w)}}
	}
}

// Betweenness performs a betweenness centrality analysis on g.
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
		nodes[id].Attributes = []dot.Attribute{{"betweenness", fmt.Sprint(w)}}
	}
}

// EdgeBetweenness performs an edge betweenness centrality analysis on g.
func EdgeBetweenness(g *Graph) {
	rank := network.EdgeBetweenness(g)

	for ids, w := range rank {
		e := g.EdgeBetween(simple.Node(ids[0]), simple.Node(ids[1]))
		e.(*Edge).Attributes = []dot.Attribute{{"edge_betweenness", fmt.Sprint(w)}}
	}
}

// Communities performs a community modularisation of the graph g at the
// specified resolution.
func Communities(g *Graph, resolution float64) {
	r := community.Modularize(g, resolution, nil)

	nodes := g.NodeMap()
	for i, c := range r.Communities() {
		for _, n := range c {
			nodes[n.ID()].Attributes = []dot.Attribute{{"community", fmt.Sprint(i)}}
		}
	}
}

// Clique performs a maximal clique analysis on g where cliques must be
// k-cliques or larger.
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
			for j, a := range attrs {
				if a.Key == "clique" {
					attrs[j] = dot.Attribute{"clique", fmt.Sprintf("%s,%d", a.Value, i)}
					found = true
					break
				}
			}
			if !found {
				nodes[n.ID()].Attributes = []dot.Attribute{{"clique", fmt.Sprint(i)}}
			}
		}
		i++
	}
	for _, n := range nodes {
		for _, a := range n.Attributes {
			if a.Key == "clique" {
				n.Attributes = append(n.Attributes, dot.Attribute{
					"clique_count", fmt.Sprint(len(strings.Split(a.Value, ","))),
				})
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
