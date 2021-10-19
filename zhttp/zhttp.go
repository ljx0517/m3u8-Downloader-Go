package zhttp

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

type Zhttp struct {
	client *http.Client
}

func New(timeout time.Duration, proxy string) (*Zhttp, error) {
	zhttp := &Zhttp{
		client: http.DefaultClient,
	}

	if timeout > 0 {
		zhttp.client.Timeout = timeout
	}

	if proxy != "" {
		p, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}

		t := http.DefaultTransport.(*http.Transport).Clone()
		t.Proxy = func(*http.Request) (*url.URL, error) {
			return p, nil
		}
		zhttp.client.Transport = t
	}

	return zhttp, nil
}

func (zhttp *Zhttp) Get(url string, headers map[string]string, retry int) (int, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	var body []byte
	var code int

	var count int
	for count < retry {
		count++

		var resp *http.Response
		resp, err = zhttp.client.Do(req)
		if err != nil {
			wait := time.Duration(3 * count )
			log.Println("[-] fetch error:", url, "error:",  err, ",wait", wait, "secs")
			time.Sleep(time.Second * wait )
			continue
		}

		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			wait := time.Duration(3 * count )
			log.Println("[-] read error:", url, "error:",  err, ",wait", wait, "secs")
			time.Sleep(time.Second * wait )
			continue
		}

		code = resp.StatusCode

		break
	}

	if err != nil {
		return 0, nil, err
	}

	return code, body, nil
}
