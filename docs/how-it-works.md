# How vdb-mysql Works

vdb-mysql is built in three layers that stack on top of each other. Understanding them will help you reason about what it can and cannot do, and where plugins fit in.

---

## The Three Layers

### Layer 1 — MySQL Wire Protocol

The outermost layer is powered by [Vitess `go/mysql`](https://github.com/dolthub/vitess). It handles raw TCP connections, the MySQL handshake, and authentication. From the perspective of any MySQL client or driver, vdb-mysql looks identical to a standard MySQL 8 server.

vdb-mysql holds no user list of its own. On every new connection it opens a short-lived probe connection to the source MySQL server to verify the client's credentials, fetches the user's grants via `SHOW GRANTS FOR CURRENT_USER()`, and then closes the probe immediately. See [Authentication](#authentication) below.

### Layer 2 — SQL Engine

SQL parsing, planning, and execution are handled by [go-mysql-server](https://github.com/dolthub/go-mysql-server), an embeddable MySQL-compatible SQL engine. It receives each statement after the wire layer decodes it.

For reads, go-mysql-server calls the VDB storage backend to fetch rows. Those rows come from the source database via a persistent, read-only connection pool — the same pool that serves all clients. The schema is read from `INFORMATION_SCHEMA` and cached.

For writes (`INSERT`, `UPDATE`, `DELETE`), go-mysql-server calls the VDB storage backend instead of issuing anything to the source. Those writes go into the **delta store** (see below) and never reach the source database.

### Layer 3 — VDB Core

The innermost layer is the pipeline-and-event bus provided by `github.com/virtual-db/core`. Every significant moment in a connection's lifecycle — a query arriving, rows being fetched, a row being inserted — fires one or more named pipeline points. Plugins attach handlers to those points to intercept, transform, or observe data.

---

## The Delta Store

The delta is the central concept that makes vdb-mysql different from a transparent proxy.

When a client issues a write:

1. go-mysql-server calls the VDB storage backend.
2. The change is recorded in the delta as an **insert**, **update overlay**, or **tombstone** — entirely in memory.
3. On the next read, source rows are fetched from the source, the delta is overlaid on top of them, and the merged result set is sent through the plugin pipeline before being returned to the client.

### What the delta tracks

| Operation | What is stored |
|---|---|
| `INSERT` | The full new row, keyed by its primary key. |
| `UPDATE` | An overlay of the changed column values, keyed to the original source row's primary key. |
| `DELETE` | A tombstone keyed to the source row's primary key. Deleted rows are excluded from merged results. |

### Delta lifetime

The delta is **process-scoped and in-memory**. It does not persist across restarts, is not replicated, and cannot be snapshotted in the current release. All captured writes are lost when the process exits.

### Chained writes

vdb-mysql correctly handles chained mutations across both explicit and implicit (autocommit) transaction boundaries. An `UPDATE` followed by another `UPDATE` on the same row, or an `UPDATE` followed by a `DELETE`, resolves the stable source key through a fallback mechanism so the delta always reflects the correct final state.

### Transaction isolation

Each transaction gets its own private delta (`TxDelta`). On commit, the transaction delta is merged into the live (process-wide) delta. On rollback, the transaction delta is discarded.

---

## Authentication

vdb-mysql never stores passwords or maintains a user list. The authentication flow for every new connection is:

1. The client connects and begins the MySQL handshake.
2. vdb-mysql opens a short-lived TCP connection to `VDB_AUTH_SOURCE_ADDR` and replays the handshake with the credentials the client provided.
3. If the source accepts the credentials, vdb-mysql runs `SHOW GRANTS FOR CURRENT_USER()` on the probe connection to capture the user's grants.
4. The probe connection is closed immediately.
5. The client's session is created, carrying the grants for its lifetime.

**TLS behaviour:**

- When `VDB_TLS_CERT_FILE` is set, vdb-mysql advertises `caching_sha2_password`. TLS is required for the full-auth path, which ensures the plaintext password is delivered securely to the probe.
- When TLS is not configured, `mysql_clear_password` is advertised instead. This is suitable for local development or environments where transport security is handled at the infrastructure level.

The grants fetched at connection time are carried on the session object and are available to the storage backend and plugins throughout the connection's lifetime.

---

## Query Lifecycle (Step by Step)

Here is the full path of a `SELECT` statement from the moment it arrives to the moment results leave:

```
Client
  │
  ▼
[Layer 1] Vitess go/mysql
  Decodes the MySQL wire packet, extracts the SQL string.
  │
  ▼
[VDB Core] vdb.query.received pipeline
  Plugins at the `intercept` point may rewrite the query.
  │
  ▼
[Layer 2] go-mysql-server
  Parses, plans, and begins execution.
  Calls the VDB storage backend to fetch rows.
  │
  ▼
[VDB Core] vdb.records.source pipeline
  Source rows are passed through.
  Plugins at the `transform` point may add, remove, or modify rows.
  │
  [Delta overlay applied]
  Tombstoned rows are removed. Update overlays are applied. Inserted rows are appended.
  │
  ▼
[VDB Core] vdb.records.merged pipeline
  The final merged row set is passed through.
  Plugins at the `transform` point may make last-mile adjustments.
  │
  ▼
[Layer 1] Vitess go/mysql
  Encodes and sends the result set back to the client.
```

And for a `INSERT` / `UPDATE` / `DELETE`:

```
Client
  │
  ▼
[Layer 1] Vitess go/mysql → [Layer 2] go-mysql-server
  Parses and plans the write. Calls the VDB storage backend per affected row.
  │
  ▼
[VDB Core] vdb.write.insert / vdb.write.update / vdb.write.delete pipeline
  The `apply` point records the mutation in the delta.
  Plugins at this point may transform or reject the write.
  │
  [Source database is NOT contacted for writes.]
  │
  ▼
Client receives an OK packet.
```

---

## Related Documentation

- [Configuration Reference](./configuration.md)
- [Plugins](./plugins.md)
- [Pipelines and Events Reference](./pipelines-and-events.md)
- [Known Gaps and Limitations](./limitations.md)