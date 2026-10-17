# TODO 059 - Accept Relay URL From the Environment

Allow `make relay` to use `PROMISE_ADVERTISE_URL` directly as its relay URL so
the documented launcher command cannot omit a required Make variable.

Completed 2026-10-17. A dry Make expansion confirms that the environment URL
is passed to the relay launcher.
