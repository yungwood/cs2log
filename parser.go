package cs2log

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	// ErrNoMatch means the line does not start with a supported CS2 timestamp.
	ErrNoMatch = errors.New("cs2 log line did not match")

	logLinePattern = regexp.MustCompile(`^(?:L )?(\d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}(?:\.\d{3})?)(?::| -) (.*)$`)
	timeLayouts    = []string{"01/02/2006 - 15:04:05.000", "01/02/2006 - 15:04:05"}
)

// Config controls parser behavior.
type Config struct {
	// LogTimezone is the IANA timezone name used to interpret zone-less CS2 log
	// timestamps. It defaults to UTC when empty.
	LogTimezone string
}

// Parser parses CS2 server log lines into typed events.
type Parser struct {
	location *time.Location
	matchers []eventMatcher
}

// NewParser creates a parser using the configured server log timezone.
func NewParser(config Config) (*Parser, error) {
	timezone := strings.TrimSpace(config.LogTimezone)
	if timezone == "" {
		timezone = "UTC"
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("load log timezone %q: %w", timezone, err)
	}

	matchers, err := compileDefinitions(defaultDefinitions)
	if err != nil {
		return nil, err
	}
	return &Parser{location: location, matchers: matchers}, nil
}

// Location returns the timezone used to interpret CS2 log timestamps.
func (p *Parser) Location() *time.Location {
	return p.location
}

// ParseLine parses one CS2 server log line.
func (p *Parser) ParseLine(line string) (Event, error) {
	result := logLinePattern.FindStringSubmatch(line)
	if result == nil {
		return nil, ErrNoMatch
	}

	parsedTime, err := parseTimestampInLocation(result[1], p.location)
	if err != nil {
		return nil, err
	}
	timeUTC := parsedTime.UTC()
	payload := result[2]

	for _, matcher := range p.matchers {
		event, ok, err := matcher.match(BaseEvent{Type: matcher.definition.Type, TimeUTC: timeUTC, Raw: line}, payload)
		if err != nil {
			return nil, err
		}
		if ok {
			return event, nil
		}
	}

	return Unknown{
		BaseEvent: BaseEvent{Type: "Unknown", TimeUTC: timeUTC, Raw: line},
		Payload:   payload,
	}, nil
}

func parseTimestampInLocation(value string, location *time.Location) (time.Time, error) {
	for _, layout := range timeLayouts {
		parsed, err := time.ParseInLocation(layout, value, location)
		if err == nil {
			return parsed, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse cs2 log timestamp %q", value)
}
