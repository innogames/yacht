package healthcheck

import (
	"context"
	"fmt"
	"github.com/innogames/yacht/logger"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HCHttp stores all properties of a HTTP or HTTPS healthcheck.
type HCHttp struct {
	HCBase
	host    string
	url     string
	okCodes []int
}

type httpCodeError struct {
	badCode int
}

func (e *httpCodeError) Error() string {
	return fmt.Sprintf("Bad HTTP status code %d", e.badCode)
}

// NewHCHttp creates new HTTP or HTTPs healthcheck struct and populates it with data from Json config
func newHCHttp(logPrefix string, json JSONMap) (*HCHttp, *HCBase) {
	hc := new(HCHttp)
	hc.hcType = json["type"].(string)

	if host, ok := json["host"].(string); ok {
		hc.host = host
	}
	if url, ok := json["url"].(string); ok {
		hc.url = url
	}
	if codes, ok := json["ok_codes"].(string); ok {
		for _, code := range strings.Split(codes, ",") {
			code, _ := strconv.Atoi(code)
			hc.okCodes = append(hc.okCodes, code)
		}
	}

	hc.logPrefix = logPrefix + fmt.Sprintf("healthcheck: %s url: %s", hc.hcType, hc.url)

	if len(hc.okCodes) == 0 {
		logger.Info.Printf(hc.logPrefix+" unable to parse ok codes from %s", json["ok_codes"].([]string))
		hc.okCodes = []int{200}
	}

	logger.Info.Printf(hc.logPrefix + " created")
	return hc, &hc.HCBase
}

// check performs the healthckeck. It is called from the main goroutine of HealthcheckBase.
func (hc *HCHttp) do(hcr chan (ResultError)) context.CancelFunc {

	// Prepare context for canceling of requests
	ctx, cancel := context.WithCancel(context.Background())

	// Disable handling of HTTP redirects. We want to get 3xx code and parse it.
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Millisecond * time.Duration(hc.HCBase.timeout),
	}

	// Build HTTP request
	req, err := http.NewRequest("HEAD", hc.hcType+"://"+hc.HCBase.ipAddress+hc.url, nil)
	if err != nil {
		logger.Error.Printf(hc.logPrefix + err.Error())
		return nil
	}
	// Override host header if a custom one is configured.
	if len(hc.host) > 0 {
		req.Host = hc.host
	}

	// Wrap context around request
	req = req.WithContext(ctx)

	// Spawn HTTP request in another goroutine.
	go func() {
		res := HCError              // Start with default return code: error of healthcheck.
		resp, err := client.Do(req) // Launch the request.
		select {
		case <-ctx.Done():
			// Cancelled or timed out.
			res = HCBad
		default:
			// Normal exit.
			if resp != nil {
				for _, okCode := range hc.okCodes {
					if resp.StatusCode == okCode {
						res = HCGood
						break
					}
				}
				if res != HCGood {
					err = &httpCodeError{resp.StatusCode}
				}
			}
		}
		hcr <- ResultError{
			res: res,
			err: err,
		}
	}()

	return cancel
}

// Run starts operation of this healthcheck, in fact it calls the Base class.
func (hc *HCHttp) Run(wg *sync.WaitGroup) {
	hc.HCBase.run(wg, hc.do)
}

// Stop terminates this healthcheck, in fact it calls the Base class.
func (hc *HCHttp) Stop() {
	hc.HCBase.Stop()
}
