# Contributing to mysql

> [!IMPORTANT]
> ### Public Contributions Are Not Yet Open
>
> VirtualDB is not currently accepting public pull requests. We intend to open the project to community contributions in the future — when we do, this document will be updated accordingly.
>
> **In the meantime, we encourage you to participate by [opening a GitHub Issue](https://github.com/virtual-db/mysql/issues).** Issues are the best way to report bugs, request features, ask questions, and start discussions with the team.
>
> Thank you for your interest in VirtualDB.

---

## What This Repository Is

`mysql` is the runnable server binary. It wires together `core` and `mysql-driver` into a deployable process. The logic for how data is intercepted and transformed lives in those two downstream modules, not here.

Contributions to this repository are narrow in scope:

- Changes to the `main.go` entry point (environment variable handling, startup wiring).
- Dockerfile or CI workflow improvements.
- Documentation corrections.

If you want to contribute to query interception, row transformation, authentication, schema handling, the plugin protocol, or the delta layer, the correct repositories are:

| Area | Repository |
|---|---|
| MySQL protocol, schema, rows, auth proxy | [mysql-driver](https://github.com/virtual-db/mysql-driver) |
| Delta layer, plugin protocol, framework pipelines | [core](https://github.com/virtual-db/core) |
| Integration test suite | [tests](https://github.com/virtual-db/tests) |

---

## Reporting Bugs

Open an issue. Include:

- Go version and OS.
- The MySQL source version you are connecting to.
- The `vdb-mysql` binary version (from the release tag or `git describe`).
- Relevant environment variable values (with credentials redacted).
- Observed vs. expected behaviour, with any relevant log output.

---

## Security Issues

Do not open a public issue for security vulnerabilities. Email the maintainers directly. You will receive a response within 5 business days.

---

## License

Contributions will be licensed under the Elastic License 2.0. See [LICENSE.md](LICENSE.md).