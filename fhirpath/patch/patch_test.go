package patch_test

import (
	"errors"
	"testing"

	"github.com/fhir-fli/fhirpath-go/fhir"
	"github.com/fhir-fli/fhirpath-go/fhirpath"
	"github.com/fhir-fli/fhirpath-go/fhirpath/patch"
	"github.com/fhir-fli/fhirpath-go/internal/element/extension"
	"github.com/fhir-fli/fhirpath-go/internal/element/reference"
	"github.com/fhir-fli/fhirpath-go/pkg/containedresource"
	cpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/codes_go_proto"
	dtpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/datatypes_go_proto"
	bcrpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/bundle_and_contained_resource_go_proto"
	epb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/encounter_go_proto"
	ispb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/imaging_study_go_proto"
	opb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/observation_go_proto"
	ppb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/patient_go_proto"
	rgpb "github.com/google/fhir/go/proto/google/fhir/proto/r4/core/resources/request_group_go_proto"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
)

var patientWithBirthDate = &ppb.Patient{
	BirthDate: fhir.MustParseDate("1993-05-16"),
}

func TestAdd_ValidInputs_ModifiesResource(t *testing.T) {
	patientRef, _ := reference.Typed("Patient", "123")

	testCases := []struct {
		name  string
		path  string
		field string
		input fhir.Resource
		value fhir.Base
		want  fhir.Resource
	}{
		{
			name:  "Adds scalar field",
			path:  "Patient",
			field: "birthDate",
			input: &ppb.Patient{},
			value: fhir.MustParseDate("1993-05-16"),
			want: &ppb.Patient{
				BirthDate: fhir.MustParseDate("1993-05-16"),
			},
		}, {
			name:  "Adds scalar field with reserved name",
			path:  "Encounter",
			field: "class",
			input: &epb.Encounter{},
			value: fhir.Coding("", ""),
			want: &epb.Encounter{
				ClassValue: fhir.Coding("", ""),
			},
		}, {
			name:  "Adds non-enum string field",
			path:  "Patient.maritalStatus",
			field: "text",
			input: &ppb.Patient{
				MaritalStatus: &dtpb.CodeableConcept{},
			},
			value: fhir.String("H0H0H0"),
			want: &ppb.Patient{
				MaritalStatus: fhir.CodeableConcept("H0H0H0"),
			},
		}, {
			name:  "Adds enum field",
			path:  "Patient",
			field: "gender",
			input: &ppb.Patient{},
			value: fhir.String("male"),
			want: &ppb.Patient{
				Gender: &ppb.Patient_GenderCode{
					Value: cpb.AdministrativeGenderCode_MALE,
				},
			},
		}, {
			name:  "Adds valid integer to positiveInt field",
			path:  "Patient.telecom[0]",
			field: "rank",
			input: &ppb.Patient{
				Telecom: []*dtpb.ContactPoint{{}},
			},
			value: fhir.Integer(1),
			want: &ppb.Patient{
				Telecom: []*dtpb.ContactPoint{
					{
						Rank: fhir.PositiveInt(1),
					},
				},
			},
		}, {
			name:  "Adds negative integer to integer field",
			path:  "Patient.extension[0]",
			field: "value",
			input: &ppb.Patient{
				Extension: []*dtpb.Extension{
					{
						Url: fhir.URI(""),
					},
				},
			},
			value: fhir.Integer(-10),
			want: &ppb.Patient{
				Extension: []*dtpb.Extension{
					extension.New("", fhir.Integer(-10)),
				},
			},
		}, {
			name:  "Adds integer to unsigned integer field",
			path:  "ImagingStudy",
			field: "numberOfSeries",
			input: &ispb.ImagingStudy{},
			value: fhir.Integer(0),
			want: &ispb.ImagingStudy{
				NumberOfSeries: fhir.UnsignedInt(0),
			},
		}, {
			name:  "Appends extension field",
			path:  "Patient",
			field: "extension",
			input: &ppb.Patient{},
			value: extension.New("", fhir.String("hello world")),
			want: &ppb.Patient{
				Extension: []*dtpb.Extension{
					extension.New("", fhir.String("hello world")),
				},
			},
		},
		{
			name:  "Adds reference field",
			path:  "Observation",
			field: "subject",
			input: &opb.Observation{},
			value: patientRef,
			want: &opb.Observation{
				Subject: patientRef,
			},
		},
		{
			name:  "Adds id to existing reference field",
			path:  "Observation.subject",
			field: "patientId",
			input: &opb.Observation{
				Subject: &dtpb.Reference{
					Type: fhir.URI("Patient"),
				},
			},
			value: fhir.String("123"),
			want: &opb.Observation{
				Subject: patientRef,
			},
		},
		{
			name:  "Adds extension oneof field",
			path:  "Patient.extension[0]",
			field: "value",
			input: &ppb.Patient{
				Extension: []*dtpb.Extension{
					{},
				},
			},
			value: fhir.String("hello world"),
			want: &ppb.Patient{
				Extension: []*dtpb.Extension{
					{
						Value: &dtpb.Extension_ValueX{
							Choice: &dtpb.Extension_ValueX_StringValue{
								StringValue: fhir.String("hello world"),
							},
						},
					},
				},
			},
		}, {
			name:  "Adds contained resource oneof field",
			path:  "Bundle.entry[0]",
			field: "resource",
			input: &bcrpb.Bundle{
				Entry: []*bcrpb.Bundle_Entry{
					{},
				},
			},
			value: &ppb.Patient{},
			want: &bcrpb.Bundle{
				Entry: []*bcrpb.Bundle_Entry{
					{
						Resource: containedresource.Wrap(&ppb.Patient{}),
					},
				},
			},
		}, {
			name:  "Appends bundle entry",
			path:  "Bundle",
			field: "entry",
			input: &bcrpb.Bundle{
				Entry: []*bcrpb.Bundle_Entry{
					{},
				},
			},
			value: &bcrpb.Bundle_Entry{
				Resource: containedresource.Wrap(&ppb.Patient{}),
			},
			want: &bcrpb.Bundle{
				Entry: []*bcrpb.Bundle_Entry{
					{},
					{
						Resource: containedresource.Wrap(&ppb.Patient{}),
					},
				},
			},
		}, {
			name:  "Setting start field of RequestGroup extension period",
			path:  "RequestGroup.extension.where(url='123').value as FHIR.Period",
			field: "start",
			input: &rgpb.RequestGroup{
				Extension: []*dtpb.Extension{
					{},
					{
						Url: fhir.URI("123"),
						Value: &dtpb.Extension_ValueX{
							Choice: &dtpb.Extension_ValueX_Period{
								Period: &dtpb.Period{},
							},
						},
					},
					{},
				},
			},
			value: fhir.MustParseDateTime("2006-01-02T15:04:05Z"),
			want: &rgpb.RequestGroup{
				Extension: []*dtpb.Extension{
					{},
					{
						Url: fhir.URI("123"),
						Value: &dtpb.Extension_ValueX{
							Choice: &dtpb.Extension_ValueX_Period{
								Period: &dtpb.Period{
									Start: fhir.MustParseDateTime("2006-01-02T15:04:05Z"),
								},
							},
						},
					},
					{},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Add(tc.input, tc.path, tc.field, tc.value, &patch.Options{})
			if err != nil {
				t.Fatalf("Add(%s): unexpected err = %v", tc.name, err)
			}

			got, want := tc.input, tc.want
			if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
				t.Errorf("Add(%s): (-got +want):\n%v", tc.name, diff)
			}
		})
	}
}

func TestAdd_InvalidInputs(t *testing.T) {
	testCases := []struct {
		name    string
		path    string
		field   string
		input   fhir.Resource
		value   fhir.Base
		wantErr error
	}{
		{
			name:    "Invalid text case",
			path:    "Patient",
			field:   "birth_date",
			input:   &ppb.Patient{},
			value:   fhir.MustParseDate("1993-05-16"),
			wantErr: patch.ErrInvalidField,
		}, {
			name:    "Underlying evaluation error",
			path:    "Patient.i_dont_exist",
			field:   "thisDoesntMatter",
			input:   &ppb.Patient{},
			value:   fhir.String(""),
			wantErr: patch.ErrInvalidField,
		}, {
			name:    "Field does not exist",
			path:    "Patient",
			field:   "badField",
			input:   &ppb.Patient{},
			value:   fhir.MustParseDate("1993-05-16"),
			wantErr: patch.ErrInvalidField,
		}, {
			name:  "Non-singleton result",
			path:  "Patient.name",
			field: "family",
			input: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{},
					{},
				},
			},
			value:   &dtpb.HumanName{},
			wantErr: patch.ErrNotSingleton,
		}, {
			name:    "enum value with bad casing",
			path:    "Patient",
			field:   "gender",
			input:   &ppb.Patient{},
			value:   fhir.String("MALE"),
			wantErr: patch.ErrInvalidEnum,
		}, {
			name:    "Invalid enum value",
			path:    "Patient",
			field:   "gender",
			input:   &ppb.Patient{},
			value:   fhir.String("not_a_gender"),
			wantErr: patch.ErrInvalidEnum,
		}, {
			name:  "Invalid int for positiveInt field",
			path:  "Patient.telecom[0]",
			field: "rank",
			input: &ppb.Patient{
				Telecom: []*dtpb.ContactPoint{{}},
			},
			value:   fhir.Integer(-1),
			wantErr: patch.ErrInvalidUnsignedInt,
		}, {
			name:  "Unpatchable result",
			path:  "Patient.active.value",
			field: "something",
			input: &ppb.Patient{
				Active: fhir.Boolean(true),
			},
			value:   fhir.Boolean(false),
			wantErr: patch.ErrNotPatchable,
		}, {
			name:  "Field already exists",
			path:  "Patient",
			field: "active",
			input: &ppb.Patient{
				Active: fhir.Boolean(true),
			},
			value:   fhir.Boolean(false),
			wantErr: patch.ErrNotPatchable,
		}, {
			name:    "Wrong input type",
			path:    "Patient",
			field:   "active",
			input:   &ppb.Patient{},
			value:   fhir.String("true"),
			wantErr: patch.ErrInvalidInput,
		}, {
			name:  "Invalid oneof entry",
			path:  "Bundle.entry[0]",
			field: "resource",
			input: &bcrpb.Bundle{
				Entry: []*bcrpb.Bundle_Entry{
					{},
				},
			},
			value:   fhir.String("I am not a resource"),
			wantErr: patch.ErrInvalidInput,
		}, {
			name:  "Nil replacement value",
			path:  "Bundle.entry[0]",
			field: "resource",
			input: &bcrpb.Bundle{
				Entry: []*bcrpb.Bundle_Entry{
					{},
				},
			},
			value:   nil,
			wantErr: patch.ErrInvalidInput,
		}, {
			name:    "Nil input resource value",
			path:    "Bundle.entry[0]",
			field:   "resource",
			input:   nil,
			value:   fhir.String(""),
			wantErr: patch.ErrInvalidInput,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Add(tc.input, tc.path, tc.field, tc.value, &patch.Options{})

			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Add(%s): got err '%v', want err '%v'", tc.name, got, want)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name string
		res  fhir.Resource
		path string
		want fhir.Resource
	}{
		{
			name: "Deletes scalar field",
			res: &ppb.Patient{
				BirthDate: fhir.MustParseDate("1993-05-16"),
			},
			path: "Patient.birthDate",
			want: &ppb.Patient{},
		},
		{
			name: "Deletes single entry from end of list",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Betty"), fhir.String("Sue")},
					},
				},
			},
			path: "Patient.name.given[1]",
			want: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Betty")},
					},
				},
			},
		},
		{
			name: "Deletes single entry from beginning of list",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Betty"), fhir.String("Sue")},
					},
				},
			},
			path: "Patient.name.given[0]",
			want: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Sue")},
					},
				},
			},
		},
		{
			name: "Deletes list containing single entry",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Betty")},
					},
				},
			},
			path: "Patient.name.given",
			want: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{},
				},
			},
		},
		{
			name: "No-ops on empty but valid field",
			res:  &ppb.Patient{},
			path: "Patient.birthDate",
			want: &ppb.Patient{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Delete(tc.res, tc.path)
			if err != nil {
				t.Fatalf("Delete(%v): got unexpected err = %v", tc.name, err)
			}

			got, want := tc.res, tc.want
			if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
				t.Errorf("Delete(%s): (-got +want):\n%v", tc.name, diff)
			}
		})
	}
}

func TestDelete_BadInput_ReturnsError(t *testing.T) {
	testCases := []struct {
		name    string
		res     fhir.Resource
		path    string
		wantErr error
	}{
		{
			name:    "Nil input",
			res:     nil,
			path:    "Patient.birthDate",
			wantErr: patch.ErrInvalidInput,
		},
		{
			name:    "Evaluation fails",
			res:     &ppb.Patient{},
			path:    "Patient.no_exist",
			wantErr: fhirpath.ErrInvalidField,
		},
		{
			name: "Attempting to delete more than one value",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Jieun"), fhir.String("IU")},
					},
				},
			},
			path:    "Patient.name.given",
			wantErr: patch.ErrNotSingleton,
		},
		{
			name: "Attempting to delete primitive value",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Jieun"), fhir.String("IU")},
					},
				},
			},
			path:    "Patient.name.given[0].value",
			wantErr: patch.ErrNotPatchable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Delete(tc.res, tc.path)

			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("Delete(%s): got error %v, want %v", tc.name, got, want)
			}
		})
	}
}

func TestInsert(t *testing.T) {
	testCases := []struct {
		name  string
		res   fhir.Resource
		path  string
		value fhir.Base
		index int
		want  fhir.Resource
	}{
		{
			name: "Inserts name at beginning",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU")},
					},
				},
			},
			path:  "Patient.name[0].given",
			value: fhir.String("Jieun"),
			index: 0,
			want: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Jieun"), fhir.String("IU")},
					},
				},
			},
		}, {
			name: "Inserts name at end",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU")},
					},
				},
			},
			path:  "Patient.name[0].given",
			value: fhir.String("Jieun"),
			index: 1,
			want: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU"), fhir.String("Jieun")},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Insert(tc.res, tc.path, tc.value, tc.index)
			if err != nil {
				t.Fatalf("Insert(%v): got unexpected err = %v", tc.name, err)
			}

			got, want := tc.res, tc.want
			if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
				t.Errorf("Insert(%s): (-got +want):\n%v", tc.name, diff)
			}
		})
	}
}

func TestInsert_InvalidCondition_ReturnsError(t *testing.T) {
	testCases := []struct {
		name    string
		res     fhir.Resource
		path    string
		value   fhir.Base
		index   int
		wantErr error
	}{
		{
			name:    "Nil Input",
			res:     nil,
			path:    "Patient.name[0].given",
			value:   fhir.String("Jieun"),
			index:   0,
			wantErr: patch.ErrInvalidInput,
		},
		{
			name:    "Evaluation fails",
			res:     &ppb.Patient{},
			path:    "Patient.no_exist",
			value:   fhir.String("Jieun"),
			index:   0,
			wantErr: fhirpath.ErrInvalidField,
		},
		{
			name: "Ambiguous insertion",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Jonathan")},
					},
					{
						Given: []*dtpb.String{fhir.String("Jon")},
					},
				},
			},
			path:    "Patient.name.given",
			value:   fhir.String("Jonny-Boy"),
			index:   0,
			wantErr: patch.ErrNotSingleton,
		},
		{
			name: "Extraction type is not FHIR type",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Jonathan")},
					},
				},
			},
			path:    "Patient.name.given.value.toString()",
			value:   fhir.String("Jonny-Boy"),
			index:   0,
			wantErr: patch.ErrNotPatchable,
		},
		{
			name: "Output is disconnected value",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("Jonathan")},
					},
				},
			},
			path:    "Patient.name.given.now()",
			value:   fhir.String("Jonny-Boy"),
			index:   0,
			wantErr: patch.ErrNotPatchable,
		},
		{
			name: "Insert target is not a list",
			res: &ppb.Patient{
				BirthDate: fhir.DateNow(),
			},
			path:    "Patient.birthDate.value",
			value:   fhir.DateNow(),
			index:   0,
			wantErr: patch.ErrNotPatchable,
		},
		{
			name: "Index is out of range",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU")},
					},
				},
			},
			path:    "Patient.name[0].given",
			value:   fhir.String("Jieun"),
			index:   42,
			wantErr: patch.ErrNotPatchable,
		},
		{
			name: "Index is negative",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU")},
					},
				},
			},
			path:    "Patient.name[0].given",
			value:   fhir.String("Jieun"),
			index:   -1,
			wantErr: patch.ErrNotPatchable,
		},
		{
			name: "Input value is wrong type",
			res: &ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU")},
					},
				},
			},
			path:    "Patient.name[0].given",
			value:   fhir.ID("Jieun"),
			index:   0,
			wantErr: patch.ErrNotPatchable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Insert(tc.res, tc.path, tc.value, tc.index)

			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("Insert(%s): error got = %v, want = %v", tc.name, got, want)
			}
		})
	}
}

func TestMove(t *testing.T) {
	testCases := []struct {
		name     string
		res      fhir.Resource
		path     string
		srcIndex int
		dstIndex int
		wantRes  fhir.Resource
		wantErr  error
	}{
		{
			"moves name",
			&ppb.Patient{
				Name: []*dtpb.HumanName{
					{
						Given: []*dtpb.String{fhir.String("IU"), fhir.String("Jieun")},
					},
				},
			},
			"Patient.name[0].given",
			0,
			1,
			nil,
			patch.ErrNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Move(tc.res, tc.path, tc.srcIndex, tc.dstIndex)

			if got, want := err, tc.wantErr; !errors.Is(got, want) {
				t.Fatalf("Move(%s): error got = %v, want = %v", tc.name, got, want)
			}

			got, want := fhir.Resource(nil), tc.wantRes
			if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
				t.Errorf("Move(%s): (-got +want):\n%v", tc.name, diff)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	testCases := []struct {
		name    string
		res     fhir.Resource
		path    string
		value   any
		wantRes fhir.Resource
		wantErr error
	}{
		{
			"replaces birthDate",
			patientWithBirthDate,
			"Patient.birthDate",
			"2007-07-05",
			nil,
			patch.ErrNotImplemented,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := patch.Replace(tc.res, tc.path, tc.value)

			if got, want := err, tc.wantErr; !errors.Is(got, want) {
				t.Fatalf("Replace(%s): error got = %v, want = %v", tc.name, got, want)
			}

			got, want := fhir.Resource(nil), tc.wantRes
			if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
				t.Errorf("Replace(%s): (-got +want):\n%v", tc.name, diff)
			}
		})
	}
}
