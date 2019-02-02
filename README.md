
## WeeProxy

WeeProxy is a wee bit http proxy to access http services mapped over URL path.

#### Quikstart

* start server

```
dep ensure

go run weeproxy.go
```

* check http

```
curl localhost:8080/nothing

curl localhost:8080/google
```

---
