# CLAUDE.md

## Project overview

pgweasel is a fast CLI PostgreSQL log parser written in Rust (edition 2024, MSRV 1.92). It aggregates errors, locks, slow queries, connections, and system events from Postgres log files.

## Build & test commands

```bash
cargo build            # Build
cargo test             # Run all tests (unit + integration)
cargo test --test connections  # Run a single integration test file
cargo clippy           # Lint
cargo fmt --all -- --check # Check formatting
```

Each change should check tests and formatting, and if necessary fix.

## Commiting changes

Usually changes are done according to GitHub issues which are stated in subject:

`#12 Added sorting for connections analysis`

## Project structure

- `src/main.rs` — CLI entry point (clap-based subcommands: `err`, `locks`, `slow`, `conn`, `system`)
- `src/aggregators/` — One aggregator per feature (connections, error_frequency, error_histogram, top_slow_query)
- `src/filters/` — Log line filters
- `src/format/` — Log format detection and parsing
- `src/output_results/` — Output orchestration
- `tests/` — Integration tests using `assert_cmd` + `predicates`, one file per subcommand
- `tests/files/` — Sample log files for tests

## Conventions

- Integration tests run the compiled binary via `assert_cmd::cargo::cargo_bin!("pgweasel")` and assert on stdout/stderr
- Aggregators implement the `Aggregator` trait (`update`, `merge_box`, `print`, `boxed_clone`, `as_any`)
- Sorted output sections use `Vec::sort_by` on collected HashMap entries, sorted descending by count
- Use `memchr::memmem` for fast byte-level string searches in hot paths
- Parallel processing via `rayon`; aggregator results merged with `merge_box()`
