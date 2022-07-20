---
title: RUN command
keywords: build, dockerfile, reference
---

RUN has 2 forms:

- `RUN <command>` (*shell* form, the command is run in a shell, which by
  default is `/bin/sh -c` on Linux or `cmd /S /C` on Windows)
- `RUN ["executable", "param1", "param2"]` (*exec* form)

The `RUN` instruction will execute any commands in a new layer on top of the
current image and commit the results. The resulting committed image will be
used for the next step in the `Dockerfile`.

Layering `RUN` instructions and generating commits conforms to the core
concepts of Docker where commits are cheap and containers can be created from
any point in an image's history, much like source control.

The *exec* form makes it possible to avoid shell string munging, and to `RUN`
commands using a base image that does not contain the specified shell executable.

The default shell for the *shell* form can be changed using the `SHELL`
command.

In the *shell* form you can use a `\` (backslash) to continue a single
RUN instruction onto the next line. For example, consider these two lines:

```dockerfile
RUN /bin/bash -c 'source $HOME/.bashrc; \
echo $HOME'
```

Together they are equivalent to this single line:

```dockerfile
RUN /bin/bash -c 'source $HOME/.bashrc; echo $HOME'
```

To use a different shell, other than '/bin/sh', use the *exec* form passing in
the desired shell. For example:

```dockerfile
RUN ["/bin/bash", "-c", "echo hello"]
```

> **Note**
>
> The *exec* form is parsed as a JSON array, which means that
> you must use double-quotes (") around words not single-quotes (').

Unlike the *shell* form, the *exec* form does not invoke a command shell.
This means that normal shell processing does not happen. For example,
`RUN [ "echo", "$HOME" ]` will not do variable substitution on `$HOME`.
If you want shell processing then either use the *shell* form or execute
a shell directly, for example: `RUN [ "sh", "-c", "echo $HOME" ]`.
When using the exec form and executing a shell directly, as in the case for
the shell form, it is the shell that is doing the environment variable
expansion, not docker.

> **Note**
>
> In the *JSON* form, it is necessary to escape backslashes. This is
> particularly relevant on Windows where the backslash is the path separator.
> The following line would otherwise be treated as *shell* form due to not
> being valid JSON, and fail in an unexpected way:
>
> ```dockerfile
> RUN ["c:\windows\system32\tasklist.exe"]
> ```
>
> The correct syntax for this example is:
>
> ```dockerfile
> RUN ["c:\\windows\\system32\\tasklist.exe"]
> ```

The cache for `RUN` instructions isn't invalidated automatically during
the next build. The cache for an instruction like
`RUN apt-get dist-upgrade -y` will be reused during the next build. The
cache for `RUN` instructions can be invalidated by using the `--no-cache`
flag, for example `docker build --no-cache`.

See the [`Dockerfile` Best Practices
guide](https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/)
for more information.

The cache for `RUN` instructions can be invalidated by [`ADD`](add.md) and
[`COPY`](copy.md) instructions.

### Known issues

- [Issue 783](https://github.com/docker/docker/issues/783) is about file
  permissions problems that can occur when using the AUFS file system. You
  might notice it during an attempt to `rm` a file, for example.

  For systems that have recent aufs version (i.e., `dirperm1` mount option can
  be set), docker will attempt to fix the issue automatically by mounting
  the layers with `dirperm1` option. More details on `dirperm1` option can be
  found at [`aufs` man page](https://github.com/sfjro/aufs3-linux/tree/aufs3.18/Documentation/filesystems/aufs)

  If your system doesn't have support for `dirperm1`, the issue describes a workaround.

## RUN --mount

> **Note**
>
> Added in [`docker/dockerfile:1.2`](../parser-directives.md#syntax)

`RUN --mount` allows you to create mounts that process running as part of the
build can access. This can be used to bind files from other part of the build
without copying, accessing build secrets or ssh-agent sockets, or creating cache
locations to speed up your build.

Syntax: `--mount=[type=<TYPE>][,option=<value>[,option=<value>]...]`

### Mount types

| Type                                     | Description                                                                                               |
|------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| [`bind`](#run---mounttypebind) (default) | Bind-mount context directories (read-only).                                                               |
| [`cache`](#run---mounttypecache)         | Mount a temporary directory for to cache directories for compilers and package managers.                  |
| [`secret`](#run---mounttypesecret)       | Allow the build container to access SSH keys via SSH agents, with support for passphrases.                |
| [`ssh`](#run---mounttypessh)             | Allow the build container to access secure files such as private keys without baking them into the image. |

### RUN --mount=type=bind

This mount type allows binding directories (read-only) in the context or in an
image to the build container.

| Option               | Description                                                                          |
|----------------------|--------------------------------------------------------------------------------------|
| `target`[^1]         | Mount path.                                                                          |
| `source`             | Source path in the `from`. Defaults to the root of the `from`.                       |
| `from`               | Build stage or image name for the root of the source. Defaults to the build context. |
| `rw`,`readwrite`     | Allow writes on the mount. Written data will be discarded.                           |

### RUN --mount=type=cache

This mount type allows the build container to cache directories for compilers
and package managers.

| Option              | Description                                                                                                                                                                                                                                                                |
|---------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `id`                | Optional ID to identify separate/different caches. Defaults to value of `target`.                                                                                                                                                                                          |
| `target`[^1]        | Mount path.                                                                                                                                                                                                                                                                |
| `ro`,`readonly`     | Read-only if set.                                                                                                                                                                                                                                                          |
| `sharing`           | One of `shared`, `private`, or `locked`. Defaults to `shared`. A `shared` cache mount can be used concurrently by multiple writers. `private` creates a new mount if there are multiple writers. `locked` pauses the second writer until the first one releases the mount. |
| `from`              | Build stage to use as a base of the cache mount. Defaults to empty directory.                                                                                                                                                                                              |
| `source`            | Subpath in the `from` to mount. Defaults to the root of the `from`.                                                                                                                                                                                                        |
| `mode`              | File mode for new cache directory in octal. Default `0755`.                                                                                                                                                                                                                |
| `uid`               | User ID for new cache directory. Default `0`.                                                                                                                                                                                                                              |
| `gid`               | Group ID for new cache directory. Default `0`.                                                                                                                                                                                                                             |

Contents of the cache directories persists between builder invocations without
invalidating the instruction cache. Cache mounts should only be used for better
performance. Your build should work with any contents of the cache directory as
another build may overwrite the files or GC may clean it if more storage space
is needed.

#### Example: cache Go packages

```dockerfile
# syntax=docker/dockerfile:1
FROM golang
RUN --mount=type=cache,target=/root/.cache/go-build \
  go build ...
```

#### Example: cache apt packages

```dockerfile
# syntax=docker/dockerfile:1
FROM ubuntu
RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt \
  --mount=type=cache,target=/var/lib/apt \
  apt update && apt-get --no-install-recommends install -y gcc
```

### RUN --mount=type=tmpfs

This mount type allows mounting tmpfs in the build container.

| Option              | Description                                           |
|---------------------|-------------------------------------------------------|
| `target`[^1]        | Mount path.                                           |
| `size`              | Specify an upper limit on the size of the filesystem. |

### RUN --mount=type=secret

This mount type allows the build container to access secure files such as
private keys without baking them into the image.

| Option              | Description                                                                                       |
|---------------------|---------------------------------------------------------------------------------------------------|
| `id`                | ID of the secret. Defaults to basename of the target path.                                        |
| `target`            | Mount path. Defaults to `/run/secrets/` + `id`.                                                   |
| `required`          | If set to `true`, the instruction errors out when the secret is unavailable. Defaults to `false`. |
| `mode`              | File mode for secret file in octal. Default `0400`.                                               |
| `uid`               | User ID for secret file. Default `0`.                                                             |
| `gid`               | Group ID for secret file. Default `0`.                                                            |

#### Example: access to S3

```dockerfile
# syntax=docker/dockerfile:1
FROM python:3
RUN pip install awscli
RUN --mount=type=secret,id=aws,target=/root/.aws/credentials \
  aws s3 cp s3://... ...
```

```console
$ docker buildx build --secret id=aws,src=$HOME/.aws/credentials .
```

### RUN --mount=type=ssh

This mount type allows the build container to access SSH keys via SSH agents,
with support for passphrases.

| Option              | Description                                                                                    |
|---------------------|------------------------------------------------------------------------------------------------|
| `id`                | ID of SSH agent socket or key. Defaults to "default".                                          |
| `target`            | SSH agent socket path. Defaults to `/run/buildkit/ssh_agent.${N}`.                             |
| `required`          | If set to `true`, the instruction errors out when the key is unavailable. Defaults to `false`. |
| `mode`              | File mode for socket in octal. Default `0600`.                                                 |
| `uid`               | User ID for socket. Default `0`.                                                               |
| `gid`               | Group ID for socket. Default `0`.                                                              |

#### Example: access to Gitlab

```dockerfile
# syntax=docker/dockerfile:1
FROM alpine
RUN apk add --no-cache openssh-client
RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan gitlab.com >> ~/.ssh/known_hosts
RUN --mount=type=ssh \
  ssh -q -T git@gitlab.com 2>&1 | tee /hello
# "Welcome to GitLab, @GITLAB_USERNAME_ASSOCIATED_WITH_SSHKEY" should be printed here
# with the type of build progress is defined as `plain`.
```

```console
$ eval $(ssh-agent)
$ ssh-add ~/.ssh/id_rsa
(Input your passphrase here)
$ docker buildx build --ssh default=$SSH_AUTH_SOCK .
```

You can also specify a path to `*.pem` file on the host directly instead of `$SSH_AUTH_SOCK`.
However, pem files with passphrases are not supported.

## RUN --network

> **Note**
>
> Added in [`docker/dockerfile:1.1`](../parser-directives.md#syntax)

`RUN --network` allows control over which networking environment the command
is run in.

Syntax: `--network=<TYPE>`

### Network types

| Type                                         | Description                            |
|----------------------------------------------|----------------------------------------|
| [`default`](#run---networkdefault) (default) | Run in the default network.            |
| [`none`](#run---networknone)                 | Run with no network access.            |
| [`host`](#run---networkhost)                 | Run in the host's network environment. |

### RUN --network=default

Equivalent to not supplying a flag at all, the command is run in the default
network for the build.

### RUN --network=none

The command is run with no network access (`lo` is still available, but is
isolated to this process)

#### Example: isolating external effects

```dockerfile
# syntax=docker/dockerfile:1
FROM python:3.6
ADD mypackage.tgz wheels/
RUN --network=none pip install --find-links wheels mypackage
```

`pip` will only be able to install the packages provided in the tarfile, which
can be controlled by an earlier build stage.

### RUN --network=host

The command is run in the host's network environment (similar to
`docker build --network=host`, but on a per-instruction basis)

> **Warning**
>
> The use of `--network=host` is protected by the `network.host` entitlement,
> which needs to be enabled when starting the buildkitd daemon with
> `--allow-insecure-entitlement network.host` flag or in [buildkitd config](https://github.com/moby/buildkit/blob/master/docs/buildkitd.toml.md),
> and for a build request with [`--allow network.host` flag](https://docs.docker.com/engine/reference/commandline/buildx_build/#allow).
{:.warning}

## RUN --security

> **Note**
>
> Not yet available in stable syntax, use [`docker/dockerfile:1-labs`](../parser-directives.md#syntax) version.

### RUN --security=insecure

With `--security=insecure`, builder runs the command without sandbox in insecure
mode, which allows to run flows requiring elevated privileges (e.g. containerd).
This is equivalent to running `docker run --privileged`.

> **Warning**
>
> In order to access this feature, entitlement `security.insecure` should be
> enabled when starting the buildkitd daemon with
> `--allow-insecure-entitlement security.insecure` flag or in [buildkitd config](https://github.com/moby/buildkit/blob/master/docs/buildkitd.toml.md),
> and for a build request with [`--allow security.insecure` flag](https://docs.docker.com/engine/reference/commandline/buildx_build/#allow).
{:.warning}

#### Example: check entitlements

```dockerfile
# syntax=docker/dockerfile:1-labs
FROM ubuntu
RUN --security=insecure cat /proc/self/status | grep CapEff
```
```text
#84 0.093 CapEff:	0000003fffffffff
```

### RUN --security=sandbox

Default sandbox mode can be activated via `--security=sandbox`, but that is no-op.

___

[^1]: Value required
