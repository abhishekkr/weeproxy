/*
WeeProxy is a wee bit http proxy to access http services mapped over URL path.
*/

package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	golconfig "github.com/abhishekkr/gol/golconfig"
	golenv "github.com/abhishekkr/gol/golenv"
)

var (
	ConfigJson = golenv.OverrideIfEnv("WEEPROXY_CONFIG", "sample-config.json")

	ListenAt    string
	UrlProxyMap map[string]string
)

func getBackend(proxyConditionRaw string) string {
	return UrlProxyMap[proxyConditionRaw]
}

func logRequest(urlpath string, proxyUrl string) {
	log.Printf("proxy_condition: %s, proxy_url: %s\n", urlpath, proxyUrl)
}

func reverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

func handleProxy(res http.ResponseWriter, req *http.Request) {
	url := getBackend(req.URL.Path)

	if url != "" {
		logRequest(req.URL.Path, url)
		reverseProxy(url, res, req)
	}
}

func loadConfig() {
	config := make(map[string]map[string]string)
	configfile, err := ioutil.ReadFile(ConfigJson)
	if err != nil {
		log.Fatalf("cannot read config file %s; can override default config file path setting env var WEEPROXY_CONFIG", ConfigJson)
	}

	jsonCfg := golconfig.GetConfigurator("json")
	jsonCfg.Unmarshal(string(configfile), &config)

	ListenAt = config["server"]["listen-at"]
	UrlProxyMap = config["url-proxy"]
}

func main() {
	loadConfig()
	log.Printf("listening at: %s\n", ListenAt)
	for urlpath, proxy := range UrlProxyMap {
		log.Printf("+ %s => %s", urlpath, proxy)
	}

	http.HandleFunc("/", handleProxy)
	if err := http.ListenAndServe(ListenAt, nil); err != nil {
		panic(err)
	}
}
