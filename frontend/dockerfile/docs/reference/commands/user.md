---
title: USER command
keywords: build, dockerfile, reference
---

```dockerfile
USER <user>[:<group>]
```

or

```dockerfile
USER <UID>[:<GID>]
```

The `USER` instruction sets the user name (or UID) and optionally the user
group (or GID) to use as the default user and group for the remainder of the
current stage. The specified user is used for `RUN` instructions and at
runtime, runs the relevant `ENTRYPOINT` and `CMD` commands.

> Note that when specifying a group for the user, the user will have _only_ the
> specified group membership. Any other configured group memberships will be ignored.

> **Warning**
>
> When the user doesn't have a primary group then the image (or the next
> instructions) will be run with the `root` group.
>
> On Windows, the user must be created first if it's not a built-in account.
> This can be done with the `net user` command called as part of a Dockerfile.

```dockerfile
FROM microsoft/windowsservercore
# Create Windows user in the container
RUN net user /add patrick
# Set it for subsequent commands
USER patrick
```
