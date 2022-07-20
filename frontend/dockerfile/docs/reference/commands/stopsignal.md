---
title: STOPSIGNAL command
keywords: build, dockerfile, reference
---

```dockerfile
STOPSIGNAL signal
```

The `STOPSIGNAL` instruction sets the system call signal that will be sent to the
container to exit. This signal can be a signal name in the format `SIG<NAME>`,
for instance `SIGKILL`, or an unsigned number that matches a position in the
kernel's syscall table, for instance `9`. The default is `SIGTERM` if not
defined.

The image's default stopsignal can be overridden per container, using the
`--stop-signal` flag on `docker run` and `docker create`.
