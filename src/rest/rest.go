package rest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	HTTP_CLIENT    *http.Client
	ROBLOX_API_KEY string
)

func init() {
	HTTP_CLIENT = &http.Client{}

}

func robloxRequest(method string, path string, requestBody []byte) ([]byte, error) {
	fullUrl := fmt.Sprintf("https://apis.roblox.com/%s", path)
	if _, err := url.Parse(fullUrl); err != nil {
		return []byte{}, err
	}

	request, err := http.NewRequest(method, fullUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return []byte{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("x-api-key", ROBLOX_API_KEY)

	resp, err := HTTP_CLIENT.Do(request)
	if err != nil {
		return []byte{}, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode > 299 {
		return []byte{}, errors.New("non-200 status code: " + string(body))
	}

	defer resp.Body.Close()
	return body, nil
}

func RobloxGet(path string) ([]byte, error) {
	return robloxRequest("GET", path, []byte{})
}

func RobloxPost(path string, requestBody []byte) ([]byte, error) {
	return robloxRequest("POST", path, requestBody)
}
