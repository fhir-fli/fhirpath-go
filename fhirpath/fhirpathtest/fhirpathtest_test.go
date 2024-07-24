package fhirpathtest_test

import (
	"errors"
	"testing"

	"github.com/fhir-fli/fhirpath-go/fhir"
	"github.com/fhir-fli/fhirpath-go/fhirpath/fhirpathtest"
	"github.com/fhir-fli/fhirpath-go/fhirpath/system"
	"github.com/fhir-fli/fhirpath-go/internal/fhirtest"
	"github.com/fhir-fli/fhirpath-go/internal/resource"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestError_Evaluates_ReturnsErr(t *testing.T) {
	wantErr := errors.New("test error")
	expr := fhirpathtest.Error(wantErr)

	_, err := expr.Evaluate([]fhir.Resource{})

	if got, want := err, wantErr; !errors.Is(got, want) {
		t.Errorf("Error: want err %v, got %v", want, got)
	}
}

func TestReturn_Evaluates_ReturnsCollectionOfEntries(t *testing.T) {
	res := fhirtest.NewResource(t, resource.Patient)
	want := system.Collection{res}
	expr := fhirpathtest.Return(res)

	got, err := expr.Evaluate([]fhir.Resource{})
	if err != nil {
		t.Fatalf("Return: unexpected err: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("Return: mismatched size; want %v, got %v", len(got), len(want))
	}
	if !cmp.Equal(got[0], want[0], protocmp.Transform()) {
		t.Errorf("Return: want %v, got %v", want, got)
	}
}

func TestReturnCollection_Evaluates_ReturnsCollectionOfEntries(t *testing.T) {
	res := fhirtest.NewResource(t, resource.Patient)
	want := system.Collection{res}
	expr := fhirpathtest.ReturnCollection(want)

	got, err := expr.Evaluate([]fhir.Resource{})
	if err != nil {
		t.Fatalf("Return: unexpected err: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("ReturnCollection: mismatched size; want %v, got %v", len(got), len(want))
	}
	if !cmp.Equal(got[0], want[0], protocmp.Transform()) {
		t.Errorf("ReturnCollection: want %v, got %v", want, got)
	}
}
