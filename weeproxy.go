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

	revProxy "github.com/abhishekkr/weeproxy/revProxy"

	golconfig "github.com/abhishekkr/gol/golconfig"
	golconv "github.com/abhishekkr/gol/golconv"
	golenv "github.com/abhishekkr/gol/golenv"
	gollb "github.com/abhishekkr/gol/gollb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/pseidemann/finish"
)

var (
	configJSON  = golenv.OverrideIfEnv("WEEPROXY_CONFIG", "sample-config.json")
	lbSeparator = golenv.OverrideIfEnv("WEEPROXY_LB_SEPARATOR", " ")

	saneProxy     *revProxy.SaneProxy
	listenAt      string
	urlProxyMap   gollb.RoundRobin
	customHeaders map[string]string
)

func reverseProxy(target string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse(target)

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Transport = saneProxy

	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = url.Host

	proxy.ServeHTTP(res, req)
}

func handleProxy(res http.ResponseWriter, req *http.Request) {
	url := urlProxyMap.GetBackend(req.URL.Path)
	if saneProxy.Banned(url) || url == "" {
		log.Println("banned url:", url)
		return
	}
	log.Printf("proxy/condition: %s, proxy/url: %s\n", req.URL.Path, url)
	for headerKey, headerVal := range customHeaders {
		req.Header[headerKey] = []string{headerVal}
	}
	reverseProxy(url, res, req)
}

func loadConfig() {
	config := make(map[string]map[string]string)
	configfile, err := ioutil.ReadFile(configJSON)
	if err != nil {
		log.Fatalf("cannot read config file %s; can override default config file path setting env var WEEPROXY_CONFIG", configJSON)
	}

	jsonCfg := golconfig.GetConfigurator("json")
	jsonCfg.Unmarshal(string(configfile), &config)

	for urlpath, proxy := range config["url-proxy"] {
		log.Printf("+ %s => %s", urlpath, strings.Split(proxy, lbSeparator))
	}

	listenAt = config["server"]["listen-at"]
	urlProxyMap.LoadWithSeparator(config["url-proxy"], lbSeparator)
	customHeaders = config["custom-headers"]

	maxReq := golconv.StringToUint64(config["sanity"]["max-request-per-sec"],
		uint64(7000))
	maxErr := golconv.StringToUint64(config["sanity"]["max-errors-per-sec"],
		uint64(100))
	saneProxy = revProxy.NewSaneProxy(maxReq, maxErr, config["url-proxy"],
		lbSeparator)
}

func main() {
	loadConfig()
	log.Printf("listening at: %s\n", listenAt)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", handleProxy)

	srv := &http.Server{Addr: listenAt}

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
