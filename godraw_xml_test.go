package godraw_test

import (
	"encoding/xml"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/pauloqueiroga/godraw"
)

const (
	layer     = "1"
	gopherURL = "https://example.com/gopher.svg"

	// rootCells is the XML of the root cell and the layer that
	// NewGraph(layer) creates.
	rootCells = `<mxCell id="root"></mxCell><mxCell id="1" parent="root"></mxCell>`

	// marshalRuns is how many times each value is marshalled. Style is a
	// map, so an encoder that follows map iteration order gives different
	// output on different runs. Repeating the marshal makes that failure
	// show up reliably instead of now and then.
	marshalRuns = 20
)

// graph returns a graph with one layer and the given cells added to it.
func graph(cells ...*godraw.Cell) godraw.GraphModel {
	g := godraw.NewGraph(layer)
	for _, c := range cells {
		g.Add(c)
	}
	return g
}

// graphXML returns the XML of a default graph (as made by graph) whose
// layer is followed by the given cells.
func graphXML(cells string) string {
	return `<mxGraphModel dx="640" dy="480"><root>` + rootCells + cells + `</root></mxGraphModel>`
}

type graphCase struct {
	name  string
	model godraw.GraphModel
	want  string
}

// graphCases returns fresh models with their expected XML. Each call builds
// new values, so a test cannot affect another through shared pointers.
func graphCases() []graphCase {
	return []graphCase{
		{
			name:  "empty graph",
			model: graph(),
			want:  graphXML(""),
		},
		{
			name: "arrows attribute",
			model: func() godraw.GraphModel {
				g := graph()
				g.Arrows = "1"
				return g
			}(),
			want: `<mxGraphModel dx="640" dy="480" arrows="1"><root>` + rootCells + `</root></mxGraphModel>`,
		},
		{
			name: "shape with label",
			model: func() godraw.GraphModel {
				b1 := godraw.NewShape("b1", layer)
				b1.Value = "Play with this package"
				return graph(b1)
			}(),
			want: graphXML(`<mxCell id="b1" value="Play with this package" parent="1" vertex="1">` +
				`<mxGeometry x="10" y="10" as="geometry"></mxGeometry></mxCell>`),
		},
		{
			name: "shape with size",
			model: func() godraw.GraphModel {
				b1 := godraw.NewShape("b1", layer)
				b1.Geometry.Width = "120"
				b1.Geometry.Height = "60"
				return graph(b1)
			}(),
			want: graphXML(`<mxCell id="b1" parent="1" vertex="1">` +
				`<mxGeometry x="10" y="10" width="120" height="60" as="geometry"></mxGeometry></mxCell>`),
		},
		{
			name: "label is XML-escaped",
			model: func() godraw.GraphModel {
				b1 := godraw.NewShape("b1", layer)
				b1.Value = "<b>Tom & \"Jerry\"</b>\nsecond line"
				return graph(b1)
			}(),
			want: graphXML(`<mxCell id="b1" value="&lt;b&gt;Tom &amp; &#34;Jerry&#34;&lt;/b&gt;&#xA;second line" parent="1" vertex="1">` +
				`<mxGeometry x="10" y="10" as="geometry"></mxGeometry></mxCell>`),
		},
		{
			name:  "image",
			model: graph(godraw.NewImage("img", layer, gopherURL)),
			want: graphXML(`<mxCell id="img" style="image=https://example.com/gopher.svg;imageAspect=0;shape=image;" parent="1" vertex="1">` +
				`<mxGeometry x="10" y="10" as="geometry"></mxGeometry></mxCell>`),
		},
		{
			name:  "image URL with query string",
			model: graph(godraw.NewImage("img", layer, gopherURL+"?size=64&v=2")),
			want: graphXML(`<mxCell id="img" style="image=https://example.com/gopher.svg?size=64&amp;v=2;imageAspect=0;shape=image;" parent="1" vertex="1">` +
				`<mxGeometry x="10" y="10" as="geometry"></mxGeometry></mxCell>`),
		},
		{
			name:  "image at coordinates",
			model: graph(godraw.NewImageXY("img", layer, gopherURL, 470, 20)),
			want: graphXML(`<mxCell id="img" style="image=https://example.com/gopher.svg;imageAspect=0;shape=image;" parent="1" vertex="1">` +
				`<mxGeometry x="470" y="20" as="geometry"></mxGeometry></mxCell>`),
		},
		{
			name: "two shapes joined by an edge",
			model: func() godraw.GraphModel {
				a := godraw.NewShape("a", layer)
				b := godraw.NewShape("b", layer)
				b.Geometry.X = 200
				e := &godraw.Cell{
					ID:       "e1",
					ParentID: layer,
					Edge:     "1",
					SourceID: "a",
					TargetID: "b",
					Style:    godraw.Style{Attributes: map[string]string{"html": "1", "endArrow": "classic"}},
					Geometry: &godraw.Geometry{Relative: "1", As: "geometry"},
				}
				return graph(a, b, e)
			}(),
			want: graphXML(`<mxCell id="a" parent="1" vertex="1"><mxGeometry x="10" y="10" as="geometry"></mxGeometry></mxCell>` +
				`<mxCell id="b" parent="1" vertex="1"><mxGeometry x="200" y="10" as="geometry"></mxGeometry></mxCell>` +
				`<mxCell id="e1" style="endArrow=classic;html=1;" parent="1" edge="1" source="a" target="b">` +
				`<mxGeometry relative="1" as="geometry"></mxGeometry></mxCell>`),
		},
	}
}

func TestMarshalGraph(t *testing.T) {
	for _, tc := range graphCases() {
		t.Run(tc.name, func(t *testing.T) {
			marshalAndCheck(t, tc.model, tc.want)
		})
	}
}

func TestRoundTripGraph(t *testing.T) {
	for _, tc := range graphCases() {
		t.Run(tc.name, func(t *testing.T) {
			blob, err := xml.Marshal(tc.model)
			if err != nil {
				t.Fatalf("xml.Marshal: %v", err)
			}

			var got godraw.GraphModel
			if err := xml.Unmarshal(blob, &got); err != nil {
				t.Fatalf("xml.Unmarshal: %v\ninput: %s", err, blob)
			}
			clearXMLNames(&got)

			if !reflect.DeepEqual(got, tc.model) {
				t.Errorf("round trip changed the model\nXML: %s\n got: %s\nwant: %s",
					blob, describeGraph(got), describeGraph(tc.model))
			}
		})
	}
}

// TestNewImageXYGeometry checks the coordinates NewImageXY sets. It marshals
// only the geometry, so problems with the style attribute cannot mask the
// result.
func TestNewImageXYGeometry(t *testing.T) {
	tests := []struct {
		name string
		x, y int
		want string
	}{
		{
			name: "x and y differ",
			x:    470,
			y:    20,
			want: `<mxGeometry x="470" y="20" as="geometry"></mxGeometry>`,
		},
		{
			name: "x and y equal",
			x:    300,
			y:    300,
			want: `<mxGeometry x="300" y="300" as="geometry"></mxGeometry>`,
		},
		{
			name: "x is zero",
			x:    0,
			y:    20,
			want: `<mxGeometry y="20" as="geometry"></mxGeometry>`,
		},
		{
			name: "y is zero",
			x:    470,
			y:    0,
			want: `<mxGeometry x="470" as="geometry"></mxGeometry>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := godraw.NewImageXY("img", layer, gopherURL, tt.x, tt.y)
			if img.Geometry == nil {
				t.Fatal("NewImageXY returned a cell without a geometry")
			}
			marshalAndCheck(t, img.Geometry, tt.want)
		})
	}
}

func TestMarshalStyle(t *testing.T) {
	tests := []struct {
		name  string
		attrs map[string]string
		want  string // XML of a cell with ID "a" and the given style
	}{
		{
			name:  "nil map omits the attribute",
			attrs: nil,
			want:  `<mxCell id="a"></mxCell>`,
		},
		{
			name:  "empty map omits the attribute",
			attrs: map[string]string{},
			want:  `<mxCell id="a"></mxCell>`,
		},
		{
			name:  "single pair",
			attrs: map[string]string{"rounded": "1"},
			want:  `<mxCell id="a" style="rounded=1;"></mxCell>`,
		},
		{
			name: "keys in sorted order",
			attrs: map[string]string{
				"whiteSpace":  "wrap",
				"rounded":     "1",
				"html":        "1",
				"fillColor":   "#dae8fc",
				"strokeColor": "#6c8ebf",
			},
			want: `<mxCell id="a" style="fillColor=#dae8fc;html=1;rounded=1;strokeColor=#6c8ebf;whiteSpace=wrap;"></mxCell>`,
		},
		{
			name:  "key without a value",
			attrs: map[string]string{"text": "", "html": "1"},
			want:  `<mxCell id="a" style="html=1;text;"></mxCell>`,
		},
		{
			name:  "value containing an equals sign",
			attrs: map[string]string{"image": "data:image/svg+xml,PHN2Zz4="},
			want:  `<mxCell id="a" style="image=data:image/svg+xml,PHN2Zz4=;"></mxCell>`,
		},
		{
			name:  "value needing XML escaping",
			attrs: map[string]string{"image": "https://example.com/a.svg?a=1&b=2"},
			want:  `<mxCell id="a" style="image=https://example.com/a.svg?a=1&amp;b=2;"></mxCell>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := godraw.Cell{ID: "a", Style: godraw.Style{Attributes: tt.attrs}}
			blob := marshalAndCheck(t, cell, tt.want)

			var got godraw.Cell
			if err := xml.Unmarshal(blob, &got); err != nil {
				t.Fatalf("xml.Unmarshal: %v\ninput: %s", err, blob)
			}
			if !maps.Equal(got.Style.Attributes, tt.attrs) {
				t.Errorf("round trip of %s gave style %q, want %q", blob, got.Style.Attributes, tt.attrs)
			}
		})
	}
}

func TestUnmarshalStyle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "attribute missing",
			input: `<mxCell id="a"/>`,
			want:  nil,
		},
		{
			name:  "attribute empty",
			input: `<mxCell id="a" style=""/>`,
			want:  nil,
		},
		{
			name:  "trailing semicolon",
			input: `<mxCell id="a" style="rounded=1;html=1;"/>`,
			want:  map[string]string{"rounded": "1", "html": "1"},
		},
		{
			name:  "no trailing semicolon",
			input: `<mxCell id="a" style="rounded=1;html=1"/>`,
			want:  map[string]string{"rounded": "1", "html": "1"},
		},
		{
			name:  "empty segments",
			input: `<mxCell id="a" style=";rounded=1;;html=1;"/>`,
			want:  map[string]string{"rounded": "1", "html": "1"},
		},
		{
			name:  "key without a value",
			input: `<mxCell id="a" style="text;html=1;"/>`,
			want:  map[string]string{"text": "", "html": "1"},
		},
		{
			name:  "value containing an equals sign",
			input: `<mxCell id="a" style="image=data:image/svg+xml,PHN2Zz4=;"/>`,
			want:  map[string]string{"image": "data:image/svg+xml,PHN2Zz4="},
		},
		{
			name:  "XML-escaped value",
			input: `<mxCell id="a" style="image=https://example.com/a.svg?a=1&amp;b=2;"/>`,
			want:  map[string]string{"image": "https://example.com/a.svg?a=1&b=2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got godraw.Cell
			if err := xml.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("xml.Unmarshal: %v", err)
			}
			// maps.Equal treats a nil map and an empty map as equal.
			if !maps.Equal(got.Style.Attributes, tt.want) {
				t.Errorf("style = %q, want %q", got.Style.Attributes, tt.want)
			}
		})
	}
}

// TestStyleValueCannotInjectKeys checks that a style value containing ";"
// cannot add keys to the style or change another key's value. The encoder
// may reject such a value or escape it. If it encodes the value, decoding
// the result must give back exactly the original keys, and every key whose
// value has no ";" must keep its value. The value that holds the ";" is not
// compared, because an escaping encoder may legitimately change it.
func TestStyleValueCannotInjectKeys(t *testing.T) {
	tests := []struct {
		name string
		cell *godraw.Cell
	}{
		{
			name: "key with a value",
			cell: godraw.NewImage("img", layer, gopherURL+";rounded=1"),
		},
		{
			name: "key without a value",
			cell: godraw.NewImage("img", layer, gopherURL+";html"),
		},
		{
			// The injected "shape" is written right after "image". The real
			// "shape" sorts after "image", so it comes later and wins on
			// decode. This case alone cannot detect an override.
			name: "overrides a key that sorts later",
			cell: godraw.NewImage("img", layer, gopherURL+";shape=ellipse"),
		},
		{
			// The injected "fillColor" is written right after "strokeColor".
			// The real "fillColor" sorts before "strokeColor", so the
			// injected value comes later and wins on decode.
			name: "overrides a key that sorts earlier",
			cell: &godraw.Cell{ID: "a", Style: godraw.Style{Attributes: map[string]string{
				"fillColor":   "#dae8fc",
				"strokeColor": "#6c8ebf;fillColor=#ff0000",
			}}},
		},
		{
			name: "trailing semicolon",
			cell: godraw.NewImage("img", layer, gopherURL+";"),
		},
		{
			name: "semicolon only",
			cell: godraw.NewImage("img", layer, ";"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blob, err := xml.Marshal(tt.cell)
			if err != nil {
				t.Logf("xml.Marshal rejected the value: %v", err)
				return
			}

			var got godraw.Cell
			if err := xml.Unmarshal(blob, &got); err != nil {
				t.Fatalf("xml.Unmarshal: %v\ninput: %s", err, blob)
			}

			want := tt.cell.Style.Attributes
			gotKeys := slices.Sorted(maps.Keys(got.Style.Attributes))
			wantKeys := slices.Sorted(maps.Keys(want))
			if !slices.Equal(gotKeys, wantKeys) {
				t.Errorf("style keys after round trip = %q, want %q\nXML: %s", gotKeys, wantKeys, blob)
			}
			for _, k := range wantKeys {
				if strings.Contains(want[k], ";") {
					continue
				}
				if v := got.Style.Attributes[k]; v != want[k] {
					t.Errorf("%s after round trip = %q, want %q\nXML: %s", k, v, want[k], blob)
				}
			}
		})
	}
}

// marshalAndCheck encodes v marshalRuns times and reports an error unless
// every run gives want. It returns the encoded bytes of the last run.
func marshalAndCheck(t *testing.T, v any, want string) []byte {
	t.Helper()

	var blob []byte
	for run := range marshalRuns {
		var err error
		blob, err = xml.Marshal(v)
		if err != nil {
			t.Fatalf("xml.Marshal: %v", err)
		}
		if got := string(blob); got != want {
			t.Errorf("xml.Marshal (run %d of %d):\n got: %s\nwant: %s", run+1, marshalRuns, got, want)
			break
		}
	}
	return blob
}

// clearXMLNames resets the XMLName fields that xml.Unmarshal fills in,
// so a decoded model can be compared with one built in Go.
func clearXMLNames(g *godraw.GraphModel) {
	g.XMLName = xml.Name{}
	for i := range g.Root {
		g.Root[i].XMLName = xml.Name{}
		if g.Root[i].Geometry != nil {
			g.Root[i].Geometry.XMLName = xml.Name{}
		}
	}
}

// describeGraph formats a model for failure messages, following the
// Geometry pointers that %v would print as addresses.
func describeGraph(g godraw.GraphModel) string {
	var b strings.Builder
	fmt.Fprintf(&b, "{Dx:%d Dy:%d Arrows:%q}", g.Dx, g.Dy, g.Arrows)
	for _, c := range g.Root {
		geometry := "nil"
		if c.Geometry != nil {
			geometry = fmt.Sprintf("{X:%d Y:%d Width:%q Height:%q Relative:%q As:%q}",
				c.Geometry.X, c.Geometry.Y, c.Geometry.Width, c.Geometry.Height, c.Geometry.Relative, c.Geometry.As)
		}
		fmt.Fprintf(&b, "\n  {ID:%q Value:%q Style:%q Parent:%q Vertex:%q Edge:%q Source:%q Target:%q Geometry:%s}",
			c.ID, c.Value, c.Style.Attributes, c.ParentID, c.Vertex, c.Edge, c.SourceID, c.TargetID, geometry)
	}
	return b.String()
}
