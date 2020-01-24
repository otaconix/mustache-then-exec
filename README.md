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

Which you can now run as follows (image name depending on how you tagged your
image during `docker build`):

```shellsession
$ docker run --rm -p 80:80 mustache-then-exec-nginx-proxy
Filling template: /etc/nginx/fastcgi.conf
Filling template: /etc/nginx/nginx.conf
Error: Missing variable "NGINX_PROXY_UPSTREAM"
$ docker run --rm -d -p 80:80 -e NGINX_PROXY_UPSTREAM=https://httpbin.org/get mustache-then-exec-nginx-proxy
f3a67fa793821c3815997cfcc78a6619a01836f3c61380606922c6bf779ab30d
$ docker exec f3a67fa793821c3815997cfcc78a6619a01836f3c61380606922c6bf779ab30d pstree -p
nginx(1)---nginx(11)
$ curl http://localhost:80/get
curl http://localhost:80
{
  "args": {}, 
  "headers": {
    "Accept": "*/*", 
    "Host": "httpbin.org", 
    "User-Agent": "curl/7.68.0",
  }, 
  "origin": "1.1.1.1", 
  "url": "https://httpbin.org/get"
}
$ docker kill f3a67fa793821c3815997cfcc78a6619a01836f3c61380606922c6bf779ab30d
```
