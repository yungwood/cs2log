package cs2log

import (
	"fmt"
	"regexp"
)

type eventMatcher struct {
	definition definition
	regexp     *regexp.Regexp
	fields     []compiledField
}

type compiledField struct {
	name string
	typ  string
}

func compileDefinitions(definitions []definition) ([]eventMatcher, error) {
	matchers := make([]eventMatcher, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Type == "" {
			return nil, fmt.Errorf("definition type must not be empty")
		}
		re, fields, err := compileDefinitionPattern(definition)
		if err != nil {
			return nil, fmt.Errorf("compile %s pattern: %w", definition.Type, err)
		}
		if err := validateDefinitionFields(definition, fields); err != nil {
			return nil, err
		}
		matchers = append(matchers, eventMatcher{definition: definition, regexp: re, fields: fields})
	}
	return matchers, nil
}

func compileDefinitionPattern(definition definition) (*regexp.Regexp, []compiledField, error) {
	if definition.Regex == "" {
		return nil, nil, fmt.Errorf("definition regex must not be empty")
	}
	re, err := regexp.Compile("^" + definition.Regex + "$")
	if err != nil {
		return nil, nil, err
	}
	if re.NumSubexp() != len(definition.Fields) {
		return nil, nil, fmt.Errorf("fields must match regex captures")
	}
	fields := make([]compiledField, 0, len(definition.Fields))
	for _, field := range definition.Fields {
		fields = append(fields, compiledField{name: field.Name, typ: field.Type})
	}
	return re, fields, nil
}

func (m eventMatcher) match(base BaseEvent, payload string) (Event, bool, error) {
	result := m.regexp.FindStringSubmatch(payload)
	if result == nil {
		return nil, false, nil
	}
	values := make(map[string]string, len(m.fields))
	for i, field := range m.fields {
		values[field.name] = result[i+1]
	}
	event, err := m.definition.Build(base, Matches{values: values})
	if err != nil {
		return nil, false, fmt.Errorf("build %s: %w", m.definition.Type, err)
	}
	return event, true, nil
}

func validateDefinitionFields(definition definition, captures []compiledField) error {
	if len(captures) != len(definition.Fields) {
		return fmt.Errorf("%s fields must match regex captures", definition.Type)
	}
	for i, capture := range captures {
		field := definition.Fields[i]
		if field.Name != capture.name {
			return fmt.Errorf("%s field %d = %q, want %q", definition.Type, i, field.Name, capture.name)
		}
		if field.Type != capture.typ {
			return fmt.Errorf("%s.%s type = %q, want %q", definition.Type, field.Name, field.Type, capture.typ)
		}
	}
	return nil
}
