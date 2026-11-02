// Package core is the domain of aurora-exporter: what a pipeline is,
// the ports (Source, Sink, Transformer) every connector implements,
// and the Runner that moves data between them.
//
// It imports nothing outside the standard library and knows nothing
// about Redshift, PostgreSQL, YAML, HTTP or schedulers. Every concrete
// technology lives in an adapter package that depends on core, never
// the other way around (Dependency Inversion): adding a new database
// means writing a new adapter, not editing this package (Open/Closed).
package core
