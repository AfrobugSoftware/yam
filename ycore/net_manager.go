package ycore

import (
	"context"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/rs/xid"
)

const (
	MAXBUFFERSIZE = 2048
)

const (
	SERVER = iota
	CLIENT
)

const (
	TY_INIT int16 = iota
	TY_CLIENT_ACCEPT
	TY_NEW_CLIENT
	TY_PING
)

type NetPackage struct {
	TY   uint32
	Data []byte
}

type NetClient struct {
	Out chan NetPackage
	In  chan NetPackage
}

type NetManager struct {
	NetType    uint32
	Ip         string
	Port       string
	Address    string
	Client     net.Conn
	BlockCount int16
	LastSent   int16
	Ctx        context.Context
	mu         sync.Mutex
	Clients    map[xid.ID]*NetClient
}

// I am not sure
type StateFunc func(w io.Writer, r io.Reader) error

func NewNetManager(address, ip, port string) *NetManager {
	return &NetManager{
		Ip:      ip,
		Port:    port,
		Address: address,
		Clients: make(map[xid.ID]*NetClient),
	}
}

func (n *NetManager) StartServer() {
	go func() {
		ln, err := net.Listen("tcp", net.JoinHostPort(n.Ip, n.Port))
		if err != nil {

		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Panicf("server crashed: %v", err)
			}
			out := make(chan NetPackage, 64)
			in := make(chan NetPackage, 64)
			id := xid.New()
			n.mu.Lock()
			n.Clients[id] = &NetClient{
				Out: out,
				In:  in,
			}
			n.mu.Unlock()

			go ProcessClient(n.Ctx, id, conn, out, in)
		}
	}()
}

func ProcessClient(ctx context.Context, id xid.ID, c net.Conn, out <-chan NetPackage, in chan<- NetPackage) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		<-ctx.Done()
		c.Close()
	}()

	encoder := gob.NewEncoder(c)
	decoder := gob.NewDecoder(c)

	encode := func(np NetPackage) error {
		if err := c.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return err
		}
		return encoder.Encode(np)
	}
	np := NetPackage{
		TY:   uint32(TY_NEW_CLIENT),
		Data: id.Bytes(),
	}
	err := encoder.Encode(np)
	if err != nil {
		log.Println(err)
		close(in)
		return
	}

	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(in)
		defer cancel()
		for {
			if err := c.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				log.Printf("client %s: set read deadline: %v", id, err)
				return
			}
			var np NetPackage
			if err := decoder.Decode(&np); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					log.Printf("client %s: read: %v", id, err)
				}
				return
			}
			if np.TY == uint32(TY_PING) {
				continue
			}
			select {
			case in <- np:
			case <-ctx.Done():
				return
			}
		}
	})
	defer func() {
		cancel()
		wg.Wait()
	}()
	ping := time.NewTicker(10 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case np, ok := <-out:
			if !ok {
				return
			}
			if err := encode(np); err != nil {
				log.Printf("write: %v", err)
				return
			}
			ping.Reset(10 * time.Second)
		case <-ping.C:
			if err := encode(NetPackage{TY: uint32(TY_PING)}); err != nil {
				log.Printf("ping: %v", err)
				return
			}
		}
	}
}

func (n *NetManager) CreateClient() error {
	conn, err := net.Dial("tcp", net.JoinHostPort(n.Address, n.Port))
	if err != nil {
		return err
	}
	n.Client = conn
	return nil
}
