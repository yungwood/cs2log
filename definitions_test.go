package cs2log

import (
	"testing"
	"time"
)

func TestDefinitionsCompile(t *testing.T) {
	if _, err := compileDefinitions(defaultDefinitions); err != nil {
		t.Fatalf("compile definitions: %v", err)
	}
}

func TestDefinitionExamplesParseToDocumentedUTC(t *testing.T) {
	for _, definition := range defaultDefinitions {
		for _, example := range definition.Examples {
			parser, err := NewParser(Config{LogTimezone: example.Timezone})
			if err != nil {
				t.Fatalf("%s example timezone %q: %v", definition.Type, example.Timezone, err)
			}
			event, err := parser.ParseLine(example.Line)
			if err != nil {
				t.Fatalf("%s example parse: %v", definition.Type, err)
			}
			want, err := time.Parse(time.RFC3339, example.UTC)
			if err != nil {
				t.Fatalf("%s example UTC %q: %v", definition.Type, example.UTC, err)
			}
			if !event.Timestamp().Equal(want) {
				t.Fatalf("%s example timestamp = %s, want %s", definition.Type, event.Timestamp().Format(time.RFC3339Nano), want.Format(time.RFC3339Nano))
			}
			if event.EventType() != definition.Type {
				t.Fatalf("example event type = %s, want %s", event.EventType(), definition.Type)
			}
		}
	}
}
