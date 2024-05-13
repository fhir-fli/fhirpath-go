package funcs_test

import (
	"errors"
	"reflect"
	"testing"

	ppb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/verily-src/fhirpath-go/internal/fhir"
	"github.com/verily-src/fhirpath-go/fhirpath/internal/expr"
	"github.com/verily-src/fhirpath-go/fhirpath/internal/expr/exprtest"
	"github.com/verily-src/fhirpath-go/fhirpath/internal/funcs"
	"github.com/verily-src/fhirpath-go/fhirpath/system"
)

func TestToFunction_EvaluatesCorrectly(t *testing.T) {
	patient := &ppb.Patient{
		Id: fhir.ID("1234"),
	}
	fns := map[string]any{
		"take": func(input system.Collection, num system.Integer) (system.Collection, error) {
			result := system.Collection{}
			for i := 0; i < int(num); i++ {
				if i >= len(input) {
					continue
				}
				result = append(result, input[i])
			}
			return result, nil
		},
		"findResource": func(input system.Collection, resource fhir.Resource) (system.Collection, error) {
			for i, elem := range input {
				if reflect.DeepEqual(elem, resource) {
					return system.Collection{system.Integer(i)}, nil
				}
			}
			return system.Collection{}, nil
		},
	}
	testCases := []struct {
		name  string
		fn    any
		args  []expr.Expression
		input system.Collection
		want  system.Collection
	}{
		{
			name:  "test custom take function",
			fn:    fns["take"],
			args:  []expr.Expression{&expr.LiteralExpression{Literal: system.Integer(2)}},
			input: system.Collection{"1", "2", "3"},
			want:  system.Collection{"1", "2"},
		},
		{
			name:  "findResource returns index of desired resource",
			fn:    fns["findResource"],
			args:  []expr.Expression{exprtest.Return(patient)},
			input: system.Collection{system.Boolean(true), system.Boolean(false), patient},
			want:  system.Collection{system.Integer(2)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotFunc, err := funcs.ToFunction(tc.fn)
			if err != nil {
				t.Fatalf("ToFunction(%T) raised unexpected invalid signature error: %v", tc.fn, err)
			}
			gotCollection, err := gotFunc.Func(&expr.Context{}, tc.input, tc.args...)
			if err != nil {
				t.Fatalf("Evaluating function generated by ToFunction raised unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, gotCollection); diff != "" {
				t.Errorf("Evaluating function generated by ToFunction returned unexpected diff (-want, +got)\n%s", diff)
			}
		})
	}
}

func TestToFunction_RaisesSignatureError(t *testing.T) {
	testCases := []struct {
		name string
		fn   any
	}{
		{
			name: "not a function",
			fn:   4,
		},
		{
			name: "no arguments",
			fn:   func() {},
		},
		{
			name: "doesn't contain an input collection as first argument",
			fn:   func(num system.Integer) {},
		},
		{
			name: "only returns one input",
			fn:   func(in system.Collection) system.Collection { return system.Collection{} },
		},
		{
			name: "doesn't return a collection",
			fn:   func(in system.Collection) (int, error) { return 1, nil },
		},
		{
			name: "doesn't return an error",
			fn:   func(in system.Collection) (system.Collection, bool) { return system.Collection{}, false },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := funcs.ToFunction(tc.fn)
			if err == nil {
				t.Fatalf("ToFunction(%T) didn't raise error when expected to on function signature mismatch", tc.fn)
			}
		})
	}
}

func TestToFunction_RaisesEvaluationError(t *testing.T) {
	testCases := []struct {
		name  string
		fn    any
		args  []expr.Expression
		input system.Collection
	}{
		{
			name:  "function arity doesn't match number of arguments",
			fn:    func(in system.Collection) (system.Collection, error) { return system.Collection{}, nil },
			args:  []expr.Expression{&expr.LiteralExpression{Literal: system.Boolean(true)}},
			input: system.Collection{},
		},
		{
			name: "argument expression raises error",
			fn: func(in system.Collection, num system.Integer) (system.Collection, error) {
				return system.Collection{}, nil
			},
			args:  []expr.Expression{exprtest.Error(errors.New("mock error"))},
			input: system.Collection{},
		},
		{
			name: "argument expression doesn't evaluate to singleton",
			fn: func(in system.Collection, num system.Integer) (system.Collection, error) {
				return system.Collection{}, nil
			},
			args:  []expr.Expression{exprtest.Return(1, 2)},
			input: system.Collection{},
		},
		{
			name: "argument expression evaluates to different type",
			fn: func(in system.Collection, num system.Integer) (system.Collection, error) {
				return system.Collection{}, nil
			},
			args:  []expr.Expression{&expr.LiteralExpression{Literal: system.Boolean(false)}},
			input: system.Collection{},
		},
		{
			name: "second argument expression evaluates to wrong type",
			fn: func(in system.Collection, num system.Integer, num2 system.Integer) (system.Collection, error) {
				return system.Collection{}, nil
			},
			args:  []expr.Expression{&expr.LiteralExpression{Literal: system.Integer(3)}, &expr.LiteralExpression{Literal: system.Boolean(true)}},
			input: system.Collection{},
		},
		{
			name:  "function returns error",
			fn:    func(in system.Collection) (system.Collection, error) { return nil, errors.New("some error") },
			args:  []expr.Expression{},
			input: system.Collection{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotFunc, err := funcs.ToFunction(tc.fn)
			if err != nil {
				t.Fatalf("ToFunction(%T) raised unexpected invalid signature error: %v", tc.fn, err)
			}
			_, err = gotFunc.Func(&expr.Context{}, tc.input, tc.args...)
			if err == nil {
				t.Fatalf("ToFunction() did not raise evaluation error when calling generated function")
			}
		})
	}
}