# TODO 066 - Fix Relay Setting Persistence and Verification

The phone repeatedly retained the old LAN relay URL after attempts to update
it. Add an explicit **Use USB relay** action that saves the known USB address
without clipboard or typing, and show the currently persisted relay URL on the
screen so a person can verify the setting before sending.
