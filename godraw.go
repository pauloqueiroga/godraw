// Copyright 2021 Paulo Queiroga. All rights reserved.
// Use of this source code is governed by the license that can
// be found in the LICENSE file.

// Package godraw implements basic types and helper functions to
// work with mxGraph Model.
package godraw

import (
	"encoding/xml"
	"strings"
)

// A GraphModel implements a wrapper around the cells which are
// in charge of storing the actual graph data structure.
//
// The model must have a top-level root cell which contains the
// layers (typically one layer is enough). All cells with Parent
// ID pointing to the root cell is a layer. All other cells shall
// have parent IDs pointing to layers, not the root cell.
type GraphModel struct {
	XMLName xml.Name `xml:"mxGraphModel"`
	Dx      int      `xml:"dx,attr"`
	Dy      int      `xml:"dy,attr"`
	Arrows  string   `xml:"arrows,attr,omitempty"`
	Root    []Cell   `xml:"root>mxCell"`
}

// A Cell represents an element of the graph model. A cell can be
// a layer (its ParentID points to the root cell's ID);
// a vertex in a graph (Vertex="1");
// or an edge in a graph (Edge="1").
type Cell struct {
	XMLName  xml.Name `xml:"mxCell"`
	ID       string   `xml:"id,attr"`
	Value    string   `xml:"value,attr,omitempty"`
	Style    Style    `xml:"style,attr,omitempty"`
	ParentID string   `xml:"parent,attr,omitempty"`
	Vertex   string   `xml:"vertex,attr,omitempty"`
	Edge     string   `xml:"edge,attr,omitempty"`
	SourceID string   `xml:"source,attr,omitempty"`
	TargetID string   `xml:"target,attr,omitempty"`
	Geometry *Geometry
}

// A Geometry can be slightly different for vertices or edges.
// When used with vertices, it carries the x and y coordinates,
// the width, and the height of the vertex.
// For edges, contains optional terminal and control points.
// In most cases, the "As" attribute will have the value
// "geometry".
type Geometry struct {
	XMLName  xml.Name `xml:"mxGeometry"`
	X        int      `xml:"x,attr,omitempty"`
	Y        int      `xml:"y,attr,omitempty"`
	Width    string   `xml:"width,attr,omitempty"`
	Height   string   `xml:"height,attr,omitempty"`
	Relative string   `xml:"relative,attr,omitempty"`
	As       string   `xml:"as,attr"`
}

// A Style is a map of key-value pairs to describe the style
// properties of each cell.
type Style struct {
	Attributes map[string]string
}

// MarshalXMLAttr returns an XML attribute with the encoded value
// of Style. It implements xml.MarshalerAttr interface.
func (a Style) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	var text string

	for k, v := range a.Attributes {
		text += k

		if v != "" {
			text += "="
			text += v
		}

		text += ";"
	}

	return xml.Attr{Name: xml.Name{Local: "style"}, Value: text}, nil
}

// UnmarshalXMLAttr decodes a single XML attribute of type Style.
// It implements xml.UnmarshalerAttr interface.
func (a *Style) UnmarshalXMLAttr(attr xml.Attr) error {
	a.Attributes = make(map[string]string)
	pairs := strings.Split(attr.Value, ";")

	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) < 2 {
			kv = append(kv, "")
		}
		a.Attributes[kv[0]] = kv[1]
	}

	return nil
}

// NewGraph returns a new graph model containing a root cell and
// one layer with ID layerId.
func NewGraph(layerId string) GraphModel {
	return GraphModel{
		Dx: 640,
		Dy: 480,
		Root: []Cell{
			{ID: "root"},
			{
				ID:       layerId,
				ParentID: "root",
			},
		},
	}
}

// Add adds the given Cell to the root cell of the receiving
// graph model.
func (g *GraphModel) Add(c *Cell) *GraphModel {
	g.Root = append(g.Root, *c)
	return g
}

// NewShape returns a new Vertex Cell, configured with the given
// unique ID (id) and parent ID (layerId). The new cell contains
// a default geometry which you might want to change.
func NewShape(id, layerId string) *Cell {
	s := newCell(id, layerId)
	s.Vertex = "1"
	s.Geometry = newGeometry()
	return s
}

// NewImage returns a new Vertex Cell, configured as an image,
// with given unique ID (id), parent ID (layerId) and image
// source URL (url). The new cell contains a default geometry
// which you might want to change.
func NewImage(id, layerId, url string) *Cell {
	i := NewShape(id, layerId)
	i.Style = Style{
		Attributes: map[string]string{
			"shape":       "image",
			"imageAspect": "0",
			"image":       url,
		},
	}
	return i
}

// NewImageXY returns a new Vertex Cell, configured as an image,
// with given unique ID (id), parent ID (layerId), image source
// URL (url), and coordinates (x and y).
func NewImageXY(id, layerId, url string, x int, y int) *Cell {
	i := NewImage(id, layerId, url)
	i.Geometry.X = x
	i.Geometry.X = y
	return i
}

// newCell returns a new Cell object configured with id and
// parent ID.
func newCell(id string, layerId string) *Cell {
	return &Cell{
		ID:       id,
		ParentID: layerId,
	}

}

// newGeometry returns a new Geometry object configured with
// default values.
func newGeometry() *Geometry {
	return &Geometry{
		X:  10,
		Y:  10,
		As: "geometry",
	}
}
