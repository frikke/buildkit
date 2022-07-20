---
title: MAINTAINER command
keywords: build, dockerfile, reference
---

> **Warning:**
>
> MAINTAINER instruction is deprecated
{:.warning}

```dockerfile
MAINTAINER <name>
```

The `MAINTAINER` instruction sets the *Author* field of the generated images.
The `LABEL` instruction is a much more flexible version of this and you should use
it instead, as it enables setting any metadata you require, and can be viewed
easily, for example with `docker inspect`. To set a label corresponding to the
`MAINTAINER` field you could use:

```dockerfile
LABEL org.opencontainers.image.authors="SvenDowideit@home.org.au"
```

This will then be visible from `docker inspect` with the other labels.
