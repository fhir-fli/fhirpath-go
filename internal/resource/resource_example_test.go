package resource_test

import (
	"fmt"

	"github.com/fhir-fli/fhirpath-go/fhir"
	"github.com/fhir-fli/fhirpath-go/internal/resource"
	dtpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	"github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
)

func ExampleGetIdentifierList() {
	patient := &patient_go_proto.Patient{
		Id: fhir.ID("12345"),
		Identifier: []*dtpb.Identifier{
			&dtpb.Identifier{
				System: &dtpb.Uri{Value: "http://fake.com"},
				Value:  &dtpb.String{Value: "9efbf82d-7a58-4d14-bec1-63f8fda148a8"},
			},
		},
	}

	ids, err := resource.GetIdentifierList(patient)
	if err != nil {
		panic(err)
	} else if ids == nil || len(ids) == 0 {
		panic("no identifiers")
	} else {
		fmt.Printf("Identifier value: %#v", ids[0].GetValue().Value)
		// Output: Identifier value: "9efbf82d-7a58-4d14-bec1-63f8fda148a8"
	}
}
