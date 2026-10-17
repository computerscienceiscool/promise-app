# TODO 047 - Build the Computer Peer CLI

Build and validate the `promise-peer` desktop recipient CLI.

Validation commands, run from `backend/`:

```sh
go build ./cmd/promise-peer
go test ./cmd/promise-peer
```

Or run the combined host build from the project root:

```sh
./scripts/build-phone-computer-host.sh --offline
```

The build is complete only when both commands pass. The CLI is a local POC
implementation; its signed-message verification follows the frozen POC
protocol profile, while the desktop runtime and local files are implementation
details.

Host-side validation completed on 2026-10-17 using a writable temporary Go
cache: `go test ./...` and `go build ./cmd/promise-peer` both passed. The
remaining peer TODOs require a real LAN relay and a phone-created message; see
`docs/todo-audit-2026-10-17.md`.
