# TODO 066 - Fix Relay Setting Persistence and Verification

The phone repeatedly retained the old LAN relay URL after attempts to update
it. Add an explicit **Use USB relay** action that saves the known USB address
without clipboard or typing, and show the currently persisted relay URL on the
screen so a person can verify the setting before sending.

Completed and physically evidenced in the phone screenshot:

- **Use USB relay** persists `http://127.0.0.1:8080` directly.
- **Active relay:** displays the persisted value immediately after saving and
  on later launches.
- `docs/example-poc.md` retains the screenshot that shows the active USB URL
  and an uploaded signed envelope.
