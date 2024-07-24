package impl_test

import (
	"testing"
	"time"

	"github.com/fhir-fli/fhirpath-go/fhirpath/internal/expr"
	"github.com/fhir-fli/fhirpath-go/fhirpath/internal/funcs/impl"
	"github.com/fhir-fli/fhirpath-go/fhirpath/system"
	"github.com/google/go-cmp/cmp"
)

func TestTimeOfDay(t *testing.T) {
	ctx := &expr.Context{Now: time.Date(0, time.January, 1, 19, 30, 5, 1000000, time.UTC)}
	wantCollection := system.Collection{system.MustParseTime("19:30:05.001")}

	got, err := impl.TimeOfDay(ctx, []any{})
	if err != nil {
		t.Fatalf("impl.TimeOfDay() returned unexpected error: %v", err)
	}
	if !cmp.Equal(got, wantCollection) {
		t.Errorf("impl.TimeOfDay() returned unexpected result: got %v, want %v", got, wantCollection)
	}
}

func TestToday(t *testing.T) {
	ctx := &expr.Context{Now: time.Date(2010, time.February, 12, 0, 0, 0, 0, time.UTC)}
	wantCollection := system.Collection{system.MustParseDate("2010-02-12")}

	got, err := impl.Today(ctx, []any{})
	if err != nil {
		t.Fatalf("impl.Today() returned unexpected error: %v", err)
	}
	if !cmp.Equal(got, wantCollection) {
		t.Errorf("impl.Today() returned unexpected result: got %v, want %v", got, wantCollection)
	}
}

func TestNow(t *testing.T) {
	ctx := &expr.Context{Now: time.Date(2010, time.February, 12, 12, 30, 34, 2000000, time.UTC)}
	wantCollection := system.Collection{system.MustParseDateTime("2010-02-12T12:30:34.002Z")}

	got, err := impl.Now(ctx, []any{})
	if err != nil {
		t.Fatalf("impl.Now() returned unexpected error: %v", err)
	}
	if !cmp.Equal(got, wantCollection) {
		t.Errorf("impl.Now() returned unexpected result: got %v, want %v", got, wantCollection)
	}
}
