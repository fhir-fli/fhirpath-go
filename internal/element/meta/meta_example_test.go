package meta_test

import (
	"fmt"

	"github.com/fhir-fli/fhirpath-go/fhir"
	"github.com/fhir-fli/fhirpath-go/internal/element/canonical"
	"github.com/fhir-fli/fhirpath-go/internal/element/meta"
	dtpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
)

func ExampleUpdate() {
	m := &dtpb.Meta{}

	meta.Update(m,
		meta.WithTags(fhir.Coding("urn:oid:verily/sample-tag-system", "sample-tag-value")),
		meta.WithProfiles(canonical.New("urn:oid:verily/sample-profile")),
	)

	fmt.Printf("meta.profile: %q\n", m.Profile[0].Value)
	fmt.Printf("meta.tag: {%q, %q}", m.Tag[0].System.Value, m.Tag[0].Code.Value)
	// Output:
	// meta.profile: "urn:oid:verily/sample-profile"
	// meta.tag: {"urn:oid:verily/sample-tag-system", "sample-tag-value"}
}
