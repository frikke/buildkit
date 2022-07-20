---
title: COPY command
keywords: build, dockerfile, reference
---

COPY has two forms:

```dockerfile
COPY [--chown=<user>:<group>] <src>... <dest>
COPY [--chown=<user>:<group>] ["<src>",... "<dest>"]
```

This latter form is required for paths containing whitespace

> **Note**
>
> The `--chown` feature is only supported on Dockerfiles used to build Linux containers,
> and will not work on Windows containers. Since user and group ownership concepts do
> not translate between Linux and Windows, the use of `/etc/passwd` and `/etc/group` for
> translating user and group names to IDs restricts this feature to only be viable for
> Linux OS-based containers.

The `COPY` instruction copies new files or directories from `<src>`
and adds them to the filesystem of the container at the path `<dest>`.

Multiple `<src>` resources may be specified but the paths of files and
directories will be interpreted as relative to the source of the context
of the build.

Each `<src>` may contain wildcards and matching will be done using Go's
[filepath.Match](https://golang.org/pkg/path/filepath#Match) rules. For example:

To add all files starting with "hom":

```dockerfile
COPY hom* /mydir/
```

In the example below, `?` is replaced with any single character, e.g., "home.txt".

```dockerfile
COPY hom?.txt /mydir/
```

The `<dest>` is an absolute path, or a path relative to `WORKDIR`, into which
the source will be copied inside the destination container.

The example below uses a relative path, and adds "test.txt" to `<WORKDIR>/relativeDir/`:

```dockerfile
COPY test.txt relativeDir/
```

Whereas this example uses an absolute path, and adds "test.txt" to `/absoluteDir/`

```dockerfile
COPY test.txt /absoluteDir/
```

When copying files or directories that contain special characters (such as `[`
and `]`), you need to escape those paths following the Golang rules to prevent
them from being treated as a matching pattern. For example, to copy a file
named `arr[0].txt`, use the following;

```dockerfile
COPY arr[[]0].txt /mydir/
```

All new files and directories are created with a UID and GID of 0, unless the
optional `--chown` flag specifies a given username, groupname, or UID/GID
combination to request specific ownership of the copied content. The
format of the `--chown` flag allows for either username and groupname strings
or direct integer UID and GID in any combination. Providing a username without
groupname or a UID without GID will use the same numeric UID as the GID. If a
username or groupname is provided, the container's root filesystem
`/etc/passwd` and `/etc/group` files will be used to perform the translation
from name to integer UID or GID respectively. The following examples show
valid definitions for the `--chown` flag:

```dockerfile
COPY --chown=55:mygroup files* /somedir/
COPY --chown=bin files* /somedir/
COPY --chown=1 files* /somedir/
COPY --chown=10:11 files* /somedir/
```

If the container root filesystem does not contain either `/etc/passwd` or
`/etc/group` files and either user or group names are used in the `--chown`
flag, the build will fail on the `COPY` operation. Using numeric IDs requires
no lookup and does not depend on container root filesystem content.

> **Note**
>
> If you build using STDIN (`docker build - < somefile`), there is no
> build context, so `COPY` can't be used.

Optionally `COPY` accepts a flag `--from=<name>` that can be used to set
the source location to a previous build stage (created with `FROM .. AS <name>`)
that will be used instead of a build context sent by the user. In case a build
stage with a specified name can't be found an image with the same name is
attempted to be used instead.

`COPY` obeys the following rules:

- The `<src>` path must be inside the *context* of the build;
  you cannot `COPY ../something /something`, because the first step of a
  `docker build` is to send the context directory (and subdirectories) to the
  docker daemon.

- If `<src>` is a directory, the entire contents of the directory are copied,
  including filesystem metadata.

> **Note**
>
> The directory itself is not copied, just its contents.

- If `<src>` is any other kind of file, it is copied individually along with
  its metadata. In this case, if `<dest>` ends with a trailing slash `/`, it
  will be considered a directory and the contents of `<src>` will be written
  at `<dest>/base(<src>)`.

- If multiple `<src>` resources are specified, either directly or due to the
  use of a wildcard, then `<dest>` must be a directory, and it must end with
  a slash `/`.

- If `<dest>` does not end with a trailing slash, it will be considered a
  regular file and the contents of `<src>` will be written at `<dest>`.

- If `<dest>` doesn't exist, it is created along with all missing directories
  in its path.

> **Note**
>
> The first encountered `COPY` instruction will invalidate the cache for all
> following instructions from the Dockerfile if the contents of `<src>` have
> changed. This includes invalidating the cache for `RUN` instructions.
> See the [`Dockerfile` Best Practices
guide – Leverage build cache](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/#leverage-build-cache)
> for more information.

## COPY --link

> **Note**
>
> Added in [`docker/dockerfile:1.4`](../parser-directives.md#syntax)

Enabling this flag in `COPY` or `ADD` commands allows you to copy files with
enhanced semantics where your files remain independent on their own layer and
don't get invalidated when commands on previous layers are changed.

When `--link` is used your source files are copied into an empty destination
directory. That directory is turned into a layer that is linked on top of your
previous state.

```dockerfile
# syntax=docker/dockerfile:1
FROM alpine
COPY --link /foo /bar
```

Is equivalent of doing two builds:

```dockerfile
FROM alpine
```

and

```dockerfile
FROM scratch
COPY /foo /bar
```

and merging all the layers of both images together.

### Benefits of using `--link`

Use `--link` to reuse already built layers in subsequent builds with
`--cache-from` even if the previous layers have changed. This is especially
important for multi-stage builds where a `COPY --from` statement would
previously get invalidated if any previous commands in the same stage changed,
causing the need to rebuild the intermediate stages again. With `--link` the
layer the previous build generated is reused and merged on top of the new
layers. This also means you can easily rebase your images when the base images
receive updates, without having to execute the whole build again. In backends
that support it, BuildKit can do this rebase action without the need to push or
pull any layers between the client and the registry. BuildKit will detect this
case and only create new image manifest that contains the new layers and old
layers in correct order.

The same behavior where BuildKit can avoid pulling down the base image can also
happen when using `--link` and no other commands that would require access to
the files in the base image. In that case BuildKit will only build the layers
for the `COPY` commands and push them to the registry directly on top of the
layers of the base image.

### Incompatibilities with `--link=false`

When using `--link` the `COPY/ADD` commands are not allowed to read any files
from the previous state. This means that if in previous state the destination
directory was a path that contained a symlink, `COPY/ADD` can not follow it.
In the final image the destination path created with `--link` will always be a
path containing only directories.

If you don't rely on the behavior of following symlinks in the destination
path, using `--link` is always recommended. The performance of `--link` is
equivalent or better than the default behavior and, it creates much better
conditions for cache reuse.
