package ddp

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Proxy proxies requests
type Proxy struct {
	path    string
	target  Target
	limiter *RateLimiter
	logger  *Logger
	queue   interface{} // mechanism to retry old, rated requests before proxying new ones
}

// NewProxy returns a new Proxy
func NewProxy(path string, logger *Logger) *Proxy {
	return &Proxy{
		target:  http.DefaultClient,
		path:    path,
		limiter: NewRateLimiter(),
		logger:  logger,
	}
}

// ServeHTTP implements http.Handler
// destination path is set on querystring
func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	p.logger.Append(req)

	if req.URL.Path != p.path {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	q := req.URL.Query().Get("q")
	if q == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dst, err := url.Parse(q)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check for ratelimit
	buc := p.limiter.UncleDrew(dst.Path)
	if buc.Remaining > 0 {
		// forward req to dst
		res, err := p.target.Do(req)
		buc.Remaining--
		if err != nil {
			// separate log for invalid requests
			p.logger.Append(err)
			return
		}
		defer res.Body.Close()

		p.logger.Append(res)
		b, _ := ioutil.ReadAll(res.Body)
		w.Write(b)
		return
	}
	// else lock until rate reset
	// and put to queue
	wait := p.limiter.GetWaitTime(buc, 1)
	w.Write([]byte(fmt.Sprintf("rate-limited, wait for %d", wait)))
}
