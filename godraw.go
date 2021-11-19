package godraw

import (
	"encoding/xml"
	"strings"
)

type GraphModel struct {
	XMLName xml.Name `xml:"mxGraphModel"`
	Dx      int      `xml:"dx,attr"`
	Dy      int      `xml:"dy,attr"`
	Arrows  string   `xml:"arrows,attr,omitempty"`
	Root    []Cell   `xml:"root>mxCell"`
}

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

type Geometry struct {
	XMLName  xml.Name `xml:"mxGeometry"`
	X        int      `xml:"x,attr,omitempty"`
	Y        int      `xml:"y,attr,omitempty"`
	Width    string   `xml:"width,attr,omitempty"`
	Height   string   `xml:"height,attr,omitempty"`
	Relative string   `xml:"relative,attr,omitempty"`
	As       string   `xml:"as,attr"`
}

type Style struct {
	Attributes map[string]string
}

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

func (g *GraphModel) Add(c *Cell) *GraphModel {
	g.Root = append(g.Root, *c)
	return g
}

func NewShape(id, layerId string) *Cell {
	s := newCell(id, layerId)
	s.Vertex = "1"
	s.Geometry = newGeometry()
	return s
}

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

func NewImageXY(id, layerId, url string, x int, y int) *Cell {
	i := NewImage(id, layerId, url)
	i.Geometry.X = x
	i.Geometry.X = y
	return i
}

func newCell(id string, layerId string) *Cell {
	return &Cell{
		ID:       id,
		ParentID: layerId,
	}

}

func newGeometry() *Geometry {
	return &Geometry{
		X:  10,
		Y:  10,
		As: "geometry",
	}
}
