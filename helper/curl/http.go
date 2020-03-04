package curl

import (
	"errors"
	"github.com/derekyu332/goii/helper/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

func newClient() *http.Client {
	//TBD pool
	return &http.Client{
		Timeout: 10 * time.Second,
	}
}

func Get(url string) (string, error) {
	respond, err := newClient().Get(url)

	if err != nil {
		return "", err
	}

	defer respond.Body.Close()
	body, err := ioutil.ReadAll(respond.Body)
	logger.Info("HttpGetRequest Err Status = %d body = %v", respond.StatusCode, string(body))

	if err != nil {
		logger.Warning("Err %v", err.Error())
		return "", err
	} else if respond.StatusCode != 200 {
		logger.Warning("Status Code = %d", respond.StatusCode)
		return "", errors.New("Bad Status Code")
	}

	return string(body), nil
}

func PostForm(postUrl string, params map[string]string) (string, error) {
	data := url.Values{}

	for key, value := range params {
		data.Add(key, value)
	}

	respond, err := newClient().PostForm(postUrl, data)

	if err != nil {
		return "", err
	}

	defer respond.Body.Close()
	body, err := ioutil.ReadAll(respond.Body)

	if err != nil || respond.StatusCode != 200 {
		logger.Warning("Status Code = %d", respond.StatusCode)
		return "", errors.New("Bad Status Code")
	}

	return string(body), nil
}
