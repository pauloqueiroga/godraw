package godraw_test

import (
	"encoding/xml"
	"errors"
	"maps"
	"strings"
	"testing"

	"github.com/pauloqueiroga/godraw"
)

// styleAttr is the attribute name encoding/xml passes to the Style codec
// for the Cell.Style field.
var styleAttr = xml.Name{Local: "style"}

// validStyle reports whether attrs follows the rules documented on
// godraw.Style: keys are non-empty and contain neither ";" nor "=", and
// values do not contain ";".
func validStyle(attrs map[string]string) bool {
	for k, v := range attrs {
		if k == "" || strings.ContainsAny(k, ";=") || strings.Contains(v, ";") {
			return false
		}
	}
	return true
}

// FuzzStyleRoundTrip encodes a style made of two pairs and decodes the
// result. A style that breaks the rules on godraw.Style must be rejected
// with an error wrapping ErrInvalidStyle. Any other style must decode to
// exactly the pairs that were encoded.
//
// The codec is called directly rather than through encoding/xml, so the
// test covers the style format alone and not XML escaping.
func FuzzStyleRoundTrip(f *testing.F) {
	seeds := []struct{ k1, v1, k2, v2 string }{
		{"shape", "image", "image", gopherURL},
		{"image", "data:image/svg+xml,PHN2Zz4=", "html", "1"}, // "=" in a value
		{"text", "", "html", "1"},                             // key without a value
		{"rounded", "1", "rounded", "1"},                      // the same key twice
		{"image", gopherURL + ";rounded=1", "shape", "image"}, // injected key
		{"fillColor", "#dae8fc", "strokeColor", "#6c8ebf;fillColor=#ff0000"},
		{"", "", "html", "1"},  // empty key
		{"a=b", "c", "d", "e"}, // "=" in a key
		{"a;b", "c", "d", "e"}, // ";" in a key
	}
	for _, s := range seeds {
		f.Add(s.k1, s.v1, s.k2, s.v2)
	}

	f.Fuzz(func(t *testing.T, k1, v1, k2, v2 string) {
		in := godraw.Style{Attributes: map[string]string{k1: v1, k2: v2}}

		attr, err := in.MarshalXMLAttr(styleAttr)
		if !validStyle(in.Attributes) {
			if !errors.Is(err, godraw.ErrInvalidStyle) {
				t.Fatalf("MarshalXMLAttr(%q) = %q, %v; want an error wrapping ErrInvalidStyle",
					in.Attributes, attr.Value, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("MarshalXMLAttr(%q) rejected a valid style: %v", in.Attributes, err)
		}
		if attr.Name != styleAttr {
			t.Fatalf("MarshalXMLAttr(%q) returned attribute name %+v, want %+v",
				in.Attributes, attr.Name, styleAttr)
		}

		var out godraw.Style
		if err := out.UnmarshalXMLAttr(attr); err != nil {
			t.Fatalf("UnmarshalXMLAttr(%q): %v", attr.Value, err)
		}
		if !maps.Equal(out.Attributes, in.Attributes) {
			t.Errorf("style %q encoded as %q decoded to %q", in.Attributes, attr.Value, out.Attributes)
		}
	})
}

// FuzzStyleDecode decodes arbitrary style text, like a style read from a
// file that draw.io or a person wrote. Whatever it decodes to must encode
// without an error and decode back to the same style.
func FuzzStyleDecode(f *testing.F) {
	for _, s := range []string{
		"",
		"rounded=1;whiteSpace=wrap;html=1;",
		"text;html=1",
		"image=data:image/svg+xml,PHN2Zz4=;",
		"image=https://example.com/a.svg?a=1&b=2;",
		"shape=image;shape=ellipse;",
		";;",
		"=",
		"=1;rounded=1",
		"a==b;",
	} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, s string) {
		var first godraw.Style
		if err := first.UnmarshalXMLAttr(xml.Attr{Name: styleAttr, Value: s}); err != nil {
			t.Fatalf("UnmarshalXMLAttr(%q): %v", s, err)
		}

		attr, err := first.MarshalXMLAttr(styleAttr)
		if err != nil {
			t.Fatalf("style %q decoded from %q cannot be encoded: %v", first.Attributes, s, err)
		}

		var second godraw.Style
		if err := second.UnmarshalXMLAttr(attr); err != nil {
			t.Fatalf("UnmarshalXMLAttr(%q): %v", attr.Value, err)
		}
		if !maps.Equal(second.Attributes, first.Attributes) {
			t.Errorf("%q decoded to %q, which encoded as %q and decoded to %q",
				s, first.Attributes, attr.Value, second.Attributes)
		}
	})
}
