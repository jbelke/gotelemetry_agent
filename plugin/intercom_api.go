package plugin

import (
	"encoding/json"
	"net/http"
	"net/url"
)

func (p *IntercomPlugin) performRequestAndGetHeaders(endpoint string, target interface{}) (http.Header, error) {
	p.Limiter.Wait()

	finalUrl := ""

	u, err := url.Parse(endpoint)

	if !u.IsAbs() {
		finalUrl = "https://api.intercom.io/" + endpoint
	} else {
		finalUrl = endpoint
	}

	req, err := http.NewRequest("GET", finalUrl, nil)

	if err != nil {
		return http.Header{}, err
	}

	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(p.AppID, p.APIKey)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return http.Header{}, err
	}

	if err = NewHTTPError(res); err != nil {
		println(err == nil)
		return http.Header{}, err
	}

	if target != nil {
		err = json.NewDecoder(res.Body).Decode(target)
	}

	return res.Header, err
}

func (p *IntercomPlugin) performRequest(endpoint string, target interface{}) error {
	_, err := p.performRequestAndGetHeaders(endpoint, target)

	return err
}
