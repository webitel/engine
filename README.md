# engine

## Email Agent API — EM-21

Substeps 5.1–5.4 are prepared on `feature/EM-21-email-distribution`.
On 2026-10-23 build, focused Email tests, real PostgreSQL 18 store/handler smoke
and eight concurrency cases passed. Overall vet/Queue tests are not green:
three vet findings and three progressive-queue test failures also reproduce on
unchanged HEAD. Acceptance of these baseline exceptions remains pending.
Local 7.4 also passed the existing ExtraChat zero/nonzero test and five actual
Agent handler/store groups with distinct channel limits: Email-only Patch,
non-Email Patch, omitted Email limits in Update and masked-field isolation.
No new regression was found in that scope, not full other-channel E2E.
Keep production Email polling/distribution off until the release gates pass.
Agent permissions/domain/RBAC code and other-channel update rules are unchanged.

### Limits

`email_count` (base) and `extra_email_count` (Extra) accept 0–32767.
Zero disables only that slot type. Negative JSON is rejected by decoding;
overflow is rejected before writes by `model.Agent.valid.EmailCount` or
`model.Agent.valid.ExtraEmailCount`. Existing transport error mapping is reused.

| Operation | Base | Extra |
|---|---|---|
| Create | Omitted: 1; explicit 0: 0 | Omitted: 0 |
| Update | Omitted: preserve; explicit 0: 0 | Omitted: preserve; explicit 0: 0 |
| Patch | Only when `email_count` is in `fields` | Only when `extra_email_count` is in `fields` |

A masked Patch field without a scalar value sets 0; unmasked values are ignored.
Internal writes use named `model.AgentEmailLimits{Base: ..., Extra: ...}`.
SQL retains `coalesce(:EmailCount::smallint, email_count)` and the Extra equivalent:
omitted limits use current columns, not a stale read. Explicit writes remain
last-writer-wins; this does not protect all Agent fields from concurrent edits.
Actual NULL binding and eight controlled races passed on isolated PostgreSQL 18.6.
Limits are read/default search fields; lowering them does not release attempts.
An agent Extra limit does not grant a letter permission to use Extra.

### Contracts, compatibility and rollout

Generate only selected local EM-21 inputs from the engine checkout:

```sh
buf generate ../protos/engine --template buf/buf.gen.yaml --path ../protos/engine/agent.proto
buf generate ../protos/cc --template buf/buf.gen.cc.yaml --path ../protos/cc/cc_member.proto
```

Do not edit generated code/templates or use `--clean` / `--include-imports`.
Extra tags: Create/Patch 18, Update 19, Agent 28; preceding gaps belong to later
postponed-limit contracts. Base tags are unchanged; Create/Update use presence.
Legacy scalar clients usually omit zero: Create then uses base 1, Update keeps
current limits. Unsent explicit-zero intent cannot be recovered; use updated
optional-aware clients or a Patch mask. Wire tests cover decoding, not old-client E2E.

**Admin UI/Gateway/Swagger handoff — pending, not sent:** update compatible
descriptors/schema and client types, preserve absent versus explicit 0 in requests,
and treat omitted scalar limits in responses as 0 (JSON emission depends on codec
default options). Engine Go generation does not update these external components.
Deploy the call_center routing schema/view first (`extra_email_count`,
`cc_agent_list`); no extra engine migration or rewrite of existing limits is added.
When adapting EM-23, extend the named limits structure with `Postponed`, not a
positional pointer argument; check all three fields independently.

Queue Sticky reuses `sticky_agent` and payload keys without new validation.
`payload.*` Patch merges all supplied map keys; full Update replaces payload.
Type 11 retains calendar/non-nil payload requirements. Runtime rules and safe
parsing are documented in [call_center](../call_center/README.md). Queue code is unchanged.

### Verification scope — executed 2026-10-23

1. Start with `go build ./...`; run vet for changed packages and focused tests
   in model/app/grpc_api, including existing Queue validation. No new framework.
2. Isolated PostgreSQL 18/current schema: actual store Create/Get/Search with
   and without RBAC, Update/Patch, explicit zero and NULL parameter binding.
3. Real base/Extra edit versus an omitted-limit Update/Patch in both transaction
   orders. Use the actual store; a unit fake is not PostgreSQL concurrency proof.
4. Authorized API smoke: rights/domain/RBAC, Extra inside/outside Patch mask,
   unchanged other-channel limits and Sticky payload write/read.

Handler smoke used synthetic sessions and a local audit sink, not external auth,
the production interceptor, HTTP Gateway or RabbitMQ audit delivery. No full
server startup/E2E claim is made. Baseline findings were not fixed or suppressed.
Workspace reports: `tmp/Emails_Routing_Step5_Verification.md` and
`tmp/Emails_Routing_Step7_4_Verification.md` (not shipped in this repository).
Record snapshot, environment, results and unexecuted checks. Do not use production
data. Targeted High-10 clean/upgrade verification in call_center passed in 7.2.1.
Full other-channel E2E and explicit baseline disposition remain open; the local
step-7 summary is not merge/rollout approval. Lifecycle/Transfer/events remain
later work; see the workspace routing update plan for remaining gates.
