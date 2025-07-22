package api

import (
	"net/http"
	"time"

	http_client "github.com/lllllan02/http-client"
)

type Client struct {
	*http_client.Client
}

func NewClient(cookie string) (*Client, error) {
	headers := map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6",
		"Cache-Control":             "max-age=0",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"User-Agent":                "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36 Edg/138.0.0.0",
	}

	handler := func(*http.Response) error {
		return nil
	}

	retryStrategy := http_client.DefaultRetryStrategy.
		WithMaxRetries(3).
		WithInterval(2 * time.Second).
		WithRetryCondition(func(resp *http.Response, err error) bool {
			//  如果返回 403，则重试
			if err != nil || resp.StatusCode == http.StatusForbidden {
				return true
			}

			return false
		})

	client := http_client.New(
		http_client.WithCookie("https://tiku.scratchor.com", cookie),
		http_client.WithHeaders(headers),
		http_client.WithLimiter(5),
		http_client.WithResponseHandler(handler),
		http_client.WithRetryStrategy(retryStrategy),
	)

	return &Client{Client: client}, nil
}

func (client *Client) Get(url string) (string, error) {
	return client.Client.Get(url)
}
