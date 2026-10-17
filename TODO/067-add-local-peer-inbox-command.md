# TODO 067 - Add Local Computer Peer Inbox Command

Provide a no-argument Make command that fetches and verifies the persistent
computer peer inbox from the locally running POC relay after a phone upload.

Completed 2026-10-17. `make -n peer-inbox-local` confirms the persistent
identity, local relay, and retained-inbox arguments; peer tests pass.
