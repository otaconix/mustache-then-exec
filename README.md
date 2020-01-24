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
- Executes the binary of your choice after rendering templates
- Errors out when rendering a template containing a reference to a non-existing
  environment variable (can be disabled with the `--allow-missing` flag)

## `mustache-then-exec --help`

```
  Replaces TEMPLATE files (which are Go globs) with their contents
  rendered as a Mustache template, where the environment variables are
  passed as data to those tempales.

  -allow-missing
    	Whether to allow missing variables (default: false).
```

## Example

Here's a Dockerfile for a simple nginx proxy:

```dockerfile
FROM nginx:alpine

COPY mustache-then-exec /

RUN echo $'\
  events { \n\
  } \n\
  http { \n\
    server { \n\
      listen 80; \n\
      location / { \n\
        proxy_pass {{{NGINX_PROXY_UPSTREAM}}}; \n\
      } \n\
    } \n\
  }' > /etc/nginx/nginx.conf 

ENTRYPOINT [ "/mustache-then-exec", "/etc/nginx/*.conf", "--", "/usr/sbin/nginx" ]

CMD [ "-g", "daemon off;" ]
```

Which you can now run as follows (depending on how you tagged your image during
`docker build`):

```shell-session
$ docker run --rm -d --name nginx-proxy -e NGINX_PROXY_UPSTREAM=https://httpbin.org mustache-then-exec-nginx-proxy
$ docker exec nginx-proxy pstree -p
nginx(1)---nginx(11)
$ docker kill nginx-proxy
```
