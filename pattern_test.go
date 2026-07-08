package cs2log

import (
	"errors"
	"regexp"
	"strings"
	"testing"
)

func TestCompileDefinitionsRejectsInvalidDefinitions(t *testing.T) {
	tests := []struct {
		name       string
		definition definition
		want       string
	}{
		{
			name:       "missing type",
			definition: definition{Regex: `ok`},
			want:       "definition type must not be empty",
		},
		{
			name:       "missing regex",
			definition: definition{Type: "Broken"},
			want:       "definition regex must not be empty",
		},
		{
			name:       "invalid regex",
			definition: definition{Type: "Broken", Regex: `(`},
			want:       "compile Broken pattern",
		},
		{
			name: "capture count mismatch",
			definition: definition{
				Type:   "Broken",
				Regex:  `(one) (two)`,
				Fields: []field{{Name: "one", Type: "string"}},
			},
			want: "fields must match regex captures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := compileDefinitions([]definition{tt.definition})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestValidateDefinitionFieldsRejectsMismatchedCaptureMetadata(t *testing.T) {
	def := definition{
		Type: "Example",
		Fields: []field{
			{Name: "first", Type: "string"},
			{Name: "second", Type: "int"},
		},
	}

	tests := []struct {
		name     string
		captures []compiledField
		want     string
	}{
		{
			name:     "count",
			captures: []compiledField{{name: "first", typ: "string"}},
			want:     "fields must match regex captures",
		},
		{
			name:     "name",
			captures: []compiledField{{name: "other", typ: "string"}, {name: "second", typ: "int"}},
			want:     `field 0 = "first", want "other"`,
		},
		{
			name:     "type",
			captures: []compiledField{{name: "first", typ: "string"}, {name: "second", typ: "float"}},
			want:     `Example.second type = "int", want "float"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefinitionFields(def, tt.captures)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestEventMatcherReturnsNoMatch(t *testing.T) {
	matcher := eventMatcher{
		definition: definition{Type: "Example"},
		regexp:     mustCompileTestRegexp(t, `^hello$`),
	}

	event, ok, err := matcher.match(BaseEvent{Type: "Example"}, "goodbye")
	if event != nil || ok || err != nil {
		t.Fatalf("event=%#v ok=%t err=%v, want no match", event, ok, err)
	}
}

func TestEventMatcherWrapsBuildError(t *testing.T) {
	buildErr := errors.New("bad build")
	matcher := eventMatcher{
		definition: definition{
			Type: "Example",
			Build: func(BaseEvent, Matches) (Event, error) {
				return nil, buildErr
			},
		},
		regexp: mustCompileTestRegexp(t, `^(hello)$`),
		fields: []compiledField{{name: "value", typ: "string"}},
	}

	event, ok, err := matcher.match(BaseEvent{Type: "Example"}, "hello")
	if event != nil || ok || !errors.Is(err, buildErr) || !strings.Contains(err.Error(), "build Example") {
		t.Fatalf("event=%#v ok=%t err=%v", event, ok, err)
	}
}

func TestMatchesReturnConversionErrors(t *testing.T) {
	matches := Matches{values: map[string]string{
		"player": "not a player token",
		"int":    "not-an-int",
		"float":  "not-a-float",
	}}

	if _, err := matches.Player("player"); err == nil {
		t.Fatal("Player succeeded with invalid token")
	}
	if _, err := matches.Int("int"); err == nil {
		t.Fatal("Int succeeded with invalid value")
	}
	if _, err := matches.Float64("float"); err == nil {
		t.Fatal("Float64 succeeded with invalid value")
	}
}

func mustCompileTestRegexp(t *testing.T, pattern string) *regexp.Regexp {
	t.Helper()
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("compile regexp: %v", err)
	}
	return re
}
