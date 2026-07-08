// Package matchstate tracks best-known match context from parsed CS2 log records.
//
// Context is updated incrementally after each successfully parsed event. Missing
// fields mean the value is not known yet, earlier records are not retroactively
// corrected, and errored records do not update tracked state.
//
// Tracker is designed to sit after stream.Processor. Push every stream record
// into the tracker before output filtering so filtered records still receive
// context from earlier map, round, score, pause, and round-end events.
package matchstate
