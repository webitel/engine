# Email events for Workspace — EM-26

`engine` delivers the Workspace events of `email-service` to the operator's
Workspace over the existing WebSocket. An event tells Workspace that a Thread
changed; Workspace then reads it through the `EmailThreads` read API of
`email-service`. Events carry no email content and grant no access.

**Status.** This is the backend contract of stage 17. The WebSocket part
("WebSocket" and "Client synchronization" below) is a working contract waiting
for confirmation by Workspace; until then it may change, and only the mapping
in `engine` and this document change with it. The flag is off by default.

## Source

As for calls and chats, the sender addresses the event and RabbitMQ brings it
to the nodes: `email-service` writes one event per recipient, and the queue of
every node with a subscribed socket of that user is bound to its key. There is
no queue or dispatcher of `engine` in between.

| Exchange | Routing key | Contract |
|---|---|---|
| `email.events` (`topic`, durable) | `email.workspace.<type>.v1.<domain_id>.<user_id>`, `<type>` one of `thread.changed`, `message.created`, `message.delivery` | Workspace events, `webitel-emails/docs/events.md` |

- `email-service` declares the exchange and publishes with confirms and
  without `mandatory`: a user without an open socket is the usual case and
  blocks nothing. `engine` declares the same exchange with the same parameters,
  so a binding never waits for the producer.
- A subscription binds the queue of its node with
  `email.workspace.*.*.v1.<domain_id>.<user_id>`; `email.message.inbound.v1`,
  the Flow event on the same exchange, never matches.
- A node takes an event only when it is `application/json` within the size
  limit, a known type and schema, its type, domain and user equal those of the
  routing key, and the domain and user equal those of its hub and of the
  socket session; any other is dropped and counted, as the queues of the nodes
  acknowledge automatically.
- `call_center` sends `engine` nothing for Email: its snapshots reach
  `email-service` alone, which turns them into these events once its read model
  changed.
- Confirmed means accepted by the broker, also when no socket is bound. A lost
  queue of a node is seen and answered with `subscription_lost` ("Delivery"); a
  binding deleted by hand while the socket stays open is not seen until the
  next resync, as for calls and chats.

## WebSocket

With `email_workspace` (`EMAIL_WORKSPACE`) a socket subscribes to the Email
events of its own session. The domain and the user come from the session; a
client never names a user, an agent, a Thread or a routing key.

### Subscribe

```json
{"seq": 7, "action": "subscribe_email", "data": {}}
```

```json
{"status": "OK", "seq_reply": 7, "data": {"schema_version": 1}}
```

The session needs `email_thread` read; update alone is not enough, and no agent
or online status is needed. Without `email_workspace` the reply is an error
with status 400, without the right 403. The reply comes after the broker
binding of the node is confirmed, so no event of the session after the reply is
lost by the subscription itself; a binding not confirmed in 5 seconds is an
error, and the client subscribes again. Events may arrive before the reply. A
repeated subscribe answers `OK`; `unsubscribe_email` removes only this socket. A
subscription ends with the socket: after a reconnect the client subscribes
again.

### Delivery

- Every delivery checks that the session is valid, of the same user and domain,
  and has `email_thread` read. The check is repeated at most once per
  `auth_cache_expire` (`AUTH_CACHE_EXPIRE`, 30 seconds by default) and when the
  session expires, so a revoked right or session stops the events within that
  time, not at once. The check runs beside the hub: a slow auth delays only the
  Email events of that socket.
- A session that ended or lost the right gets `resync_required` with
  `subscription_lost`, and its subscription ends; the next `subscribe_email`
  answers with the error. A token of another user closes the socket.
- While auth is unavailable no event goes. The check is repeated by a timer,
  after 5 seconds and then twice as long up to 60, also when no other event
  comes; once auth answers, the socket gets `subscription_lost`.
- Each socket queues up to 64 events; a fuller queue loses events and sends
  `overflow`. A socket whose send buffer is full is closed, as for the other
  events.
- When the queue of the node comes back after a broker failure, it has no Email
  bindings: every subscription of the node ends and its socket gets
  `subscription_lost`.
- After `subscription_lost` the client sends `subscribe_email` again (it answers
  `OK` when the subscription is still there) and then reads everything in view;
  after `overflow` it reads everything in view.
- A socket gets an `event_id` once among its last 256 events; other sockets of
  the user get their own copy.
- Every subscription binds the node queue under its own id, so a subscribe that
  timed out and finished late never removes the next one of the socket. The
  bindings change on a channel of their own, one after another: a slow broker
  delays only them, not the other events of the node or its reconnect. A
  removal that failed is repeated after 1 second, then twice as long up to 30.

### Event

Every event is `event: "email"`; `data.type` tells its kind:

```json
{"event": "email", "seq": 42, "data": {
  "type": "thread_changed",
  "schema_version": 1,
  "event_id": "5b247aa1-9433-53ce-b37d-346cec14bff5",
  "thread_id": "1024",
  "access": "current",
  "workspace_version": "17",
  "changes": ["projection", "actions"]
}}
```

| `type` | Source | Fields besides `type`, `schema_version` |
|---|---|---|
| `thread_changed` | `email-service` | `event_id`, `thread_id`, `access`, `workspace_version`, `changes` |
| `message_created` | `email-service` | `event_id`, `thread_id`, `access`, `workspace_version`, `message_id` |
| `message_delivery` | `email-service` | As `message_created`; reserved for stage 17.7 |
| `resync_required` | `engine` | `reason`: `subscription_lost` or `overflow` |

- `access` is `current`, `history` or `none`; `none` keeps the version, has no
  `message_id` and an empty `changes`.
- Ids and versions are decimal **strings**, as `int64` in the read API
  (protojson). Compare them as integers (`BigInt`, or length, then text), never
  through `Number`.
- The Email commands (EM-25) keep JSON numbers. Before sending a command the
  client checks each string against `^[0-9]+$` and, as `BigInt`, the range
  1…9007199254740991 for `thread_id` and `attempt_id` and
  0…9007199254740991 for `expected_version`; outside it the command is not
  sent, since a rounded number would name another value. `engine` itself still
  accepts up to 2^53, one above that range; narrowing it is left for a change
  of the command contract.
- `seq` numbers the messages of one socket; it is not a cursor and does not
  survive a reconnect. Duplicates are possible, so a client removes repeats by
  `event_id`. Unknown `type` values and fields are ignored.
- No subject, body, address, attachment, reason or comment is ever sent.

## Client synchronization

The read API is the state; events only say when to read it, as for calls and
chats: subscribe, load the state, let events start reads, reload everything
after a reconnect.

The change of a Thread, its `workspace_version` and the outbox row of its event
are one transaction, the event is published after the commit, and the read API
reads the same database, without replicas or a cache. A read started after an
event with version W therefore returns version W or later, or `NotFound` when a
later change hid the Thread, and that change has its own event.

1. **Events.** The client acts on `thread_changed`, `message_created`,
   `message_delivery` and `resync_required`; it ignores unknown types.
2. **Subscription and sync cycle.**
   - Send `subscribe_email` until it answers `OK`, with a growing pause; no live
     sync starts without it.
   - After `OK`, load CURRENT, the HISTORY in view and the open Thread; events
     that came meanwhile are handled after this load.
   - Every full load is a new local sync cycle; answers of an earlier cycle are
     dropped and start no retry.
   - `Unauthenticated` or `PermissionDenied`, of the subscription or of a read,
     stops the sync: the client asks for a new sign-in, without a reconnect
     loop.
   - A reconnect, `resync_required` or a failed subscription start a new cycle;
     after `subscription_lost` the client subscribes again first.
3. **State of a Thread:** the applied version, which never goes down; the
   highest event version; "read needed"; "read running"; a removal mark with a
   floor; "Thread read" and "screen in sync" as two flags; unfinished loads of
   messages and actions.
   - An event with a version not above the applied one is ignored; otherwise
     the Thread needs a read.
   - One read of a Thread runs at a time; an event during it keeps "read
     needed", and the read repeats after the answer.
4. **Answers**, of a Thread read and of a list page:
   - Every read keeps a floor taken at its start: the higher of the applied
     version and the highest event version.
   - An answer above the applied version is applied. An answer with the same
     version is taken too: it refreshes the cache or ends an unfinished load. A
     lower one is ignored, and so is one not above the floor of a removal mark.
   - A Thread missing from a list page is not removed. On a full load, a Thread
     missing from every loaded page is read on its own.
5. **`NotFound` of a Thread read.**
   - It counts only when no later state was applied after the read started,
     from a list page or another answer; otherwise it is dropped and the Thread
     needs a read.
   - A counted `NotFound` removes the card and leaves a removal mark with the
     floor of that read; a later event reads the Thread again.
   - `NotFound` of a message or an attachment does not remove the Thread: the
     client reads the Thread and decides by that answer.
6. **Messages and actions.** A Thread version does not prove them current.
   - The open Thread loads its messages, from the start of the list since a
     later ready message may have an earlier receive time, and its actions
     after every applied answer, also one of the same version.
   - An unfinished load stays marked until it succeeds. Its failure keeps
     "Thread read" but not "screen in sync", and is retried by the error rules.
   - For one consistent screen the Thread version is read before and after
     the messages and actions; when it changed the load repeats a few times,
     then the Thread stays marked.
   - A cache of closed Threads, if any, turns stale on any event of the Thread.
7. **Read errors**, by gRPC code (the HTTP mapping of the Gateway for a browser
   is checked on the stand with F2):
   - `Unavailable`, `DeadlineExceeded`, `ResourceExhausted`, `Internal`: retry
     after 0.2, 0.4, 0.8, 1.6 and 3.2 seconds, then show "syncing" and retry
     every 30 seconds;
   - `NotFound`: rule 5;
   - `Unauthenticated`, `PermissionDenied`: rule 2;
   - `InvalidArgument`: a client error, logged, never retried.
8. **Commands (EM-25)** are allowed on CURRENT of the last applied answer.
   - No error repeats a command by itself: a version conflict, a wrong owner or
     state, `NotFound`, `PermissionDenied` or a temporary failure. The client
     shows the error and reads the Thread.
   - After an unknown result the client keeps the whole request: `command_id`,
     `attempt_id`, `expected_version`, the Transfer target, the reason, the
     comment and the other fields. Only the user repeats it, with the same
     request; a version from a later read is never put in.
   - After a command was refused for good, a new action of the user gets a new
     `command_id`: other fields under the old id answer
     `command_id_conflict`.

Between a change in `call_center` and the `email-service` event passes the time
the projection needs; when the projection fails, it may be long. All that time
the client shows the last confirmed state, and the checks of EM-25 (owner,
state, version) keep the commands correct. Events go to open sockets only:
there is no offline queue and no client acknowledgement.

## Configuration

| Flag | Env | Default | Effect |
|---|---|---|---|
| `email_workspace` | `EMAIL_WORKSPACE` | `false` | Accept `subscribe_email`, bind the queues of the nodes and deliver events |

The integration tests need RabbitMQ 4:
`EMAIL_TEST_AMQP_URL=amqp://… go test -tags=integration ./mq/rabbit/`.

## Rollout

Deploy `engine` first; turn on `WORKSPACE_EVENTS_ENABLED` in `email-service`
after it, and `EMAIL_WORKSPACE` last, for a client that follows "Client
synchronization". The format of the routing key is fixed by `email-service`
and `engine` together: a release that changes it changes both. The full order, checks
and rollback are in stage 17.8 of `Emails_Tasks.md` (not shipped in this
repository).
