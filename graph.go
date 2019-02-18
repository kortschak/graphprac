// Copyright ©2017 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graphprac

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
	"gonum.org/v1/gonum/graph/iterator"
	"gonum.org/v1/gonum/graph/simple"
)

// Graph is a general undirected graph with node and edge attributes.
type Graph struct {
	*simple.UndirectedGraph
	GraphAttrs, NodeAttrs, EdgeAttrs Attributes
}

// ReadGraph reads a DOT file and returns the encoded graph.
func NewGraph(file string) (*Graph, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	g := &Graph{UndirectedGraph: simple.NewUndirectedGraph()}

	err = dot.Unmarshal(b, g)
	if err != nil {
		return nil, err
	}

	return g, nil
}

// NewNode adds a new node with a unique node ID to the graph.
func (g *Graph) NewNode() graph.Node {
	return &Node{NodeID: g.UndirectedGraph.NewNode().ID()}
}

// NewEdge adds a new edge from the source to the destination node to the graph,
// or returns the existing edge if already present.
func (g *Graph) NewEdge(from, to graph.Node) graph.Edge {
	if e := g.Edge(from.ID(), to.ID()); e != nil {
		return e
	}
	e := &Edge{F: from.(*Node), T: to.(*Node)}
	g.SetEdge(e)
	return e
}

// DOTAttributers returns the global DOT attributes for the graph.
func (g *Graph) DOTAttributers() (graph, node, edge encoding.Attributer) {
	return g.GraphAttrs, g.NodeAttrs, g.EdgeAttrs
}

// NodeMap returns a mapping of ID integers to nodes in the graph.
func (g *Graph) NodeMap() map[int64]*Node {
	nodes := make(map[int64]*Node)
	for _, n := range graph.NodesOf(g.Nodes()) {
		nodes[n.ID()] = n.(*Node)
	}
	return nodes
}

// Node is a graph node able to handle DOT attributes.
type Node struct {
	NodeID int64
	Name   string
	Attributes
}

// ID returns the ID of a node.
func (n *Node) ID() int64 { return n.NodeID }

// DOTID returns the node's DOT ID.
func (n *Node) DOTID() string {
	return n.Name
}

// SetDOTID sets the node's DOT ID.
func (n *Node) SetDOTID(id string) {
	n.Name = id
}

// Edge is a graph edge able to handle DOT attributes.
type Edge struct {
	F, T *Node
	Attributes
}

// From returns the 'from' node of an edge.
func (e *Edge) From() graph.Node { return e.F }

// To returns the 'to' node of an edge.
func (e *Edge) To() graph.Node { return e.T }

// Attributes is a type to help handle DOT attributes.
type Attributes []encoding.Attribute

// Get returns the value of the given attribute. If the attribute is not
// set, the empty string is returned.
func (a Attributes) Get(attr string) string {
	for _, kv := range a {
		if kv.Key == attr {
			return kv.Value
		}
	}
	return ""
}

// Attributes returns the complete list of attributes.
func (a Attributes) Attributes() []encoding.Attribute {
	return a
}

// Set sets the given attribute to the specified value. If the attr Value
// field is the empty string, the attribute is unset.
func (a *Attributes) SetAttribute(attr encoding.Attribute) error {
	for i, kv := range *a {
		if kv.Key == attr.Key {
			if attr.Value != "" {
				(*a)[i].Value = attr.Value
			} else {
				(*a)[i], *a = (*a)[len(*a)-1], (*a)[:len(*a)-1]
			}
			return nil
		}
	}
	*a = append(*a, attr)
	return nil
}

// DOTAttributes returns the DOT attributes for the receiver.
func (a Attributes) DOTAttributes() []encoding.Attribute { return []encoding.Attribute(a) }

// Induce returns a subgraph based on g that contains only the nodes in by,
// and edges that have both ends in by.
func Induce(g *Graph, by []*Node) graph.Graph {
	i := nodeInducedGraph{
		Graph: g,
		nodes: make(map[int64]bool, len(by)),
	}
	for _, n := range by {
		i.nodes[n.ID()] = true
	}
	return i
}

type nodeInducedGraph struct {
	*Graph
	nodes map[int64]bool
}

func (g nodeInducedGraph) Node(id int64) graph.Node {
	if !g.nodes[id] {
		return nil
	}
	return g.Graph.Node(id)
}

func (g nodeInducedGraph) Nodes() graph.Nodes {
	n := graph.NodesOf(g.Graph.Nodes())
	for i := 0; i < len(n); {
		if !g.nodes[n[i].ID()] {
			n[i], n = n[len(n)-1], n[:len(n)-1]
		} else {
			i++
		}
	}
	return iterator.NewOrderedNodes(n)
}

func (g nodeInducedGraph) Edges() graph.Edges {
	e := graph.EdgesOf(g.Graph.Edges())
	for i := 0; i < len(e); {
		if !g.nodes[e[i].From().ID()] || !g.nodes[e[i].To().ID()] {
			e[i], e = e[len(e)-1], e[:len(e)-1]
		} else {
			i++
		}
	}
	return iterator.NewOrderedEdges(e)
}

func (g nodeInducedGraph) From(id int64) graph.Nodes {
	n := graph.NodesOf(g.Graph.From(id))
	for i := 0; i < len(n); {
		if !g.nodes[n[i].ID()] {
			n[i], n = n[len(n)-1], n[:len(n)-1]
		} else {
			i++
		}
	}
	return iterator.NewOrderedNodes(n)
}

func (g nodeInducedGraph) HasEdgeBetween(xid, yid int64) bool {
	if !g.nodes[xid] || !g.nodes[yid] {
		return false
	}
	return g.Graph.HasEdgeBetween(xid, yid)
}

func (g nodeInducedGraph) Edge(uid, vid int64) graph.Edge {
	return g.EdgeBetween(uid, vid)
}

func (g nodeInducedGraph) EdgeBetween(xid, yid int64) graph.Edge {
	if !g.nodes[xid] || !g.nodes[yid] {
		return nil
	}
	return g.Graph.EdgeBetween(xid, yid)
}

// DOT renders the graph as a DOT language representation.
func DOT(g graph.Graph) string {
	b, _ := dot.Marshal(g, "", "", "  ")
	return string(b)
}

// Draw renders the graph as an SVG using the GraphViz command in format.
// The format parameter can be one of "dot", "neato", "fdp" and "sfdp".
// See https://www.graphviz.org/ for a description of these commands.
func Draw(g graph.Graph, format string) (string, error) {
	switch format {
	case "dot", "neato", "fdp", "sfdp":
	default:
		return "", fmt.Errorf("invalid format: %q", format)
	}
	path, err := exec.LookPath(format)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(path, "-Tsvg", "-Gsize=10!")
	cmd.Stdin = strings.NewReader(DOT(g))
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	return buf.String(), err
}
