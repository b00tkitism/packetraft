package tui

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	dnsforwarder "github.com/b00tkitism/packetraft/dns"
	"github.com/b00tkitism/packetraft/forwarder"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sync/errgroup"
)

// load status/games
func (m Model) loadStatus() tea.Cmd {
	return func() tea.Msg {
		status, err := m.apiClient.Status()
		if err != nil {
			return errMsg{err}
		}
		return statusLoadedMsg{status: *status}
	}
}

// load servers for selected game
func (m Model) loadServers(game string) tea.Cmd {
	return func() tea.Msg {
		servers := m.status.CorrespondingServers(game)
		names := make([]string, len(servers))
		for i, s := range servers {
			names[i] = s.Name
		}
		return serversLoadedMsg{serverNames: names}
	}
}

// generate config, write tmp file, prepare replaced config and remote addr
func (m Model) genConfig(game, server string) tea.Cmd {
	return func() tea.Msg {
		m.localPort = strconv.Itoa(int(math.Abs(float64(rand.UintN(65000)) - 65000)))

		cfg, err := m.apiClient.GenerateConfig(game, server)
		if err != nil {
			return errMsg{err}
		}

		// parse remote endpoint from config
		parts := strings.Split(cfg, "Endpoint = ")
		if len(parts) < 2 {
			return errMsg{fmt.Errorf("failed to parse Endpoint from config")}
		}
		remoteLine := parts[1]
		remoteAddr := strings.ReplaceAll(remoteLine, "/udp", "")
		remoteAddr = strings.TrimSpace(remoteAddr)
		remoteAddr = strings.Split(remoteAddr, "\n")[0]

		replaced := strings.ReplaceAll(cfg, remoteAddr, "127.0.0.1:"+m.localPort)
		replaced = strings.ReplaceAll(replaced, "/udp", "")
		replaced = strings.ReplaceAll(replaced, "AllowedIPs", "AllowedIPs = 0.0.0.0/1, 128.0.0.0/1#")
		replaced = strings.ReplaceAll(replaced, "DNS = ", "DNS = 127.0.0.1#")

		fileName := "PR-" + game
		if server != "" {
			fileName += "-" + server
		}
		fileName += "-" + "*.conf" // temp suffix for CreateTemp

		tmpf, err := os.CreateTemp("", fileName)
		if err != nil {
			return errMsg{err}
		}
		if _, err := tmpf.WriteString(replaced); err != nil {
			tmpf.Close()
			return errMsg{err}
		}
		if err := tmpf.Close(); err != nil {
			return errMsg{err}
		}

		abs, _ := filepath.Abs(tmpf.Name())

		if runtime.GOOS == "windows" {
			wgFolder := `C:\Program Files\WireGuard\Data\Configurations`
			if err := os.MkdirAll(wgFolder, 0o755); err != nil {
				goto ret
			}

			dst := filepath.Join(wgFolder, filepath.Base(tmpf.Name()))
			if err := os.Rename(abs, dst); err != nil {
				input, err := os.ReadFile(abs)
				if err != nil {
					goto ret
				}
				if err := os.WriteFile(dst, input, 0o644); err != nil {
					goto ret
				}
				os.Remove(abs)
			}
			abs = dst
		}

	ret:
		return configReadyMsg{
			listenPort:   m.localPort,
			config:       cfg,
			remoteAddr:   remoteAddr,
			replacedConf: replaced,
			tmpPath:      abs,
		}
	}
}

// start forwarder local and remote loops
func (m Model) startForwarder(remote string) tea.Cmd {
	return func() tea.Msg {
		fwd, err := forwarder.NewForwarder(":"+m.localPort, remote)
		if err != nil {
			return errMsg{err}
		}

		dnsfwd := dnsforwarder.New("1.1.1.1:853", "cloudflare-dns.com", 3*time.Second)

		eg, egCtx := errgroup.WithContext(m.ctx)
		eg.Go(func() error { return fwd.StartLocal(egCtx) })
		eg.Go(func() error { return fwd.StartRemote(egCtx) })
		eg.Go(func() error { return dnsfwd.Start(egCtx) })

		go func() {
			<-m.ctx.Done()
			_ = fwd.Close()
		}()

		go func() {
			if err := eg.Wait(); err != nil {
				log.Printf("Forwarder error: %v", err)
			}
		}()

		return forwarderStartedMsg{}
	}
}
