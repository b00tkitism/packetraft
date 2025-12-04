package api

import (
	"bytes"
	"io"
	"strings"

	"github.com/itsabgr/ge"
)

func generatePayload(gameName, serverName string) string {
	gameName = strings.ToLower(gameName)

	if serverName == "" {
		return gameName
	}
	return gameName + "#" + strings.ToLower(serverName)
}

func (api *API) GenerateConfig(gameName, serverName string) (string, error) {
	if gameName == "" {
		return "", ge.New("game cannot be nil")
	}

	resp, err := api.client.Post(BASE_URI+"/generate_config", "application/x-www-form-urlencoded", bytes.NewReader([]byte(generatePayload(gameName, serverName))))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return "", err
	}

	return string(body), nil
}
