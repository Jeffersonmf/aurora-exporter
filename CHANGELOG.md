# Changelog

All notable changes to this project are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed
- Rewritten as a generic, declarative exporter: pipelines are YAML
  (source query, target table, write mode, schedule) instead of one Go
  file per table. `examples/quickstart` is a runnable example with an
  initial migration.
- Module path is now `github.com/Jeffersonmf/aurora-exporter`; Go 1.27.
- HTTP API replaced: `POST /v1/pipelines/{name}/runs` instead of one
  route per table; added run history, `/healthz`, `/readyz`, Bearer auth.

### Added
- Targets: PostgreSQL, Aurora PostgreSQL, MySQL, Aurora MySQL, JSON lines.
- Write modes: `append`, `replace`, `truncate`, `upsert`.
- Transforms: `sanitize_utf8`, `rename_columns`, `drop_columns`, `lowercase_columns`.
- CLI: `serve`, `run`, `validate`, `drivers`, `version`.
- Makefile, golangci-lint, GitHub Actions CI, integration suite, Docker image.

### Fixed
- NULL values were written as zero values (`0`, `false`, `0001-01-01`).
- DELETE + INSERT ran outside a transaction; readers saw empty tables
  and failures left them empty.
- `POST /vm_brands/` ran the brand wallet history exporter.
- Concurrent triggers of the same exporter interleaved their loads.

### Security
- Removed hard-coded database credentials from source code and the
  committed `.env`. **Rotate the old credentials**: they remain in git
  history.
