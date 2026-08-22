# Code refactoring TODO

## Perfomance

- [ ] Find big example

2026-06-14T14:11:15Z DEBUG pgweasel::output_results] Finished aggregating in: 475.911125ms

- Benchmark --debug option
- Isolate `samply record ./target/release/pgweasel slow top ../log_examples/pgbench_large.log` & `cargo run --release --features dhat-heap slow top ../log_examples/pgbench_large.log`
- Optimize Ty Claude

2026-06-14T14:50:42Z DEBUG pgweasel::output_results] Finished aggregating in: 261.124166ms

## Pedantic clippy TODO

Outstanding `cargo clippy --all-targets -- -W clippy::pedantic` findings.
Check each off as it is fixed; the file is clean once all boxes are ticked.

Re-run to refresh this list:

```bash
cargo clippy --all-targets -- -W clippy::pedantic
```

Totals: 43 warnings across 8 lints (27 in tests, 16 in `src/`).

**All clear** — `cargo clippy --all-targets -- -W clippy::pedantic` now reports
zero warnings.

## Source code

### `too_many_lines` (3) — split into smaller functions
- [x] `src/cli.rs:7` — extracted `errors_command()`
- [x] `src/main.rs:68` — extracted `handle_errors_command()`
- [x] `src/output_results/mod.rs:17` — extracted `compute_ranges()`, `process_with_context()`, `process_parallel()`

### Numeric casts — `src/aggregators/error_histogram.rs:67-69` (6)
Use checked/`try_into` conversions or document why the cast is safe (then `#[allow]`).
- [x] `cast_possible_truncation` / `cast_sign_loss` — `f64` → `usize` (line 67)
- [x] `cast_precision_loss` ×2 — `i64` → `f64` (line 67)
- [x] `cast_precision_loss` ×2 — `usize` → `f64` (lines 67, 69)

Resolved by storing bucket counts as `usize` and computing the bar width
with integer rounding instead of floats.

### Numeric casts — `src/duration.rs:47-48` (4)
- [x] `cast_possible_truncation` / `cast_sign_loss` — `f64` → `u64` (line 47)
- [x] `cast_possible_truncation` / `cast_sign_loss` — `f64` → `u64` (line 48)

Resolved by using `Duration::from_secs_f64` for the `ns`/`us` units, matching
the existing `ms`/`s`/`min` branches.

### Other
- [x] `needless_pass_by_value` — `src/filters/filter_contains_ci.rs:9` — `new(substring: String)` should take `&str`
- [x] `elidable_lifetime_names` — `src/output_results/mod.rs:279` — elide `'a` on `record_passes` (`&FilterContainer<'_>`)

## Tests

### `unnecessary_wraps` (27) — drop `-> Result<(), Box<dyn Error>>` from tests that never use `?`
- [x] `tests/grep.rs:6, 18, 30, 48`
- [x] `tests/locks.rs:6`
- [x] `tests/system.rs:7`
- [x] `tests/connections.rs:6`
- [x] `tests/help.rs:6, 18, 32, 46, 58`
- [x] `tests/slow.rs:7, 19, 31, 43`
- [x] `tests/errors.rs:9, 21, 33, 67, 86, 103, 115, 129, 154, 173, 195`

### `uninlined_format_args` (1)
- [x] `tests/errors.rs:53` — inline the variable into the `format!` string
