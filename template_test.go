package serum

import (
	"fmt"
	"testing"
)

func TestTemplateParse(t *testing.T) {
	tt := []struct {
		template string
		expect   []parsed
	}{
		{"", []parsed{{literal: ""}}},
		{"a", []parsed{{literal: "a"}}},
		{"a {{", []parsed{{literal: "a {{"}}},
		{"a {{b}}", []parsed{{literal: "a "}, {interp: "b"}}},
		{"a {{b}} c", []parsed{{literal: "a "}, {interp: "b"}, {literal: " c"}}},
		{"a {{}} x {{b}} c", []parsed{{literal: "a "}, {literal: "{{}}"}, {literal: " x "}, {interp: "b"}, {literal: " c"}}}, // didn't bother to implement literal collapsing.
		{"{{{{}}", []parsed{{interp: "{{"}}},                    // not important behavior, other than that it doesn't error -- but, note that it's not a recursive parser.
		{"{{{{}}}}", []parsed{{interp: "{{"}, {literal: "}}"}}}, // not important behavior, other than that it doesn't error -- but, it's still not a recursive parser.
		{"{{x}}}", []parsed{{interp: "x"}, {literal: "}"}}},
		{"a {{ b }}", []parsed{{literal: "a "}, {interp: "b"}}},
		{"{{b}} a", []parsed{{interp: "b"}, {literal: " a"}}},
		{"{{b}}", []parsed{{interp: "b"}}},
		{"{{b|q}}", []parsed{{interp: "b", process: "q"}}},
		{"{{b | q}}", []parsed{{interp: "b", process: "q"}}},
		{"{{ b | q }}", []parsed{{interp: "b", process: "q"}}},
	}
	for _, test := range tt {
		result := parse(test.template)
		resultPrint := fmt.Sprintf("%#v", result)
		expectPrint := fmt.Sprintf("%#v", test.expect)
		if resultPrint != expectPrint {
			t.Errorf("mismatch:\n\tresult: %s\n\texpect: %s", resultPrint, expectPrint)
		}
	}
}

func TestTemplateInterpolate(t *testing.T) {
	tt := []struct {
		template []parsed
		table    [][2]string
		expect   string
	}{
		{
			[]parsed{{literal: ""}},
			[][2]string{},
			"",
		},
		{
			[]parsed{{literal: "a"}},
			[][2]string{},
			"a",
		},
		{
			[]parsed{{interp: "a"}},
			[][2]string{},
			"{{a}}",
		},
		{
			[]parsed{{interp: "a"}},
			[][2]string{{"a", "z"}},
			"z",
		},
		{
			[]parsed{{interp: "a"}, {literal: " "}, {interp: "a"}},
			[][2]string{{"a", "z"}},
			"z z",
		},
		{
			[]parsed{{interp: "b"}},
			[][2]string{{"a", "z"}},
			"{{b}}",
		},
	}
	for _, test := range tt {
		result := interpolate(test.template, test.table)
		if result != test.expect {
			t.Errorf("mismatch:\n\tresult: %s\n\texpect: %s", result, test.expect)
		}
	}
}
