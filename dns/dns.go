package dnsforwarder

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miekg/dns"
)

type DNSForwarder struct {
	timeout      time.Duration
	server       *dns.Server
	upstreamAddr string
	upstreamName string
}

func New(dotAddr, dotName string, timeout time.Duration) *DNSForwarder {
	forwarder := &DNSForwarder{
		timeout:      timeout,
		upstreamAddr: dotAddr,
		upstreamName: dotName,
	}

	forwarder.server = &dns.Server{
		Addr:    "127.0.0.1:53",
		Net:     "udp",
		Handler: dns.HandlerFunc(forwarder.handleDNSRequest),
	}

	return forwarder
}

func (forwarder *DNSForwarder) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		errCh <- forwarder.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return forwarder.server.ShutdownContext(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (forwarder *DNSForwarder) handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	resp, err := forwarder.resolve(r)
	if err != nil {
		m := new(dns.Msg)
		m.SetRcode(r, dns.RcodeServerFailure)
		_ = w.WriteMsg(m)
		return
	}

	resp.Id = r.Id

	if err := w.WriteMsg(resp); err != nil {
		log.Printf("write msg error: %v", err)
	}
}

func (forwarder *DNSForwarder) resolve(req *dns.Msg) (*dns.Msg, error) {
	wire, err := req.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack request: %w", err)
	}

	dialer := &net.Dialer{Timeout: forwarder.timeout}
	tlsCfg := &tls.Config{
		ServerName:         forwarder.upstreamName,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", forwarder.upstreamAddr, tlsCfg)
	if err != nil {
		return nil, fmt.Errorf("tls dial %s: %w", forwarder.upstreamAddr, err)
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(forwarder.timeout))

	var lenBuf [2]byte
	binary.BigEndian.PutUint16(lenBuf[:], uint16(len(wire)))
	if _, err := conn.Write(lenBuf[:]); err != nil {
		return nil, fmt.Errorf("write length: %w", err)
	}
	if _, err := conn.Write(wire); err != nil {
		return nil, fmt.Errorf("write body: %w", err)
	}

	if _, err := conn.Read(lenBuf[:]); err != nil {
		return nil, fmt.Errorf("read length: %w", err)
	}
	respLen := binary.BigEndian.Uint16(lenBuf[:])
	if respLen == 0 {
		return nil, fmt.Errorf("upstream returned zero length")
	}

	respBuf := make([]byte, respLen)
	readOff := 0
	for readOff < int(respLen) {
		n, err := conn.Read(respBuf[readOff:])
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		readOff += n
	}

	resp := new(dns.Msg)
	if err := resp.Unpack(respBuf); err != nil {
		return nil, fmt.Errorf("failed to unpack response: %w", err)
	}
	return resp, nil
}
