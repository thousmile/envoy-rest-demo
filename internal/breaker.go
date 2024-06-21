package internal

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"log"
	"net/url"
	"strings"
)

func Breaker(apply func() (*resty.Response, error)) error {
	return BreakerWithError(
		apply,
		func(response *resty.Response, err error) error {
			return err
		},
	)
}

func BreakerWithError(apply func() (*resty.Response, error), fallback func(*resty.Response, error) error) error {
	response, err := apply()
	if err != nil {
		return fallback(response, err)
	} else {
		if !response.IsSuccess() {
			header := response.Header()
			log.Println("-------------- error response headers: --------------")
			for k, v := range header {
				log.Printf("%s : %s\n", k, v)
			}
			if _, ok := header[ErrorReason]; ok {
				errMsg, _ := url.QueryUnescape(header.Get(ErrorReason))
				return fallback(response, errors.New(errMsg))
			}
			ct := header.Get(ContentType)
			if strings.Contains(ct, TextPlain) {
				return fallback(response, errors.New(response.String()))
			}
			if strings.Contains(ct, ApplicationJson) {
				var errMap map[string]interface{}
				_ = json.Unmarshal(response.Body(), &errMap)
				if e1, ok := errMap["error"]; ok {
					return fallback(response, errors.New(e1.(string)))
				}
				if e2, ok := errMap["message"]; ok {
					return fallback(response, errors.New(e2.(string)))
				}
			}
			return fallback(response, errors.New("unknown error"))
		} else {
			return nil
		}
	}
}
