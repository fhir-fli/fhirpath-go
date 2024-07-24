package fhir_test

import (
	"testing"

	"github.com/fhir-fli/fhirpath-go/fhir"
	"github.com/google/go-cmp/cmp"
)

func TestNarrative(t *testing.T) {
	want := "<blah></blah>"

	sut := fhir.Narrative(want)

	if got := sut.GetDiv().GetValue(); !cmp.Equal(got, want) {
		t.Errorf("Narrative: got %v, want %v", got, want)
	}
}

func TestXHTML(t *testing.T) {
	want := "<blah></blah>"

	sut := fhir.XHTML(want)

	if got := sut.GetValue(); !cmp.Equal(got, want) {
		t.Errorf("XHTML: got %v, want %v", got, want)
	}
}
