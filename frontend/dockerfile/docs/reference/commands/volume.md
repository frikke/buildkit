---
title: VOLUME command
keywords: build, dockerfile, reference
---

```dockerfile
VOLUME ["/data"]
```

The `VOLUME` instruction creates a mount point with the specified name
and marks it as holding externally mounted volumes from native host or other
containers. The value can be a JSON array, `VOLUME ["/var/log/"]`, or a plain
string with multiple arguments, such as `VOLUME /var/log` or `VOLUME /var/log
/var/db`. For more information/examples and mounting instructions via the
Docker client, refer to
[*Share Directories via Volumes*](https://docs.docker.com/storage/volumes/)
documentation.

The `docker run` command initializes the newly created volume with any data
that exists at the specified location within the base image. For example,
consider the following Dockerfile snippet:

```dockerfile
FROM ubuntu
RUN mkdir /myvol
RUN echo "hello world" > /myvol/greeting
VOLUME /myvol
```

This Dockerfile results in an image that causes `docker run` to
create a new mount point at `/myvol` and copy the  `greeting` file
into the newly created volume.

### Notes about specifying volumes

Keep the following things in mind about volumes in the `Dockerfile`.

- **Volumes on Windows-based containers**: When using Windows-based containers,
  the destination of a volume inside the container must be one of:

    - a non-existing or empty directory
    - a drive other than `C:`

- **Changing the volume from within the Dockerfile**: If any build steps change the
  data within the volume after it has been declared, those changes will be discarded.

- **JSON formatting**: The list is parsed as a JSON array.
  You must enclose words with double quotes (`"`) rather than single quotes (`'`).

- **The host directory is declared at container run-time**: The host directory
  (the mountpoint) is, by its nature, host-dependent. This is to preserve image
  portability, since a given host directory can't be guaranteed to be available
  on all hosts. For this reason, you can't mount a host directory from
  within the Dockerfile. The `VOLUME` instruction does not support specifying a `host-dir`
  parameter.  You must specify the mountpoint when you create or run the container.
