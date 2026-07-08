// Package stream converts CS2 log streams into parsed records.
//
// It builds on cs2log.Parser by tracking line numbers and grouping multiline
// structures such as JSON blocks and server cvar dumps. Records preserve the
// source line range and may contain root-package events or stream-specific
// events such as RoundStats.
//
// Processor is incremental. PushLine may return no records while a multiline
// block is open; call Flush at EOF or shutdown to emit any incomplete block.
// Line numbers are per Processor instance and input stream.
package stream
