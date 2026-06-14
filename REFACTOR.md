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

## Source code

### `too_many_lines` (3) — split into smaller functions
- [ ] `src/cli.rs:7` — 113/100 lines
- [ ] `src/main.rs:68` — `main()` 105/100 lines
- [ ] `src/output_results/mod.rs:17` — `output_results()` 130/100 lines

### Numeric casts — `src/aggregators/error_histogram.rs:67-69` (6)
Use checked/`try_into` conversions or document why the cast is safe (then `#[allow]`).
- [ ] `cast_possible_truncation` / `cast_sign_loss` — `f64` → `usize` (line 67)
- [ ] `cast_precision_loss` ×2 — `i64` → `f64` (line 67)
- [ ] `cast_precision_loss` ×2 — `usize` → `f64` (lines 67, 69)

### Numeric casts — `src/duration.rs:47-48` (4)
- [ ] `cast_possible_truncation` / `cast_sign_loss` — `f64` → `u64` (line 47)
- [ ] `cast_possible_truncation` / `cast_sign_loss` — `f64` → `u64` (line 48)

### Other
- [ ] `needless_pass_by_value` — `src/filters/filter_contains_ci.rs:9` — `new(substring: String)` should take `&str`
- [ ] `elidable_lifetime_names` — `src/output_results/mod.rs:279` — elide `'a` on `record_passes` (`&FilterContainer<'_>`)

## Tests

### `unnecessary_wraps` (27) — drop `-> Result<(), Box<dyn Error>>` from tests that never use `?`
- [ ] `tests/grep.rs:6, 18, 30, 48`
- [ ] `tests/locks.rs:6`
- [ ] `tests/system.rs:7`
- [ ] `tests/connections.rs:6`
- [ ] `tests/help.rs:6, 18, 32, 46, 58`
- [ ] `tests/slow.rs:7, 19, 31, 43`
- [ ] `tests/errors.rs:9, 21, 33, 67, 86, 103, 115, 129, 154, 173, 195`

### `uninlined_format_args` (1)
- [ ] `tests/errors.rs:53` — inline the variable into the `format!` string
