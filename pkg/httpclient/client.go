package httpclient

import (
	"errors"
	"time"

	"github.com/projectdiscovery/useragent"
	"github.com/valyala/fasthttp"
)

var (
	ErrNilResponse    = errors.New("unexpected nil response")
	ErrNon200Response = errors.New("API responded with non-200 status code")
	ErrBadRequest     = errors.New("API responded with 400 status code")
)

type Header struct {
	Key   string
	Value string
}

func MakeRequest(c *fasthttp.Client, url string, maxRetries uint, timeout uint, headers ...Header) ([]byte, error) {
	var (
		req      *fasthttp.Request
		respBody []byte
		err      error
	)
	retries := int(maxRetries)
	for i := retries; i >= 0; i-- {
		req = fasthttp.AcquireRequest()

		req.Header.SetMethod(fasthttp.MethodGet)
		for _, header := range headers {
			if header.Key != "" {
				req.Header.Set(header.Key, header.Value)
			}
		}
		req.Header.Set(fasthttp.HeaderUserAgent, getUserAgent())
		req.Header.Set("Accept", "*/*")
		req.SetRequestURI(url)
		respBody, err = doReq(c, req, timeout)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

// doReq handles http requests
func doReq(c *fasthttp.Client, req *fasthttp.Request, timeout uint) ([]byte, error) {
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	defer fasthttp.ReleaseRequest(req)
	if err := c.DoTimeout(req, resp, time.Second*time.Duration(timeout)); err != nil {
		return nil, err
	}
	if resp.StatusCode() != 200 {
		if resp.StatusCode() == 400 {
			return nil, ErrBadRequest
		}
		return nil, ErrNon200Response
	}
	if resp.Body() == nil {
		return nil, ErrNilResponse
	}

	return resp.Body(), nil
}

func getUserAgent() string {
	pick := useragent.PickRandom().String()
	return pick
}
