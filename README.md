
## WeeProxy

[![Go Report Card](https://goreportcard.com/badge/github.com/abhishekkr/weeproxy)](https://goreportcard.com/report/github.com/abhishekkr/weeproxy)

WeeProxy is a wee bit http proxy to access http services mapped over URL path.

* HTTP Proxy based on URL path mapped to Backends

* round-robin load-balancing

* graceful stop/restart

* prometheus performance metrics at `/metrics`

* configurable header customization

* rate-limiting (same config for all backends)

* circuit breaker (same config for all backends)


#### Quikstart

* start server

```
dep ensure

go run weeproxy.go
```

* can also be used by downloaded pre-compiled binary from [latest release](https://github.com/abhishekkr/weeproxy/releases/latest), remember to have [sample config](./sample-config.json) in same dir or set required env var

* check http

```
curl localhost:8080/nothing

curl localhost:8080/google
```

#### Config

* config to be used can be changed by providing path to `env` var `WEEPROXY_CONFIG`

* proxies based on URL Path maps provided as configuration like, [sample config](./sample-config.json)

```
{
  "server": {
    "listen-at": ":8080"
  },
  "url-proxy": {
    "/google": "http://www.google.com http://www.google.in",
   ....more
  },
  "custom-headers": {
    "X-Proxy-By": "WeeProxy"        // custom response headers for each response
  },
  "sanity": {
    "max-errors-per-sec": "10",     // per backend max errors allowed before ban
    "max-request-per-sec": "7000"   // per backend max requests allowed before ban
   }
}
```

> in sample config above, when multiple backends need be load-balanced they are separated by a space
>
> if any other character gets used as separator, env `WEEPROXY_LB_SEPARATOR` need be set with same

* port to listen at can be modified updating `listen-at` field in above config

---

#### ToDo

* better no-backend handling

* runtime authenticated config updates over admin api

* rate-limiting (per backend config)

* circuit breaker (per backend config)

---
