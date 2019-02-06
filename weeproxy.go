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
	"strings"

	golconfig "github.com/abhishekkr/gol/golconfig"
	golenv "github.com/abhishekkr/gol/golenv"
	gollb "github.com/abhishekkr/gol/gollb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pseidemann/finish"
)

var (
	ConfigJson  = golenv.OverrideIfEnv("WEEPROXY_CONFIG", "sample-config.json")
	LbSeparator = golenv.OverrideIfEnv("WEEPROXY_LB_SEPARATOR", " ")

	ListenAt      string
	UrlProxyMap   gollb.RoundRobin
	CustomHeaders map[string]string
)

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
	url := UrlProxyMap.GetBackend(req.URL.Path)

	if url != "" {
		log.Printf("proxy/condition: %s, proxy/url: %s\n", req.URL.Path, url)
		for headerKey, headerVal := range CustomHeaders {
			req.Header[headerKey] = []string{headerVal}
		}
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

	for urlpath, proxy := range config["url-proxy"] {
		log.Printf("+ %s => %s", urlpath, strings.Split(proxy, LbSeparator))
	}

	ListenAt = config["server"]["listen-at"]
	UrlProxyMap.LoadWithSeparator(config["url-proxy"], LbSeparator)
	CustomHeaders = config["custom-headers"]
}

func main() {
	loadConfig()
	log.Printf("listening at: %s\n", ListenAt)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", handleProxy)

	srv := &http.Server{Addr: ListenAt}

	fin := finish.New()
	fin.Add(srv)

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	fin.Wait()
}
