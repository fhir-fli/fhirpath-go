package fhir_test

import (
	"fmt"
	"testing"

	"github.com/fhir-fli/fhirpath-go/fhir"
)

func TestEscapeSearchParam(t *testing.T) {

	testCases := []struct {
		input string
		want  string
	}{
		{``, ``},
		{`\`, `\\`},
		{`$`, `\$`},
		{`,`, `\,`},
		{`|`, `\|`},
		{`C:\bin\go foo, bar, baz | omg $500!`, `C:\\bin\\go foo\, bar\, baz \| omg \$500!`},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("testCases[%d]", i), func(t *testing.T) {
			got := fhir.EscapeSearchParam(tc.input)

			if got != tc.want {
				t.Errorf("got %#v, want %#v", got, tc.want)
			}
		})
	}
}
