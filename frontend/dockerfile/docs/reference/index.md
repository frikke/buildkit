---
title: Dockerfile reference
description: "Dockerfiles use a simple DSL which allows you to automate the steps you would normally manually take to create an image."
keywords: build, dockerfile, reference
redirect_from:
- /reference/builder/
---

Docker can build images automatically by reading the instructions from a
`Dockerfile`. A `Dockerfile` is a text document that contains all the commands a
user could call on the command line to assemble an image. Using `docker build`
users can create an automated build that executes several command-line
instructions in succession.

This page describes the commands you can use in a `Dockerfile`. When you are
done reading this page, refer to the [`Dockerfile` Best Practices](https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/)
for a tip-oriented guide.

## Usage

The [docker build](https://docs.docker.com/engine/reference/commandline/build/)
command builds an image from a `Dockerfile` and a *context*. The build's context
is the set of files at a specified location `PATH` or `URL`. The `PATH` is a
directory on your local filesystem. The `URL` is a Git repository location.

The build context is processed recursively. So, a `PATH` includes any subdirectories
and the `URL` includes the repository and its submodules. This example shows a
build command that uses the current directory (`.`) as build context:

```console
$ docker build .

Sending build context to Docker daemon  6.51 MB
...
```

The build is run by the Docker daemon, not by the CLI. The first thing a build
process does is send the entire context (recursively) to the daemon.  In most
cases, it's best to start with an empty directory as context and keep your
Dockerfile in that directory. Add only the files needed for building the
Dockerfile.

> **Warning**
>
> Do not use your root directory, `/`, as the `PATH` for  your build context, as
> it causes the build to transfer the entire contents of your hard drive to the
> Docker daemon.
{:.warning}

To use a file in the build context, the `Dockerfile` refers to the file specified
in an instruction, for example,  a `COPY` instruction. To increase the build's
performance, exclude files and directories by adding a `.dockerignore` file to
the context directory.  For information about how to [create a `.dockerignore`
file](#dockerignore-file) see the documentation on this page.

Traditionally, the `Dockerfile` is called `Dockerfile` and located in the root
of the context. You use the `-f` flag with `docker build` to point to a Dockerfile
anywhere in your file system.

```console
$ docker build -f /path/to/a/Dockerfile .
```

You can specify a repository and tag at which to save the new image if
the build succeeds:

```console
$ docker build -t shykes/myapp .
```

To tag the image into multiple repositories after the build,
add multiple `-t` parameters when you run the `build` command:

```console
$ docker build -t shykes/myapp:1.0.2 -t shykes/myapp:latest .
```

Before the Docker daemon runs the instructions in the `Dockerfile`, it performs
a preliminary validation of the `Dockerfile` and returns an error if the syntax is incorrect:

```console
$ docker build -t test/myapp .

[+] Building 0.3s (2/2) FINISHED
 => [internal] load build definition from Dockerfile                       0.1s
 => => transferring dockerfile: 60B                                        0.0s
 => [internal] load .dockerignore                                          0.1s
 => => transferring context: 2B                                            0.0s
error: failed to solve: rpc error: code = Unknown desc = failed to solve with frontend dockerfile.v0: failed to create LLB definition:
dockerfile parse error line 2: unknown instruction: RUNCMD
```

The Docker daemon runs the instructions in the `Dockerfile` one-by-one,
committing the result of each instruction
to a new image if necessary, before finally outputting the ID of your
new image. The Docker daemon will automatically clean up the context you
sent.

Note that each instruction is run independently, and causes a new image
to be created - so `RUN cd /tmp` will not have any effect on the next
instructions.

Whenever possible, Docker uses a build-cache to accelerate the `docker build`
process significantly. This is indicated by the `CACHED` message in the console
output. (For more information, see the [`Dockerfile` best practices guide](https://docs.docker.com/engine/userguide/eng-image/dockerfile_best-practices/)):

```console
$ docker build -t svendowideit/ambassador .

[+] Building 0.7s (6/6) FINISHED
 => [internal] load build definition from Dockerfile                       0.1s
 => => transferring dockerfile: 286B                                       0.0s
 => [internal] load .dockerignore                                          0.1s
 => => transferring context: 2B                                            0.0s
 => [internal] load metadata for docker.io/library/alpine:3.2              0.4s
 => CACHED [1/2] FROM docker.io/library/alpine:3.2@sha256:e9a2035f9d0d7ce  0.0s
 => CACHED [2/2] RUN apk add --no-cache socat                              0.0s
 => exporting to image                                                     0.0s
 => => exporting layers                                                    0.0s
 => => writing image sha256:1affb80ca37018ac12067fa2af38cc5bcc2a8f09963de  0.0s
 => => naming to docker.io/svendowideit/ambassador                         0.0s
```

By default, the build cache is based on results from previous builds on the machine
on which you are building. The `--cache-from` option also allows you to use a
build-cache that's distributed through an image registry refer to the
[specifying external cache sources](https://docs.docker.com/engine/reference/commandline/build/#specifying-external-cache-sources)
section in the `docker build` command reference.

When you're done with your build, you're ready to look into [scanning your image with `docker scan`](https://docs.docker.com/engine/scan/),
and [pushing your image to Docker Hub](https://docs.docker.com/docker-hub/repos/).

## BuildKit

Starting with version 18.09, Docker supports a new backend for executing your
builds that is provided by the [moby/buildkit](https://github.com/moby/buildkit)
project. The BuildKit backend provides many benefits compared to the old
implementation. For example, BuildKit can:

- Detect and skip executing unused build stages
- Parallelize building independent build stages
- Incrementally transfer only the changed files in your build context between builds
- Detect and skip transferring unused files in your build context
- Use external Dockerfile implementations with many new features
- Avoid side-effects with rest of the API (intermediate images and containers)
- Prioritize your build cache for automatic pruning

To use the BuildKit backend, you need to set an environment variable
`DOCKER_BUILDKIT=1` on the CLI before invoking `docker build`. [Docker Buildx](https://github.com/docker/buildx)
always enables BuildKit.

### External Dockerfile frontend

BuildKit supports loading frontends dynamically from container images. To use
an external Dockerfile frontend, the first line of your Dockerfile needs to be
`# syntax=docker/dockerfile:1` pointing to the specific image you want to use.

BuildKit also ships with Dockerfile frontend builtin, but it is recommended to
use an external image to make sure that all users use the same version on the
builder and to pick up bugfixes automatically without waiting for a new version
of BuildKit or Docker engine.

See the [`syntax` directive section](#syntax) for more information.

## Format

Here is the format of the `Dockerfile`:

```dockerfile
# Comment
INSTRUCTION arguments
```

The instruction is not case-sensitive. However, convention is for them to
be UPPERCASE to distinguish them from arguments more easily.

Docker runs instructions in a `Dockerfile` in order. A `Dockerfile` **must
begin with a `FROM` instruction**. This may be after [parser
directives](#parser-directives), [comments](#format), and globally scoped
[ARGs](commands/arg.md). The [`FROM` instruction](commands/from.md) specifies
the [*Parent Image*](https://docs.docker.com/glossary/#parent-image) from which
you are building. `FROM` may only be preceded by one or more `ARG` instructions,
which declare arguments that are used in `FROM` lines in the `Dockerfile`.

Docker treats lines that *begin* with `#` as a comment, unless the line is
a valid [parser directive](#parser-directives). A `#` marker anywhere
else in a line is treated as an argument. This allows statements like:

```dockerfile
# Comment
RUN echo 'we are running some # of cool things'
```

Comment lines are removed before the Dockerfile instructions are executed, which
means that the comment in the following example is not handled by the shell
executing the `echo` command, and both examples below are equivalent:

```dockerfile
RUN echo hello \
# comment
world
```

```dockerfile
RUN echo hello \
world
```

Line continuation characters are not supported in comments.

> **Note on whitespace**
>
> For backward compatibility, leading whitespace before comments (`#`) and
> instructions (such as `RUN`) are ignored, but discouraged. Leading whitespace
> is not preserved in these cases, and the following examples are therefore
> equivalent:
>
> ```dockerfile
>         # this is a comment-line
>     RUN echo hello
> RUN echo world
> ```
>
> ```dockerfile
> # this is a comment-line
> RUN echo hello
> RUN echo world
> ```
>
> Note however, that whitespace in instruction _arguments_, such as the commands
> following `RUN`, are preserved, so the following example prints `    hello    world`
> with leading whitespace as specified:
>
> ```dockerfile
> RUN echo "\
>      hello\
>      world"
> ```

## Commands

* [`FROM`](commands/from.md)
* [`RUN`](commands/run.md)
* [`CMD`](commands/cmd.md)
* [`LABEL`](commands/label.md)
* [`MAINTAINER`](commands/maintainer.md)
* [`EXPOSE`](commands/expose.md)
* [`ENV`](commands/env.md)
* [`ADD`](commands/add.md)
* [`COPY`](commands/copy.md)
* [`ENTRYPOINT`](commands/entrypoint.md)
* [`VOLUME`](commands/volume.md)
* [`USER`](commands/user.md)
* [`WORKDIR`](commands/workdir.md)
* [`ARG`](commands/arg.md)
* [`ONBUILD`](commands/onbuild.md)
* [`STOPSIGNAL`](commands/stopsignal.md)
* [`HEALTHCHECK`](commands/healthcheck.md)
* [`SHELL`](commands/shell.md)

## Parser directives

Parser directives are optional, and affect the way in which subsequent lines
in a `Dockerfile` are handled. Parser directives do not add layers to the build,
and will not be shown as a build step. Parser directives are written as a
special type of comment in the form `# directive=value`. A single directive
may only be used once.

Once a comment, empty line or builder instruction has been processed, Docker
no longer looks for parser directives. Instead it treats anything formatted
as a parser directive as a comment and does not attempt to validate if it might
be a parser directive. Therefore, all parser directives must be at the very
top of a `Dockerfile`.

Parser directives are not case-sensitive. However, convention is for them to
be lowercase. Convention is also to include a blank line following any
parser directives. Line continuation characters are not supported in parser
directives.

Due to these rules, the following examples are all invalid:

Invalid due to line continuation:

```dockerfile
# direc \
tive=value
```

Invalid due to appearing twice:

```dockerfile
# directive=value1
# directive=value2

FROM ImageName
```

Treated as a comment due to appearing after a builder instruction:

```dockerfile
FROM ImageName
# directive=value
```

Treated as a comment due to appearing after a comment which is not a parser
directive:

```dockerfile
# About my dockerfile
# directive=value
FROM ImageName
```

The unknown directive is treated as a comment due to not being recognized. In
addition, the known directive is treated as a comment due to appearing after
a comment which is not a parser directive.

```dockerfile
# unknowndirective=value
# knowndirective=value
```

Non line-breaking whitespace is permitted in a parser directive. Hence, the
following lines are all treated identically:

```dockerfile
#directive=value
# directive =value
#	directive= value
# directive = value
#	  dIrEcTiVe=value
```

The following parser directives are supported:

- `syntax`
- `escape`

### syntax

<a name="external-implementation-features"><!-- included for deep-links to old section --></a>

```dockerfile
# syntax=[remote image reference]
```

For example:

```dockerfile
# syntax=docker/dockerfile:1
# syntax=docker.io/docker/dockerfile:1
# syntax=example.com/user/repo:tag@sha256:abcdef...
```

This feature is only available when using the [BuildKit](#buildkit) backend,
and is ignored when using the classic builder backend.

The syntax directive defines the location of the Dockerfile syntax that is used
to build the Dockerfile. The BuildKit backend allows to seamlessly use external
implementations that are distributed as Docker images and execute inside a
container sandbox environment.

Custom Dockerfile implementations allows you to:

- Automatically get bugfixes without updating the Docker daemon
- Make sure all users are using the same implementation to build your Dockerfile
- Use the latest features without updating the Docker daemon
- Try out new features or third-party features before they are integrated in the Docker daemon
- Use [alternative build definitions, or create your own](https://github.com/moby/buildkit#exploring-llb)

#### Official releases

Docker distributes official versions of the images that can be used for building
Dockerfiles under `docker/dockerfile` repository on Docker Hub. There are two
channels where new images are released: `stable` and `labs`.

Stable channel follows [semantic versioning](https://semver.org). For example:

- `docker/dockerfile:1` - kept updated with the latest `1.x.x` minor _and_ patch release
- `docker/dockerfile:1.2` -  kept updated with the latest `1.2.x` patch release,
  and stops receiving updates once version `1.3.0` is released.
- `docker/dockerfile:1.2.1` - immutable: never updated

We recommend using `docker/dockerfile:1`, which always points to the latest stable
release of the version 1 syntax, and receives both "minor" and "patch" updates
for the version 1 release cycle. BuildKit automatically checks for updates of the
syntax when performing a build, making sure you are using the most current version.

If a specific version is used, such as `1.2` or `1.2.1`, the Dockerfile needs to
be updated manually to continue receiving bugfixes and new features. Old versions
of the Dockerfile remain compatible with the new versions of the builder.

**labs channel**

The "labs" channel provides early access to Dockerfile features that are not yet
available in the stable channel. Labs channel images are released in conjunction
with the stable releases, and follow the same versioning with the `-labs` suffix,
for example:

- `docker/dockerfile:labs` - latest release on labs channel
- `docker/dockerfile:1-labs` - same as `dockerfile:1` in the stable channel, with labs features enabled
- `docker/dockerfile:1.2-labs` -  same as `dockerfile:1.2` in the stable channel, with labs features enabled
- `docker/dockerfile:1.2.1-labs` - immutable: never updated. Same as `dockerfile:1.2.1` in the stable channel, with labs features enabled

Choose a channel that best fits your needs; if you want to benefit from
new features, use the labs channel. Images in the labs channel provide a superset
of the features in the stable channel; note that `stable` features in the labs
channel images follow [semantic versioning](https://semver.org), but "labs"
features do not, and newer releases may not be backwards compatible, so it
is recommended to use an immutable full version variant.

For documentation on "labs" features, master builds, and nightly feature releases,
refer to the description in [the BuildKit source repository on GitHub](https://github.com/moby/buildkit/blob/master/README.md).
For a full list of available images, visit the [image repository on Docker Hub](https://hub.docker.com/r/docker/dockerfile),
and the [docker/dockerfile-upstream image repository](https://hub.docker.com/r/docker/dockerfile-upstream)
for development builds.

### escape

```dockerfile
# escape=\ (backslash)
```

Or

```dockerfile
# escape=` (backtick)
```

The `escape` directive sets the character used to escape characters in a
`Dockerfile`. If not specified, the default escape character is `\`.

The escape character is used both to escape characters in a line, and to
escape a newline. This allows a `Dockerfile` instruction to
span multiple lines. Note that regardless of whether the `escape` parser
directive is included in a `Dockerfile`, *escaping is not performed in
a `RUN` command, except at the end of a line.*

Setting the escape character to `` ` `` is especially useful on
`Windows`, where `\` is the directory path separator. `` ` `` is consistent
with [Windows PowerShell](https://technet.microsoft.com/en-us/library/hh847755.aspx).

Consider the following example which would fail in a non-obvious way on
`Windows`. The second `\` at the end of the second line would be interpreted as an
escape for the newline, instead of a target of the escape from the first `\`.
Similarly, the `\` at the end of the third line would, assuming it was actually
handled as an instruction, cause it be treated as a line continuation. The result
of this dockerfile is that second and third lines are considered a single
instruction:

```dockerfile
FROM microsoft/nanoserver
COPY testfile.txt c:\\
RUN dir c:\
```

Results in:

```console
PS E:\myproject> docker build -t cmd .

Sending build context to Docker daemon 3.072 kB
Step 1/2 : FROM microsoft/nanoserver
 ---> 22738ff49c6d
Step 2/2 : COPY testfile.txt c:\RUN dir c:
GetFileAttributesEx c:RUN: The system cannot find the file specified.
PS E:\myproject>
```

One solution to the above would be to use `/` as the target of both the `COPY`
instruction, and `dir`. However, this syntax is, at best, confusing as it is not
natural for paths on `Windows`, and at worst, error prone as not all commands on
`Windows` support `/` as the path separator.

By adding the `escape` parser directive, the following `Dockerfile` succeeds as
expected with the use of natural platform semantics for file paths on `Windows`:

```dockerfile
# escape=`

FROM microsoft/nanoserver
COPY testfile.txt c:\
RUN dir c:\
```

Results in:

```console
PS E:\myproject> docker build -t succeeds --no-cache=true .

Sending build context to Docker daemon 3.072 kB
Step 1/3 : FROM microsoft/nanoserver
 ---> 22738ff49c6d
Step 2/3 : COPY testfile.txt c:\
 ---> 96655de338de
Removing intermediate container 4db9acbb1682
Step 3/3 : RUN dir c:\
 ---> Running in a2c157f842f5
 Volume in drive C has no label.
 Volume Serial Number is 7E6D-E0F7

 Directory of c:\

10/05/2016  05:04 PM             1,894 License.txt
10/05/2016  02:22 PM    <DIR>          Program Files
10/05/2016  02:14 PM    <DIR>          Program Files (x86)
10/28/2016  11:18 AM                62 testfile.txt
10/28/2016  11:20 AM    <DIR>          Users
10/28/2016  11:20 AM    <DIR>          Windows
           2 File(s)          1,956 bytes
           4 Dir(s)  21,259,096,064 bytes free
 ---> 01c7f3bef04f
Removing intermediate container a2c157f842f5
Successfully built 01c7f3bef04f
PS E:\myproject>
```

## `.dockerignore` file

Before the docker CLI sends the context to the docker daemon, it looks
for a file named `.dockerignore` in the root directory of the context.
If this file exists, the CLI modifies the context to exclude files and
directories that match patterns in it.  This helps to avoid
unnecessarily sending large or sensitive files and directories to the
daemon and potentially adding them to images using `ADD` or `COPY`.

The CLI interprets the `.dockerignore` file as a newline-separated
list of patterns similar to the file globs of Unix shells.  For the
purposes of matching, the root of the context is considered to be both
the working and the root directory.  For example, the patterns
`/foo/bar` and `foo/bar` both exclude a file or directory named `bar`
in the `foo` subdirectory of `PATH` or in the root of the git
repository located at `URL`.  Neither excludes anything else.

If a line in `.dockerignore` file starts with `#` in column 1, then this line is
considered as a comment and is ignored before interpreted by the CLI.

Here is an example `.dockerignore` file:

```gitignore
# comment
*/temp*
*/*/temp*
temp?
```

This file causes the following build behavior:

| Rule        | Behavior                                                                                                                                                                                                       |
|:------------|:---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `# comment` | Ignored.                                                                                                                                                                                                       |
| `*/temp*`   | Exclude files and directories whose names start with `temp` in any immediate subdirectory of the root.  For example, the plain file `/somedir/temporary.txt` is excluded, as is the directory `/somedir/temp`. |
| `*/*/temp*` | Exclude files and directories starting with `temp` from any subdirectory that is two levels below the root. For example, `/somedir/subdir/temporary.txt` is excluded.                                          |
| `temp?`     | Exclude files and directories in the root directory whose names are a one-character extension of `temp`.  For example, `/tempa` and `/tempb` are excluded.                                                     |


Matching is done using Go's
[filepath.Match](https://golang.org/pkg/path/filepath#Match) rules.  A
preprocessing step removes leading and trailing whitespace and
eliminates `.` and `..` elements using Go's
[filepath.Clean](https://golang.org/pkg/path/filepath/#Clean).  Lines
that are blank after preprocessing are ignored.

Beyond Go's filepath.Match rules, Docker also supports a special
wildcard string `**` that matches any number of directories (including
zero). For example, `**/*.go` will exclude all files that end with `.go`
that are found in all directories, including the root of the build context.

Lines starting with `!` (exclamation mark) can be used to make exceptions
to exclusions.  The following is an example `.dockerignore` file that
uses this mechanism:

```gitignore
*.md
!README.md
```

All markdown files *except* `README.md` are excluded from the context.

The placement of `!` exception rules influences the behavior: the last
line of the `.dockerignore` that matches a particular file determines
whether it is included or excluded.  Consider the following example:

```gitignore
*.md
!README*.md
README-secret.md
```

No markdown files are included in the context except README files other than
`README-secret.md`.

Now consider this example:

```gitignore
*.md
README-secret.md
!README*.md
```

All of the README files are included.  The middle line has no effect because
`!README*.md` matches `README-secret.md` and comes last.

You can even use the `.dockerignore` file to exclude the `Dockerfile`
and `.dockerignore` files.  These files are still sent to the daemon
because it needs them to do its job.  But the `ADD` and `COPY` instructions
do not copy them to the image.

Finally, you may want to specify which files to include in the
context, rather than which to exclude. To achieve this, specify `*` as
the first pattern, followed by one or more `!` exception patterns.

> **Note**
>
> For historical reasons, the pattern `.` is ignored.

## Here-documents

> **Note**
>
> Added in [`docker/dockerfile:1.4`](#syntax)

Here-documents allow redirection of subsequent Dockerfile lines to the input of
`RUN` or `COPY` commands. If such command contains a [here-document](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_07_04)
the Dockerfile considers the next lines until the line only containing a
here-doc delimiter as part of the same command.

### Example: Running a multi-line script

```dockerfile
# syntax=docker/dockerfile:1
FROM debian
RUN <<EOT bash
  apt-get update
  apt-get install -y vim
EOT
```

If the command only contains a here-document, its contents is evaluated with
the default shell.

```dockerfile
# syntax=docker/dockerfile:1
FROM debian
RUN <<EOT
  mkdir -p foo/bar
EOT
```

Alternatively, shebang header can be used to define an interpreter.

```dockerfile
# syntax=docker/dockerfile:1
FROM python:3.6
RUN <<EOT
#!/usr/bin/env python
print("hello world")
EOT
```

More complex examples may use multiple here-documents.

```dockerfile
# syntax=docker/dockerfile:1
FROM alpine
RUN <<FILE1 cat > file1 && <<FILE2 cat > file2
I am
first
FILE1
I am
second
FILE2
```

### Example: Creating inline files

In `COPY` commands source parameters can be replaced with here-doc indicators.
Regular here-doc [variable expansion and tab stripping rules](https://pubs.opengroup.org/onlinepubs/9699919799/utilities/V3_chap02.html#tag_18_07_04)
apply.

```dockerfile
# syntax=docker/dockerfile:1
FROM alpine
ARG FOO=bar
COPY <<-EOT /app/foo
	hello ${FOO}
EOT
```

```dockerfile
# syntax=docker/dockerfile:1
FROM alpine
COPY <<-"EOT" /app/script.sh
	echo hello ${FOO}
EOT
RUN FOO=abc ash /app/script.sh
```

## Dockerfile examples

- The [Hello Build section](https://docs.docker.com/build/hellobuild/)
- The ["build images" section](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- The ["get started](https://docs.docker.com/get-started/)
- The [language-specific getting started guides](https://docs.docker.com/language/)
