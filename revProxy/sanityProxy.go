package revProxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

/*SaneProxy implements RoundTrip for ReverseProxy to manage
 * RateLimiting and CircuitBreaking
 */
type SaneProxy struct {
	MaxReqPerSec uint64
	MaxErrPerSec uint64

	BackendHostMap    map[string]string
	BackendsReqPerSec map[string]uint64
	BackendsErrPerSec map[string]uint64
	BackendsBan       map[string]bool
}

/*NewSaneProxy returns properly configured SaneProxy
 * RateLimiting and CircuitBreaking
 */
func NewSaneProxy(maxReq uint64, maxErr uint64,
	proxyMap map[string]string, lbSeparator string) *SaneProxy {

	sp := SaneProxy{}
	sp.MaxReqPerSec = maxReq
	sp.MaxErrPerSec = maxErr

	backendCount := 0
	for _, backends := range proxyMap {
		backendCount += len(strings.Split(backends, lbSeparator))
	}

	sp.BackendHostMap = make(map[string]string, backendCount)
	sp.BackendsReqPerSec = make(map[string]uint64, backendCount)
	sp.BackendsErrPerSec = make(map[string]uint64, backendCount)
	sp.BackendsBan = make(map[string]bool, backendCount)
	for _, backends := range proxyMap {
		sp.initBackends(backends, lbSeparator)
	}
	go sp.perSecSanity()
	return &sp
}

func (sp *SaneProxy) initBackends(backends, lbSeparator string) {
	for _, backend := range strings.Split(backends, lbSeparator) {
		host := backend
		u, err := url.Parse(host)
		if err == nil {
			host = u.Host
		}
		sp.BackendHostMap[backend] = host
		sp.BackendsReqPerSec[host] = 0
		sp.BackendsErrPerSec[host] = 0
		sp.BackendsBan[host] = false
	}
}

/*Banned returns boolean for host string as being banned for request or not*/
func (sp *SaneProxy) Banned(backend string) bool {
	return sp.BackendsBan[sp.BackendHostMap[backend]]
}

/*RoundTrip implements Transport layer handler for response
 * managing rate-limiting and circuit-breaking.
 */
func (sp *SaneProxy) RoundTrip(request *http.Request) (*http.Response, error) {
	var err error

	sp.checkRate(request)

	if request.Body != nil {
		err = sp.checkRequest(request)
	}
	if err != nil {
		sp.errorHandler(request)
		return nil, err
	}

	response, err := sp.checkResponse(request)
	if err != nil {
		sp.errorHandler(request)
		return nil, err
	}

	return response, err
}

func (sp *SaneProxy) errorHandler(req *http.Request) {
	sp.BackendsErrPerSec[req.Host]++
}

func (sp *SaneProxy) checkRate(req *http.Request) error {
	sp.BackendsReqPerSec[req.Host]++
	return nil
}

func (sp *SaneProxy) isFaulty(host string) bool {
	if sp.BackendsReqPerSec[host] > sp.MaxReqPerSec ||
		sp.BackendsErrPerSec[host] > sp.MaxErrPerSec {
		return true
	}
	return false
}

func (sp *SaneProxy) checkRequest(request *http.Request) error {
	buf, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}
	if len(buf) == 0 {
		rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
		request.Body = rdr2
	}
	return err
}

func (sp *SaneProxy) checkResponse(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 500 {
		sp.errorHandler(request)
	}

	_, err = httputil.DumpResponse(response, true)
	if err != nil {
		return nil, err
	}
	return response, err
}

func (sp *SaneProxy) perSecSanity() {
	for {
		go sp.perSecSanityOp()
		time.Sleep(time.Second * 1)
	}
}

func (sp *SaneProxy) perSecSanityOp() {
	for host := range sp.BackendsBan {
		sp.BackendsBan[host] = sp.isFaulty(host)
	}
	for host, reqCount := range sp.BackendsReqPerSec {
		sp.BackendsReqPerSec[host] = sp.adjustReqPerSec(
			reqCount, sp.MaxReqPerSec)
	}
	for host, errCount := range sp.BackendsErrPerSec {
		sp.BackendsErrPerSec[host] = sp.adjustErrPerSec(
			errCount, sp.MaxErrPerSec)
	}
}

func (sp *SaneProxy) adjustReqPerSec(req, maxReq uint64) uint64 {
	if req <= maxReq {
		return 0
	}
	return req - maxReq
}

func (sp *SaneProxy) adjustErrPerSec(errCount, maxErr uint64) uint64 {
	if errCount <= maxErr {
		return 0
	}
	return errCount - maxErr
}
