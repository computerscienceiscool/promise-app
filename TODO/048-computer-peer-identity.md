# TODO 048 - Generate the Computer Peer Identity and Copy Its CID

Create persistent peer identity and copy its displayed CID into the phone recipient field.

From `backend/`, use one explicit state directory for every command in the
demo:

```sh
PEER_STATE="$PWD/.promise-peer"
go run ./cmd/promise-peer identity \
  -identity-file "$PEER_STATE/identity.ed25519"
```

Record the CID, rerun the same command to confirm it is stable, then copy only
that CID to the phone. Keep the `0600` private identity file on the computer.
