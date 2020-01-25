# mustache-then-exec

Render mustache templates, then exec another binary

## Problem statement

You want a container with a minimal attack surface, and you need simple
templating within that container. You want the actual _main_ binary of your
container to be PID 1.

## Features

- Mustache templates
- Environment variables as source of data for your templates
- Glob support for specifying templates
- Chosing the rendered filename by search and replace of the template filename
- Executes the binary of your choice after rendering templates
- Errors out when rendering a template containing a reference to a non-existing
  environment variable (can be disabled with the `--allow-missing` flag)

## `mustache-then-exec --help`

```
Usage: mustache-then-exec [--allow-missing] [--template TEMPLATE] [--glob-template GLOB] BINARY [ARG [ARG ...]]

Positional arguments:
  BINARY                 the binary to run after rendering the templates
  ARG                    arguments to the binary to run after rendering the templates

Options:
  --allow-missing, -a    whether to allow missing variables (default: false)
  --template TEMPLATE, -t TEMPLATE
                         path to a template to be rendered
  --glob-template GLOB, -g GLOB
                         glob for templates to be rendered
  --help, -h             display this help and exit
```

## Example

Here's a Dockerfile for a simple image that executes a `cat` binary:

```dockerfile
FROM busybox

COPY mustache-then-exec /

RUN echo "Hello, {{{THING_TO_GREET}}}!" > /mytemplate

ENTRYPOINT [ "/mustache-then-exec", "-t", "/mytemplate", "--", "/bin/cat" ]
```

Which you can now run as follows (image name depending on how you tagged your
image during `docker build`):

```shellsession
$ #What happens when I don't set the required environment variable?
$ docker run --rm cat /mytemplate /mytemplate
Rendering template: /mytemplate; output: /mytemplate
Error: Missing variable "THING_TO_GREET"
$ #And what happens when I set it?
$ docker run --rm -e THING_TO_GREET=World cat /mytemplate
Rendering template: /mytemplate; output: /mytemplate
Hello, World!
$ #I'm goingo tm make cat read from stdin, so that it keeps running
$ docker run --rm -td -e THING_TO_GREET=World cat -
c6e4910c274d7b0984924c744d147935dbd62a89fe02da3a731a9264f794020c
$ #And now I can see what processes are running in my container: no mustache-then-exec!
$ docker exec c6e4910c274d7b0984924c744d147935dbd62a89fe02da3a731a9264f794020c pstree -p
cat(1)
$ docker kill c6e4910c274d7b0984924c744d147935dbd62a89fe02da3a731a9264f794020c 
c6e4910c274d7b0984924c744d147935dbd62a89fe02da3a731a9264f794020c
```

## Example with renaming and globbing

Let's expand a little on the previous example, and add two extra features:

1. Globbing
1. Renaming with regexp search&replace

Another Dockerfile that does mostly the same:

```dockerfile
FROM busybox

COPY mustache-then-exec /

RUN echo "Hello, {{{THING_TO_GREET}}}!" | tee /mytemplate01 /mytemplate02

ENTRYPOINT [ "/mustache-then-exec", "-g", "/mytemplate*:mytemplate([0-9]+):replacement-filename-$1", "--", "/bin/cat" ]
```

And, assuming we also named this image `cat`, let's use it as follows:

```shellsession
$ #What happened to the templates?
$ docker run --rm -e THING_TO_GREET=World cat /mytemplate01 /mytemplate02
Rendering template: /mytemplate01; output: /replacement-filename-01
Rendering template: /mytemplate02; output: /replacement-filename-02
Hello, {{{THING_TO_GREET}}}!
Hello, {{{THING_TO_GREET}}}!
$ #And did we indeed get renamed files?
$ docker run --rm -e THING_TO_GREET=World cat /replacement-filename-01 /replacement-filename-02
Rendering template: /mytemplate01; output: /replacement-filename-01
Rendering template: /mytemplate02; output: /replacement-filename-02
Hello, World!
Hello, World!
```
