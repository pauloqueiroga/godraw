// Copyright 2021 Paulo Queiroga.
// Use of this source code is governed by the MIT license that can
// be found in the LICENSE file.

// Package godraw implements basic types and helper functions to
// work with mxGraph Model.
package godraw

import (
	"encoding/xml"
	"errors"
	"fmt"
	"maps"
	"slices"
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
//
// Keys must be non-empty and must not contain ";" or "=". Values
// must not contain ";". draw.io has no way to escape ";" in a style,
// so a value such as a URL must percent-encode it as "%3B".
// Marshalling a style that breaks these rules fails with an error
// that wraps ErrInvalidStyle.
type Style struct {
	Attributes map[string]string
}

// ErrInvalidStyle is wrapped by the error returned when a Style
// cannot be encoded, because a key or value would change the meaning
// of the encoded style.
var ErrInvalidStyle = errors.New("godraw: invalid style")

// MarshalXMLAttr encodes the style as "key=value;" pairs, writing
// the keys in sorted order so the output is the same on every run.
// Empty style results in no style attribute. It implements
// xml.MarshalerAttr interface.
func (a Style) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	if len(a.Attributes) == 0 {
		// encoding/xml leaves out an attribute whose Name is empty.
		return xml.Attr{}, nil
	}

	var b strings.Builder
	for _, k := range slices.Sorted(maps.Keys(a.Attributes)) {
		v := a.Attributes[k]
		switch {
		case k == "":
			return xml.Attr{}, fmt.Errorf("%w: empty key", ErrInvalidStyle)
		case strings.ContainsAny(k, ";="):
			return xml.Attr{}, fmt.Errorf("%w: key %q contains ';' or '='", ErrInvalidStyle, k)
		case strings.Contains(v, ";"):
			return xml.Attr{}, fmt.Errorf("%w: value of key %q contains ';'", ErrInvalidStyle, k)
		}

		b.WriteString(k)
		if v != "" {
			b.WriteByte('=')
			b.WriteString(v)
		}
		b.WriteByte(';')
	}

	return xml.Attr{Name: name, Value: b.String()}, nil
}

// UnmarshalXMLAttr decodes a style made of "key=value" pairs
// separated by ";". Pairs with an empty key are skipped, so every
// decoded style can be encoded again. It implements
// xml.UnmarshalerAttr interface.
func (a *Style) UnmarshalXMLAttr(attr xml.Attr) error {
	a.Attributes = make(map[string]string)
	for pair := range strings.SplitSeq(attr.Value, ";") {
		k, v, _ := strings.Cut(pair, "=")
		if k == "" {
			continue
		}
		a.Attributes[k] = v
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
//
// The URL goes into the cell's Style, so it must not contain ";".
// Percent-encode it as "%3B", or marshalling the cell will fail.
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
	i.Geometry.Y = y
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
