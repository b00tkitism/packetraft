package api

import "net/http"

type API struct {
	client *http.Client
}

const BASE_URI = "https://packetraft.ir/api"

func NewAPI() *API {
	return &API{
		client: &http.Client{},
	}
}
