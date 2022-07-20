---
title: ARG command
keywords: build, dockerfile, reference
---

```dockerfile
ARG <name>[=<default value>]
```

The `ARG` instruction defines a variable that users can pass at build-time to
the builder with the `docker build` command using the `--build-arg <varname>=<value>`
flag. If a user specifies a build argument that was not
defined in the Dockerfile, the build outputs a warning.

```console
[Warning] One or more build-args [foo] were not consumed.
```

A Dockerfile may include one or more `ARG` instructions. For example,
the following is a valid Dockerfile:

```dockerfile
FROM busybox
ARG user1
ARG buildno
# ...
```

> **Warning:**
>
> It is not recommended to use build-time variables for passing secrets like
> GitHub keys, user credentials etc. Build-time variable values are visible to
> any user of the image with the `docker history` command.
>
> Refer to the ["build images with BuildKit"](https://docs.docker.com/develop/develop-images/build_enhancements/#new-docker-build-secret-information)
> section to learn about secure ways to use secrets when building images.
{:.warning}

## Default values

An `ARG` instruction can optionally include a default value:

```dockerfile
FROM busybox
ARG user1=someuser
ARG buildno=1
# ...
```

If an `ARG` instruction has a default value and if there is no value passed
at build-time, the builder uses the default.

## Scope

An `ARG` variable definition comes into effect from the line on which it is
defined in the `Dockerfile` not from the argument's use on the command-line or
elsewhere.  For example, consider this Dockerfile:

```dockerfile
FROM busybox
USER ${user:-some_user}
ARG user
USER $user
# ...
```

A user builds this file by calling:

```console
$ docker build --build-arg user=what_user .
```

The `USER` at line 2 evaluates to `some_user` as the `user` variable is defined on the
subsequent line 3. The `USER` at line 4 evaluates to `what_user` as `user` is
defined and the `what_user` value was passed on the command line. Prior to its definition by an
`ARG` instruction, any use of a variable results in an empty string.

An `ARG` instruction goes out of scope at the end of the build
stage where it was defined. To use an arg in multiple stages, each stage must
include the `ARG` instruction.

```dockerfile
FROM busybox
ARG SETTINGS
RUN ./run/setup $SETTINGS

FROM busybox
ARG SETTINGS
RUN ./run/other $SETTINGS
```

## Using ARG variables

You can use an `ARG` or an `ENV` instruction to specify variables that are
available to the `RUN` instruction. Environment variables defined using the
`ENV` instruction always override an `ARG` instruction of the same name. Consider
this Dockerfile with an `ENV` and `ARG` instruction.

```dockerfile
FROM ubuntu
ARG CONT_IMG_VER
ENV CONT_IMG_VER=v1.0.0
RUN echo $CONT_IMG_VER
```

Then, assume this image is built with this command:

```console
$ docker build --build-arg CONT_IMG_VER=v2.0.1 .
```

In this case, the `RUN` instruction uses `v1.0.0` instead of the `ARG` setting
passed by the user:`v2.0.1` This behavior is similar to a shell
script where a locally scoped variable overrides the variables passed as
arguments or inherited from environment, from its point of definition.

Using the example above but a different `ENV` specification you can create more
useful interactions between `ARG` and `ENV` instructions:

```dockerfile
FROM ubuntu
ARG CONT_IMG_VER
ENV CONT_IMG_VER=${CONT_IMG_VER:-v1.0.0}
RUN echo $CONT_IMG_VER
```

Unlike an `ARG` instruction, `ENV` values are always persisted in the built
image. Consider a docker build without the `--build-arg` flag:

```console
$ docker build .
```

Using this Dockerfile example, `CONT_IMG_VER` is still persisted in the image but
its value would be `v1.0.0` as it is the default set in line 3 by the `ENV` instruction.

The variable expansion technique in this example allows you to pass arguments
from the command line and persist them in the final image by leveraging the
`ENV` instruction. Variable expansion is only supported for [a limited set of
Dockerfile instructions.](env.md#environment-replacement)

## Predefined ARGs

Docker has a set of predefined `ARG` variables that you can use without a
corresponding `ARG` instruction in the Dockerfile.

- `HTTP_PROXY`
- `http_proxy`
- `HTTPS_PROXY`
- `https_proxy`
- `FTP_PROXY`
- `ftp_proxy`
- `NO_PROXY`
- `no_proxy`
- `ALL_PROXY`
- `all_proxy`

To use these, pass them on the command line using the `--build-arg` flag, for
example:

```console
$ docker build --build-arg HTTPS_PROXY=https://my-proxy.example.com .
```

By default, these pre-defined variables are excluded from the output of
`docker history`. Excluding them reduces the risk of accidentally leaking
sensitive authentication information in an `HTTP_PROXY` variable.

For example, consider building the following Dockerfile using
`--build-arg HTTP_PROXY=http://user:pass@proxy.lon.example.com`

```dockerfile
FROM ubuntu
RUN echo "Hello World"
```

In this case, the value of the `HTTP_PROXY` variable is not available in the
`docker history` and is not cached. If you were to change location, and your
proxy server changed to `http://user:pass@proxy.sfo.example.com`, a subsequent
build does not result in a cache miss.

If you need to override this behaviour then you may do so by adding an `ARG`
statement in the Dockerfile as follows:

```dockerfile
FROM ubuntu
ARG HTTP_PROXY
RUN echo "Hello World"
```

When building this Dockerfile, the `HTTP_PROXY` is preserved in the
`docker history`, and changing its value invalidates the build cache.

## Automatic platform ARGs in the global scope

This feature is only available when using the [BuildKit](../index.md#buildkit) backend.

Docker predefines a set of `ARG` variables with information on the platform of
the node performing the build (build platform) and on the platform of the
resulting image (target platform). The target platform can be specified with
the `--platform` flag on `docker build`.

The following `ARG` variables are set automatically:

- `TARGETPLATFORM` - platform of the build result. Eg `linux/amd64`, `linux/arm/v7`, `windows/amd64`.
- `TARGETOS` - OS component of TARGETPLATFORM
- `TARGETARCH` - architecture component of TARGETPLATFORM
- `TARGETVARIANT` - variant component of TARGETPLATFORM
- `BUILDPLATFORM` - platform of the node performing the build.
- `BUILDOS` - OS component of BUILDPLATFORM
- `BUILDARCH` - architecture component of BUILDPLATFORM
- `BUILDVARIANT` - variant component of BUILDPLATFORM

These arguments are defined in the global scope so are not automatically
available inside build stages or for your `RUN` commands. To expose one of
these arguments inside the build stage redefine it without value.

For example:

```dockerfile
FROM alpine
ARG TARGETPLATFORM
RUN echo "I'm building for $TARGETPLATFORM"
```

## BuildKit built-in build args

| Arg                                   | Type    | Description                                                              |
|---------------------------------------|---------|--------------------------------------------------------------------------|
| `BUILDKIT_CACHE_MOUNT_NS`             | String  | Set optional cache ID namespace.                                         |
| `BUILDKIT_CONTEXT_KEEP_GIT_DIR`       | Bool    | Trigger git context to keep the `.git` directory.                        |
| `BUILDKIT_INLINE_BUILDINFO_ATTRS`[^1] | Bool    | Inline build info attributes in image config or not.                     |
| `BUILDKIT_INLINE_CACHE`[^1]           | Bool    | Inline cache metadata to image config or not.                            |
| `BUILDKIT_MULTI_PLATFORM`             | Bool    | Opt into determnistic output regardless of multi-platform output or not. |
| `BUILDKIT_SANDBOX_HOSTNAME`           | String  | Set the hostname (default `buildkitsandbox`)                             |
| `BUILDKIT_SYNTAX`                     | String  | Set frontend image                                                       |

### Example: keep `.git` dir

When using a Git context, `.git` dir is not kept on git checkouts. It can be
useful to keep it around if you want to retrieve git information during
your build:

```dockerfile
# syntax=docker/dockerfile:1
FROM alpine
WORKDIR /src
RUN --mount=target=. \
  make REVISION=$(git rev-parse HEAD) build
```

```console
$ docker build --build-arg BUILDKIT_CONTEXT_KEEP_GIT_DIR=1 https://github.com/user/repo.git#main
```

## Impact on build caching

`ARG` variables are not persisted into the built image as `ENV` variables are.
However, `ARG` variables do impact the build cache in similar ways. If a
Dockerfile defines an `ARG` variable whose value is different from a previous
build, then a "cache miss" occurs upon its first usage, not its definition. In
particular, all `RUN` instructions following an `ARG` instruction use the `ARG`
variable implicitly (as an environment variable), thus can cause a cache miss.
All predefined `ARG` variables are exempt from caching unless there is a
matching `ARG` statement in the `Dockerfile`.

For example, consider these two Dockerfile:

```dockerfile
FROM ubuntu
ARG CONT_IMG_VER
RUN echo $CONT_IMG_VER
```

```dockerfile
FROM ubuntu
ARG CONT_IMG_VER
RUN echo hello
```

If you specify `--build-arg CONT_IMG_VER=<value>` on the command line, in both
cases, the specification on line 2 does not cause a cache miss; line 3 does
cause a cache miss.`ARG CONT_IMG_VER` causes the RUN line to be identified
as the same as running `CONT_IMG_VER=<value> echo hello`, so if the `<value>`
changes, we get a cache miss.

Consider another example under the same command line:

```dockerfile
FROM ubuntu
ARG CONT_IMG_VER
ENV CONT_IMG_VER=$CONT_IMG_VER
RUN echo $CONT_IMG_VER
```

In this example, the cache miss occurs on line 3. The miss happens because
the variable's value in the `ENV` references the `ARG` variable and that
variable is changed through the command line. In this example, the `ENV`
command causes the image to include the value.

If an `ENV` instruction overrides an `ARG` instruction of the same name, like
this Dockerfile:

```dockerfile
FROM ubuntu
ARG CONT_IMG_VER
ENV CONT_IMG_VER=hello
RUN echo $CONT_IMG_VER
```

Line 3 does not cause a cache miss because the value of `CONT_IMG_VER` is a
constant (`hello`). As a result, the environment variables and values used on
the `RUN` (line 4) doesn't change between builds.

___

[^1]: For Docker-integrated BuildKit (`DOCKER_BUILDKIT=1 docker build`) and `docker buildx build`
