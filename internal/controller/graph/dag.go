/*
Copyright ApeCloud, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package graph

import (
	"errors"
	"fmt"
)

type DAG struct {
	vertices map[Vertex]Vertex
	edges    map[Edge]Edge
}

type Vertex interface{}

type Edge interface {
	From() Vertex
	To() Vertex
}

type realEdge struct {
	F, T Vertex
}

// WalkFunc defines the action should be taken when we walk through the DAG.
// the func is vertex basis
type WalkFunc func(v Vertex) error

var _ Edge = &realEdge{}

func (r *realEdge) From() Vertex {
	return r.F
}

func (r *realEdge) To() Vertex {
	return r.T
}

// AddVertex put 'v' into 'd'
func (d *DAG) AddVertex(v Vertex) bool {
	if v == nil {
		return false
	}
	d.vertices[v] = v
	return true
}

// RemoveVertex delete 'v' from 'd'
// the in&out edges are also deleted
func (d *DAG) RemoveVertex(v Vertex) bool {
	if v == nil {
		return true
	}
	for k := range d.edges {
		if k.From() == v || k.To() == v {
			delete(d.edges, k)
		}
	}
	delete(d.vertices, v)
	return true
}

// Vertices return all vertices in 'd'
func (d *DAG) Vertices() []Vertex {
	vertices := make([]Vertex, 0)
	for v := range d.vertices {
		vertices = append(vertices, v)
	}
	return vertices
}

// AddEdge put edge 'e' into 'd'
func (d *DAG) AddEdge(e Edge) bool {
	if e.From() == nil || e.To() == nil {
		return false
	}
	for k := range d.edges {
		if k.From() == e.From() && k.To() == e.To() {
			return true
		}
	}
	d.edges[e] = e
	return true
}

// RemoveEdge delete edge 'e'
func (d *DAG) RemoveEdge(e Edge) bool {
	for k := range d.edges {
		if k.From() == e.From() && k.To() == e.To() {
			delete(d.edges, k)
		}
	}
	return true
}

// Connect vertex 'from' to 'to' by a new edge if not exist
func (d *DAG) Connect(from, to Vertex) bool {
	if from == nil || to == nil {
		return false
	}
	for k := range d.edges {
		if k.From() == from && k.To() == to {
			return true
		}
	}
	edge := RealEdge(from, to)
	d.edges[edge] = edge
	return true
}

// WalkTopoOrder walk the DAG 'd' in topology order
func (d *DAG) WalkTopoOrder(walkFunc WalkFunc) error {
	if err := d.validate(); err != nil {
		return err
	}
	orders := d.topologicalOrder(false)
	for _, v := range orders {
		if err := walkFunc(v); err != nil {
			return err
		}
	}
	return nil
}

// WalkReverseTopoOrder walk the DAG 'd' in reverse topology order
func (d *DAG) WalkReverseTopoOrder(walkFunc WalkFunc) error {
	if err := d.validate(); err != nil {
		return err
	}
	orders := d.topologicalOrder(true)
	for _, v := range orders {
		if err := walkFunc(v); err != nil {
			return err
		}
	}
	return nil
}

// Root return root vertex that has no in adjacent.
// our DAG should have one and only one root vertex
func (d *DAG) Root() Vertex {
	roots := make([]Vertex, 0)
	for n := range d.vertices {
		if len(d.inAdj(n)) == 0 {
			roots = append(roots, n)
		}
	}
	if len(roots) != 1 {
		return nil
	}
	return roots[0]
}

// String return a string representation of the DAG in topology order
func (d *DAG) String() string {
	str := "|"
	walkFunc := func(v Vertex) error {
		str += fmt.Sprintf("->%v", v)
		return nil
	}
	if err := d.WalkReverseTopoOrder(walkFunc); err != nil {
		return "->err"
	}
	return str
}

// validate 'd' has single Root and has no cycles
func (d *DAG) validate() error {
	// single Root validation
	root := d.Root()
	if root == nil {
		return errors.New("no single Root found")
	}

	// self-cycle validation
	for e := range d.edges {
		if e.From() == e.To() {
			return fmt.Errorf("self-cycle found: %v", e.From())
		}
	}

	// cycle validation
	// use a DFS func to find cycles
	walked := make(map[Vertex]bool)
	marked := make(map[Vertex]bool)
	var walk func(v Vertex) error
	walk = func(v Vertex) error {
		if walked[v] {
			return nil
		}
		if marked[v] {
			return errors.New("cycle found")
		}

		marked[v] = true
		adjacent := d.outAdj(v)
		for _, vertex := range adjacent {
			if err := walk(vertex); err != nil {
				return err
			}
		}
		marked[v] = false
		walked[v] = true
		return nil
	}
	for v := range d.vertices {
		if err := walk(v); err != nil {
			return err
		}
	}
	return nil
}

// topologicalOrder return a vertex list that is in topology order
// 'd' MUST be a legal DAG
func (d *DAG) topologicalOrder(reverse bool) []Vertex {
	// orders is what we want, a (reverse) topological order of this DAG
	orders := make([]Vertex, 0)

	// walked marks vertex has been walked, to stop recursive func call
	walked := make(map[Vertex]bool)

	// walk is a DFS func
	var walk func(v Vertex)
	walk = func(v Vertex) {
		if walked[v] {
			return
		}
		var adjacent []Vertex
		if reverse {
			adjacent = d.outAdj(v)
		} else {
			adjacent = d.inAdj(v)
		}
		for _, vertex := range adjacent {
			walk(vertex)
		}
		walked[v] = true
		orders = append(orders, v)
	}
	for v := range d.vertices {
		walk(v)
	}

	return orders
}

// outAdj returns all adjacent vertices that v points to
func (d *DAG) outAdj(v Vertex) []Vertex {
	vertices := make([]Vertex, 0)
	for e := range d.edges {
		if e.From() == v {
			vertices = append(vertices, e.To())
		}
	}
	return vertices
}

// inAdj returns all adjacent vertices that point to v
func (d *DAG) inAdj(v Vertex) []Vertex {
	vertices := make([]Vertex, 0)
	for e := range d.edges {
		if e.To() == v {
			vertices = append(vertices, e.From())
		}
	}
	return vertices
}

// NewDAG new an empty DAG
func NewDAG() *DAG {
	dag := &DAG{
		vertices: make(map[Vertex]Vertex),
		edges:    make(map[Edge]Edge),
	}
	return dag
}

func RealEdge(from, to Vertex) Edge {
	return &realEdge{F: from, T: to}
}
