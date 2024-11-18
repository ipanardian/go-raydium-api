package raydium

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/google/go-querystring/query"
)

type Raydium struct {
	BaseUrl string
	pool    sync.Pool
}

type RaydiumRequest struct {
	Data        *string
	QueryParams *string
	Headers     map[string]string
}

type RaydiumQuoteRequest struct {
	InputMint   string `json:"inputMint" url:"inputMint"`
	OutputMint  string `json:"outputMint" url:"outputMint"`
	Amount      int64  `json:"amount" url:"amount"`
	SlippageBps int    `json:"slippageBps" url:"slippageBps"`
	TxVersion   string `json:"txVersion" url:"txVersion"`
}

type RaydiumData struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Version string `json:"version"`
	Data    struct {
		SwapType             string  `json:"swapType"`
		InputMint            string  `json:"inputMint"`
		InputAmount          string  `json:"inputAmount"`
		OutputMint           string  `json:"outputMint"`
		OutputAmount         string  `json:"outputAmount"`
		OtherAmountThreshold string  `json:"otherAmountThreshold"`
		SlippageBps          int     `json:"slippageBps"`
		PriceImpactPct       float64 `json:"priceImpactPct"`
		ReferrerAmount       string  `json:"referrerAmount"`
		RoutePlan            []struct {
			PoolID            string        `json:"poolId"`
			InputMint         string        `json:"inputMint"`
			OutputMint        string        `json:"outputMint"`
			FeeMint           string        `json:"feeMint"`
			FeeRate           int           `json:"feeRate"`
			FeeAmount         string        `json:"feeAmount"`
			RemainingAccounts []interface{} `json:"remainingAccounts"`
			LastPoolPriceX64  string        `json:"lastPoolPriceX64,omitempty"`
		} `json:"routePlan"`
	} `json:"data"`
}

func NewRaydium(baseUrl string) *Raydium {
	return &Raydium{
		BaseUrl: baseUrl,
		pool: sync.Pool{
			New: func() interface{} {
				timeout := 5000 * time.Millisecond
				return httpclient.NewClient(
					httpclient.WithHTTPTimeout(timeout),
				)
			},
		},
	}
}

func (m *Raydium) Get(res any, path string, data RaydiumRequest) error {
	return m.getAndUnmarshalJson(res, path, data)
}

func (m *Raydium) Post(res any, path string, data RaydiumRequest) error {
	return m.postAndUnmarshalJson(res, path, data)
}

func (m *Raydium) SwapQuote(res any, headers map[string]string, data RaydiumQuoteRequest) error {
	qs, e := query.Values(data)
	if e != nil {
		return e
	}
	headers["User-Agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36"
	headers["Cache-Content"] = "no-cache"
	jsonStr := qs.Encode()
	return m.getAndUnmarshalJson(res, "/compute/swap-base-in", RaydiumRequest{
		QueryParams: &jsonStr,
		Headers:     headers,
	})
}

func (m *Raydium) getAndUnmarshalJson(res any, path string, data RaydiumRequest) error {
	client := m.pool.Get().(*httpclient.Client)
	url := fmt.Sprintf("%s%s", m.BaseUrl, path)

	if data.QueryParams != nil {
		url = url + "?" + *data.QueryParams
	}

	jB, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataReader := strings.NewReader("")
	if data.Data != nil {
		dataReader = strings.NewReader(string(jB))
	}

	req, err := http.NewRequest(http.MethodGet, url, dataReader)
	if err != nil {
		return err
	}

	if data.Headers != nil {
		for key, value := range data.Headers {
			req.Header.Set(key, value)
		}
	}

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	return nil
}

func (m *Raydium) postAndUnmarshalJson(res any, path string, data RaydiumRequest) error {
	client := m.pool.Get().(*httpclient.Client)
	url := fmt.Sprintf("%s%s", m.BaseUrl, path)

	if data.QueryParams != nil {
		url = url + "?" + *data.QueryParams
	}

	jB, err := json.Marshal(data)
	if err != nil {
		return err
	}

	dataReader := strings.NewReader("")
	if data.Data != nil {
		dataReader = strings.NewReader(string(jB))
	}

	req, err := http.NewRequest(http.MethodPost, url, dataReader)
	if err != nil {
		return err
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("Cache-Content", "no-cache")
	if data.Headers != nil {
		for key, value := range data.Headers {
			req.Header.Set(key, value)
		}
	}

	rcv := getPointer(res)

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if err := json.NewDecoder(rsp.Body).Decode(rcv); err != nil {
		return err
	}

	return nil
}

func getPointer(v interface{}) interface{} {
	vv := valueOf(v)
	if vv.Kind() == reflect.Ptr {
		return v
	}
	return reflect.New(vv.Type()).Interface()
}

func valueOf(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}
