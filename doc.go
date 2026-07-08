// Package cs2log parses Counter-Strike 2 server log lines into typed events.
//
// CS2 log timestamps do not include a timezone. Parser configuration therefore
// requires the game server host timezone as an IANA timezone name, and parsed
// event timestamps are normalized to UTC.
//
// Parser handles one timestamped physical log line at a time. Lines without a
// supported CS2 timestamp return ErrNoMatch; timestamped lines with unknown
// payloads return Unknown. Use the stream package for whole streams and the
// matchstate package for derived match context.
package cs2log
