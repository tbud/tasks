package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	. "github.com/tbud/bud/context"
)

type RunProxyTask struct {
	ProxyAddress  string
	ServerAddress string

	proxy *httputil.ReverseProxy
}

func init() {
	proxy := &RunProxyTask{
		ProxyAddress:  "http://qianmi-resources.oss-cn-hangzhou-internal.aliyuncs.com",
		ServerAddress: "0.0.0.0:80",
	}

	Task("run", Group("proxy"), proxy, Usage("Use to start a proxy."))
}

func (p *RunProxyTask) Execute() (err error) {
	return http.ListenAndServe(p.ServerAddress, p)
}

func (p *RunProxyTask) Validate() (err error) {
	var serverUrl *url.URL
	if serverUrl, err = url.ParseRequestURI(p.ProxyAddress); err != nil {
		return err
	}

	p.proxy = httputil.NewSingleHostReverseProxy(serverUrl)
	return nil
}

func (p *RunProxyTask) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/favicon.ico" {
		return
	}

	p.proxy.ServeHTTP(rw, req)
}
