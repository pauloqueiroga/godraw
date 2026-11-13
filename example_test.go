package godraw_test

import (
	"encoding/xml"
	"fmt"
	"log"

	"github.com/pauloqueiroga/godraw"
)

// This example builds a diagram with an image and a labelled shape,
// then encodes it as the XML that draw.io reads.
func Example() {
	// Every cell sits on a layer. NewGraph creates the root cell and one
	// layer with the ID given here.
	const layerID = "1"
	g := godraw.NewGraph(layerID)

	gopher := godraw.NewImageXY("gopher", layerID,
		"https://raw.githubusercontent.com/golang-samples/gopher-vector/9fe99fbf17b019125bf649f8a921882b54e151a6/gopher.svg",
		470, 10)
	g.Add(gopher)

	// Add stores a copy of the cell, so set its fields before adding it.
	label := godraw.NewShape("label", layerID)
	label.Value = "Play with this package"
	label.Geometry.Width = "160"
	label.Geometry.Height = "40"
	g.Add(label)

	out, err := xml.MarshalIndent(g, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(out))

	// Output:
	// <mxGraphModel dx="640" dy="480">
	//   <root>
	//     <mxCell id="root"></mxCell>
	//     <mxCell id="1" parent="root"></mxCell>
	//     <mxCell id="gopher" style="image=https://raw.githubusercontent.com/golang-samples/gopher-vector/9fe99fbf17b019125bf649f8a921882b54e151a6/gopher.svg;imageAspect=0;shape=image;" parent="1" vertex="1">
	//       <mxGeometry x="470" y="10" as="geometry"></mxGeometry>
	//     </mxCell>
	//     <mxCell id="label" value="Play with this package" parent="1" vertex="1">
	//       <mxGeometry x="10" y="10" width="160" height="40" as="geometry"></mxGeometry>
	//     </mxCell>
	//   </root>
	// </mxGraphModel>
}
