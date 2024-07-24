package fhirwrapper

import (
	"fmt"

	"github.com/fhir-fli/fhirpath-go/fhir"
	"github.com/fhir-fli/fhirpath-go/fhirpath"
	"github.com/fhir-fli/fhirpath-go/pkg/containedresource"
	bcrpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
)

// CompileFHIRPath compiles a FHIRPath expression.
func CompileFHIRPath(expression string) (*fhirpath.Expression, error) {
	return fhirpath.Compile(expression)
}

// EvaluateFHIRPath evaluates a FHIRPath expression against a FHIR resource.
func EvaluateFHIRPath(compiledExpr *fhirpath.Expression, resource *bcrpb.ContainedResource) ([]interface{}, error) {
	result, err := compiledExpr.Evaluate([]fhir.Resource{containedresource.Unwrap(resource)})
	if err != nil {
		return nil, fmt.Errorf("error while evaluating FHIRPath: %w", err)
	}
	return result, nil
}
