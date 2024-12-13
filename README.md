# UrlLinker Backend Server

An url shortener backend server built with Golang and Fiber v3 framework. With Cloudflare and Cloudflare Turnstile supported.

## Environmental variables

All the environmental variables are optional.\
`IP` Address for the server to run on.\
`PORT` Port number for the server to run on.\
`CLOUDFLARE_TURNSTILE_SECRET_KEY` Secret Key of/for Cloudflare Turnstile captcha verification. \
`CORS_DOMAINS` Domain/s to be included in CORS list. Example: `example.com, www.exmaple.com, sub.example.com, *.example.com`\
`TLS_CERT` TLS/SSL Certificate\
`TLS_KEY` TLS/SSL Certificate Private/Secret Key

Exmaple:

```env
IP=
PORT=
CLOUDFLARE_TURNSTILE_SECRET_KEY=
CORS_DOMAINS=
TLS_CERT=
TLS_KEY=
```
