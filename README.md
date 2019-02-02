
## WeeProxy

WeeProxy is a wee bit http proxy to access http services mapped over URL path.

#### Quikstart

* start server

```
dep ensure

go run weeproxy.go
```

* can also be used by downloaded pre-compiled binary from [latest release](https://github.com/abhishekkr/weeproxy/releases/tag/v0.1.0), remember to have [sample config](./sample-config.json) in same dir or set required env var

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
    "/google": "http://www.google.com",
   ....more
  }
}
```

* port to listen at can be modified updating `listen-at` field in above config

---

#### ToDo

* graceful stop/restart

* performance metrics

* configurable header customization

* better no-backend handling

* rate-limiting

* circuit breaker

* runtime authenticated config updates over admin api

---
