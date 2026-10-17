# TODO 037 - Send a Signed Promise to the Second Phone

Enter the recipient CID, author a promise, queue it, and verify relay upload
state. This is a manual device action; do not check it off until the values
below are recorded in private demo notes.

1. On phone B, tap **Copy device identity**.
2. Transfer that clipboard value to phone A, then on A tap **Paste recipient
   identity**. The field must contain the complete lowercase `b...` CID.
3. On A, enter a distinctive non-secret promise, for example
   `POC-DELIVERY-2026-10-17: I will review this`.
4. Tap **Sign, queue, and upload** once. Record the complete message CID from
   the status line. The app has already persisted the exact signed bytes before
   upload begins.
5. Wait for the status line and outbox to show `UPLOADED`. Record its attempt
   count and screenshot the full message CID.
6. If it shows `RETRYABLE_FAILURE`, preserve the error and attempt count;
   correct the relay URL/connectivity and use **Retry pending uploads**. Do not
   compose a replacement message when demonstrating retry.

Expected boundary: `UPLOADED` means the relay retained bytes. It does not mean
phone B accepted, executed, or trusted the promise. Continue with TODO 038.
