package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Call ...
func Call(method, url string, in, out interface{}) ([]byte, error) {
	return CallWithOptions(Options{
		Method:          method,
		URL:             url,
		RequestPayload:  in,
		ResponsePayload: out,
		ContentType:     "application/json",
	})
}

// CallWithToken ...
func CallWithToken(method, url, token string, in, out interface{}) ([]byte, error) {
	return CallWithOptions(Options{
		Method:          method,
		URL:             url,
		Token:           token,
		RequestPayload:  in,
		ResponsePayload: out,
		ContentType:     "application/json",
	})
}

// Options ...
type Options struct {
	Method          string
	URL             string
	Token           string
	RequestPayload  interface{}
	ResponsePayload interface{}
	ContentType     string
	Accept          string
}

// CallWithOptions ...
func CallWithOptions(o Options) ([]byte, error) {
	zap.S().Infof("calling url: %s %s  --  token: %s", strings.ToUpper(o.Method), o.URL, o.Token)

	client := &http.Client{}

	var payload io.Reader
	if o.RequestPayload != nil {
		b, err := json.Marshal(o.RequestPayload)
		if err != nil {
			return nil, errors.Wrapf(err, "could not marshal JSON payload: %#v", o.RequestPayload)
		}

		zap.S().Debugf("payload: %s", string(b))

		payload = bytes.NewReader(b)
	}

	req, err := http.NewRequest(o.Method, o.URL, payload)

	if o.ContentType != "" {
		req.Header.Add("Content-Type", o.ContentType)
	}
	if o.Accept != "" {
		req.Header.Add("Accept", o.Accept)
	}
	if o.Token != "" {
		req.Header.Add("Authorization", "Bearer "+o.Token)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "could not call URL %s %s", req.Method, req.URL.String())
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read response body: %s", string(data))
	}

	dataPreview := data
	if len(data) > 1000 {
		dataPreview = data[0:1000]
	}
	zap.S().Debugf("got response: %s\n", string(dataPreview))

	if o.ResponsePayload != nil {
		err = json.Unmarshal(data, o.ResponsePayload)
		if err != nil {
			err = errors.Wrapf(err, "could not unmarshal JSON response: %s", string(data))
			zap.S().Error(err)
			return nil, err
		}
	}

	return data, nil
}
