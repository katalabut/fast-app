# Local release candidate — 2026-10-13

This is an uncommitted candidate based on v0.2.1, not a published release or a
production deployment. A version, release timing and client rollouts require a
separate decision. Existing APIs, config struct layouts/tags/defaults, listener
ports and degraded-readiness policies are preserved.

## Configuration contract and migration

Built-in precedence is environment > file > defaults. Defaults are applied
before overlay, so supplied `false`, `0`, `""`, empty slices and maps survive.
Environment fields absent from YAML now load, including nested/optional fields
and `mapstructure` names/squash. Decoder-created optional pointers and slice/map
children are initialized within their field's defaults policy before their native
source overlay. Typed struct assignment, weak singleton/container conversions,
field names/tags/case and null behavior remain the decoder's responsibility.
Decoded zero/false values are never passed through defaults again. A field tagged
`default:"-"` suppresses its whole subtree, including newly allocated children and
custom SetDefaults calls; ordinary siblings of the same type still receive defaults.
A typed struct of the target type is a complete supplied value, including its zeros.
Scoped results replay without pointer allocation wrappers, including collection
children; the native decoder preserves target pointer depth, nilness and existing
objects. Pointer-to-interface replay preserves the complete dynamic value.
Null follows native decoding: ignored null field values retain existing defaults;
null collection elements do not acquire invented children. Weak slice-to-map input
merges its elements while replacing the original default container once. `yaml` tags do not rename Viper/mapstructure keys.
Both `WithEnv("APP")` and `WithEnv("APP_")` use `APP_PORT` for a `Port` field;
a nested `Database.Host` uses `APP_DATABASE_HOST`. Empty environment values are
explicit values. Parser instances are not concurrent-safe; independent parsers
can run concurrently. Custom sources execute in registration order; replacement
moves the named source last. A source's own Viper setter still determines its
priority. `Parse` requires a non-nil struct pointer.

`WithFile(paths...)` retains first-existing-file selection. A non-empty
`CONFIG_FILE` used with `WithFileFromEnv` is authoritative: an absent path now
returns a clear error instead of silently selecting a fallback. An unset/empty
variable retains the fallback list. File readability/format/decode errors still
return errors. No validation/hot reload of arbitrary business config is implied.
Optional `config.App.Validate()` checks library endpoints; call business
validation explicitly in the composition root. For explicit profiling opt-in,
`Observability.ValidateWithPprof(true)` includes its actual routes. The service
uses its own opt-in state; dormant default-prefix pprof patterns are not reserved. Enabled servers reject malformed
or conflicting HTTP patterns rather than panicking during registration.

| Change | Effect on Links/Habbit | Preserve previous effective behavior |
| --- | --- | --- |
| Explicit false/zero/empty respected | Links `Metrics.Enabled:false` and `Debug.Enabled:false` now really disable them. Habbit's post-load switch assignment remains compatible. | Set the intended prior effective values explicitly (for example metrics true). Do not restore a broken zero-value overlay. |
| Env-only/absent YAML keys load | Previously ignored deployed values may now affect both apps. Habbit uses explicit `mapstructure` keys; Links largely uses field names. | Audit/remove unwanted env keys or provide intended values explicitly. Do not assume generic `yaml` tags change env names. |
| Prefix trailing `_` normalized | Neither inspected client calls WithEnv with a prefix. | Replace unintended legacy doubled separators with the documented single separator. |
| Strict non-empty CONFIG_FILE | Habbit calls WithFileFromEnv; Links calls WithFile and is unaffected by this variable. | Fix/unset CONFIG_FILE if fallback was intentionally wanted. |
| TTL/invalidation fixed | Health status refreshes after actual cache age; replacement/clear cannot restore stale entries. | Increase configured TTL deliberately if slower refresh was intended. Infinite stale caching is not preserved. |
| Stop-all option works | Neither inspected startup uses WithDisableStopAllOnErr. Default fail-fast remains. | Omit the option for original fail-fast. With it, the failed service cleans up independently and final exit remains 1. |
| Shutdown panic isolated | Other services can complete cleanup; exit becomes 1 rather than an unhandled panic exit 2. | Fix the service's panic; callers should treat any nonzero exit as failure. |
| HTTP checker body limit | Health response matching reads at most 64 KiB; raw actual/expected body details are removed. No inspected client reads these detail keys. | Use a purpose-built checker for larger payloads or carefully approved diagnostics. |

The read-only inspection found both backend modules pinned to fast-app v0.2.1.
Links uses `fastapp.New(cfg.App).Add(...).Start()` and file loading. Habbit also
passes WithContext/WithShutdownTimeout and already drains its main HTTP server
before closing its PostgreSQL pool. No client module, config, secret or deployment
was changed. This is source inspection, not an integration-test or rollout claim.

## Exposure settings: explicit opt-in, unchanged defaults

The historical all-interface `:port` bind, Config.Debug.Enabled default, HTTP
Details output and ports remain unchanged. Root/config structs gained no fields.
In particular, enabling new default-prefix pprof routes is explicit, preventing
the old default 404 bug from becoming an automatic new profiling exposure:

```go
app := fastapp.New(cfg.App,
    fastapp.WithObservabilityAddress("127.0.0.1:9090"),
    fastapp.WithHealthDetailsRedacted(),
    // fastapp.WithPprof(), // explicit diagnostics opt-in; requires Debug.Enabled
)
```

Direct observability services provide WithListenAddress, WithDetailsRedacted and
WithPprof builders. The legacy debug service has WithListenAddress/WithPprof;
the standalone health server has WithListenAddress/WithDetailsRedacted. Configure
builders before Run/Start and do not mutate them concurrently with serving.
An intentionally exposed listener can use an explicit `0.0.0.0:port`. Never
publish pprof or sensitive checks through a public ingress without an explicit
access policy. Redaction removes Details only: custom Message/Status still must
not contain secrets. Existing custom-prefix pprof remains available through
private handler registration; arbitrary DefaultServeMux routes are not proxied.
WithPprof uses a 35-second write budget for the standard 30-second CPU profile;
longer profiles need a different explicitly managed server. Default write budgets
remain unchanged when the option is absent.

Changing bind/debug/details defaults globally remains a separate migration
and version decision; it is not included in this candidate.

## Lifecycle and resource ownership

The accepted SIGTERM/SIGINT pilot remains: Start owns os.Exit, callers' defers do
not run, shutdown has a fresh context, and watchdog delay is saturating timeout
plus 5 seconds. Its independent final warning/flush attempt gets at most another
100 ms. A blocked logger cannot disable forced exit. Normal/error Sync still
runs before exit and is bounded by that watchdog, not a durability guarantee.

WithDisableStopAllOnErr lets siblings run until cancellation after a service
failure, logs failures immediately, cleans the failed service and eventually
exits 1. There is no automatic restart. Run and Shutdown panics are recovered as
errors at lifecycle boundaries. A spontaneous nil Run does not promise that the
application stops; Start's termination contract remains signal/error driven.

Shutdown callbacks are concurrent, and startup readiness remains explicitly
controlled by the application. Do not scatter shared-resource destructors across
independent services and assume registration order means shutdown order. One
resource-owning service should drain its HTTP/workers before closing its DB/pool.
On a drain deadline, force-close HTTP and account for cancelled work before
releasing dependencies. Checkers must cooperate with context cancellation;
FastApp cannot forcibly terminate arbitrary Go callbacks. Checker-owned nested
Details values should be immutable; cache results copy the top-level map.

HTTP service/server objects are single-use: construct another instance to restart.
Their http.Server is allocated before Run/Start, so Shutdown-before-start cannot
be undone by later server initialization. An already cancelled context prevents
startup. Call Shutdown concurrently to stop and drain a running HTTP server.

## Verification and remaining gates

Root go.mod/go.sum and previously supplied AGENTS/research/validation documents
are preserved. All eight examples are built in their own module; incomplete
nested manifests are repaired without upgrading root dependencies. Basic/simple/
advanced readiness flags are atomic, and logger example uses ContextWithKV.
Examples remain demonstrations, not production templates or DB integration tests.

CI checks the declared minimum 1.23.6 and locally tested 1.26.1 with automatic
compiler switching disabled, uses a pinned golangci-lint 2.11.4, and builds all
examples. The v9 lint action supports the v2 linter (official action docs:
https://github.com/golangci/golangci-lint-action). CI is prepared, not executed.
Local verification results/hashes are recorded outside the repository; see the
final task report before treating the candidate as a tested artifact.

Remaining decisions/gates: choose a release version; independent review of the
combined diff; run Linux and minimum-toolchain CI; decide future exposure defaults;
audit deployed env/effective config before pin upgrades; execute isolated client
compatibility tests; prepare the generator and controlled client rollout separately.
No release tag, push, merge, deploy, production DB or external API run is implied.
