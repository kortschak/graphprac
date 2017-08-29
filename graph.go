// Copyright ©2017 Dan Kortschak. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package graphprac

import (
	"io/ioutil"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/encoding"
	"gonum.org/v1/gonum/graph/encoding/dot"
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
	if e := g.Edge(from, to); e != nil {
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
	for _, n := range g.Nodes() {
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

// SetAttribute sets a single DOT attribute.
func (n *Node) SetAttribute(attr encoding.Attribute) error {
	n.Attributes = append(n.Attributes, attr)
	return nil
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
// set, the emtpy string is returned.
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
func (a *Attributes) SetAttribute(attr encoding.Attribute) {
	for i, kv := range *a {
		if kv.Key == attr.Key {
			if attr.Value != "" {
				(*a)[i].Value = attr.Value
			} else {
				(*a)[i], *a = (*a)[len(*a)-1], (*a)[:len(*a)-1]
			}
			return
		}
	}
	*a = append(*a, attr)
}

// DOTAttributes returns the DOT attributes for the receiver.
func (a Attributes) DOTAttributes() []encoding.Attribute { return []encoding.Attribute(a) }
