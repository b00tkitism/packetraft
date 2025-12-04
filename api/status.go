package api

import (
	"encoding/json"
	"strings"
)

type DownloadLink struct {
	Platform  string `json:"platform"`
	URL       string `json:"url"`
	Signature string `json:"signature"`
}

type GameServer struct {
	Region        string   `json:"region"`
	Servers       []string `json:"servers"`
	DefaultServer *string  `json:"default_server"`
	PingableIP    string
}

type ProgramAntiSanctionPassPort struct {
	Start    uint16 `json:"start"`
	End      uint16 `json:"end"`
	Protocol string `json:"protocol"`
}

type ProgramPassPort struct {
	Start uint16 `json:"start"`
	End   uint16 `json:"end"`
}

type Game struct {
	Name                         string                         `json:"name"`
	Revision                     *int                           `json:"revision"`
	Order                        *int                           `json:"order"`
	GameServers                  *[]GameServer                  `json:"game_servers"`
	Alerts                       *[]string                      `json:"alerts"`
	ConnectionTestIP             *string                        `json:"connection_test_ip"`
	SupportsLinux                *bool                          `json:"supports_linux"`
	IsProgram                    *bool                          `json:"is_program"`
	Programs                     *[]string                      `json:"programs"`
	ProgramDefaultRoutes         *string                        `json:"program_default_routes"`
	ProgramPassPorts             *[]ProgramPassPort             `json:"program_pass_ports"`
	ProgramAntiSanctionPassPorts *[]ProgramAntiSanctionPassPort `json:"program_anti_sanction_pass_ports"`
	EnableAntiSanctionByDefault  *bool                          `json:"enable_anti_sanction_by_default"`
}

type Chain struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

type Server struct {
	Name     string `json:"name"`
	Revision int    `json:"revision"`
	Tag      string `json:"tag"`
	NoChain  bool   `json:"no_chain"`
}

type Status struct {
	LatestClientVersion int            `json:"latest_client_version"`
	DownloadLinks       []DownloadLink `json:"download_links"`
	Games               []Game         `json:"games"`
	Chains              []Chain        `json:"chains"`
	Servers             []Server       `json:"servers"`
}

func (api *API) Status() (*Status, error) {
	resp, err := api.client.Get(BASE_URI + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result Status
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (status *Status) CorrespondingServers(gameName string) []Server {
	servers := []Server{}

	if gameName == "" {
		return servers
	}

	if strings.ToLower(gameName) == "all" {
		return status.Servers
	}

	for _, game := range status.Games {
		if game.Name == gameName {
			for _, server := range status.Servers {
				if game.GameServers != nil {
					for _, gameServer := range *game.GameServers {
						for _, gameServerServer := range gameServer.Servers {
							gameServerServer = strings.ReplaceAll(gameServerServer, "#", "")
							if server.Tag == gameServerServer {
								servers = append(servers, server)
							}
						}
					}
				} else {
					return servers
				}
			}
		}
	}

	return servers
}
