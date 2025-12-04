package forwarder

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"
)

var KEY = []byte("khoshtip")

type Forwarder struct {
	listener   *net.UDPConn
	remote     *net.UDPConn
	lastClient *net.UDPAddr
	mu         sync.RWMutex
}

func NewForwarder(listenAddr, remoteAddr string) (*Forwarder, error) {
	fmt.Println(listenAddr, remoteAddr)

	laddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return nil, err
	}

	remote, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, err
	}

	return &Forwarder{
		listener: listener,
		remote:   remote,
	}, nil
}

func (f *Forwarder) Close() error {
	if f.listener != nil {
		_ = f.listener.Close()
	}
	if f.remote != nil {
		_ = f.remote.Close()
	}
	return nil
}

func (f *Forwarder) setLastClient(addr *net.UDPAddr) {
	f.mu.Lock()
	f.lastClient = addr
	f.mu.Unlock()
}

func (f *Forwarder) getLastClient() *net.UDPAddr {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lastClient
}

func (forwarder *Forwarder) StartLocal(ctx context.Context) error {
	buf := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, addr, err := forwarder.listener.ReadFromUDP(buf)
			if err != nil {
				return err
			}

			forwarder.setLastClient(addr)

			_, err = forwarder.remote.Write(xorEncrypt(buf[:n]))
			if err != nil {
				return err
			}
		}
	}
}

func (forwarder *Forwarder) StartRemote(ctx context.Context) error {
	buf := make([]byte, 1500)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := forwarder.remote.Read(buf)
			if err != nil {
				return err
			}

			client := forwarder.getLastClient()
			if client == nil {
				continue
			}

			_, err = forwarder.listener.WriteToUDP(xorEncrypt(buf[:n]), client)
			if err != nil {
				return err
			}
		}
	}
}

func xorEncrypt(plain []byte) []byte {
	result := bytes.Clone(plain)
	for idx := range result {
		result[idx] ^= KEY[idx%len(KEY)]
	}

	return result
}
