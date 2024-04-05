package client

import (
	"fmt"
	"net/http"
)

type AuthDoer struct {
	APIToken string
	Version  string
	Scope    string
}

func (ad *AuthDoer) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Dagster-Cloud-Api-Token", ad.APIToken)
	req.Header.Set("Dagster-Cloud-Version", ad.Version)
	req.Header.Set("Dagster-Cloud-Scope", ad.Scope)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("url=%q error=%s", req.URL.String(), err.Error())
	}

	return resp, nil
}
