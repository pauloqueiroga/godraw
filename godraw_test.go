package godraw

import (
	"encoding/xml"
	"testing"
)

func TestExampleEncodeAndDecodeXml(t *testing.T) {
	const l = "1" // We'll name the only layer "1"
	g := NewGraph(l)
	g.Add(
		NewImageXY(
			"gopher1",
			l,
			"https://raw.githubusercontent.com/golang-samples/gopher-vector/9fe99fbf17b019125bf649f8a921882b54e151a6/gopher.svg",
			470,
			10))
	b1 := NewShape("b1", l)
	b1.Value = "Play with this package"
	g.Add(b1)

	// XML-Marshalling into blob
	blob, err := xml.Marshal(g)
	if err != nil {
		t.Fatal(err)
	}

	// Un-marshaling the XML into object
	var result GraphModel
	if err := xml.Unmarshal(blob, &result); err != nil {
		t.Fatal(err)
	}

	// Verifying if some basics match between the original object and the un-marshalled one
	if g.Dx != result.Dx {
		t.Errorf("Dx should be %d found %d", g.Dx, result.Dx)
	}
	if g.Dy != result.Dy {
		t.Errorf("Dy should be %d found %d", g.Dy, result.Dy)
	}
	if len(g.Root) != len(result.Root) {
		t.Errorf("Root should have %d found %d", len(g.Root), len(result.Root))
	}
	for _, c := range g.Root {
		found := false
		for _, r := range result.Root {
			if r.ID == c.ID {
				found = true
				if c.Edge != r.Edge {
					t.Errorf("%s edge should be %s found %s", c.ID, c.Edge, r.Edge)
				}
				if c.Vertex != r.Vertex {
					t.Errorf("%s vertex should be %s found %s", c.ID, c.Vertex, r.Vertex)
				}
				if c.ParentID != r.ParentID {
					t.Errorf("%s parent should be %s found %s", c.ID, c.ParentID, r.ParentID)
				}
				if c.Value != r.Value {
					t.Errorf("%s value should be %s found %s", c.ID, c.Value, r.Value)
				}
				break
			}
		}

		if !found {
			t.Errorf("Couldn't find cell with ID %s", c.ID)
		}
	}
}
