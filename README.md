# UrlLinker Backend Server

An url shortener backend server built with Golang and Fiber v3 framework. With Cloudflare and Cloudflare Turnstile supported.

## Environmental variables

All the environmental variables are optional.

1. `CLOUDFLARE_TURNSTILE_SECRET_KEY` Secret Key of/for Cloudflare Turnstile captcha verification.
2. `CORS_DOMAINS` Domain/s to be included in CORS list. Example: `https://example.com, https://www.exmaple.com, https://sub.example.com, https://*.example.com`.
3. `TLS_CERT` TLS/SSL Certificate.
4. `TLS_KEY` TLS/SSL Certificate Private/Secret Key.

Example:

```env
CLOUDFLARE_TURNSTILE_SECRET_KEY=
CORS_DOMAINS=
TLS_CERT=
TLS_KEY=
```

## Available Methods

### GET

1. `/{hash}` To get the original URL.

### POST

1. `/` To submit/save URL to the server and return a hash.\
   Require `url` in query.\
   And `captcha_token` if Cloudflare Turnstile is enabled.
