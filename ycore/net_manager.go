package ycore

import (
	"context"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
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
	TY_PONG
	TY_DISCONNECT
)
const (
	discoveryPort = "9999"
	discoveryMsg  = "MOCK_SERVER_DISCOVER"
	replyMsg      = "MOCK_SERVER_FOUND"
	broadcastIP   = "255.255.255.255"
)

type DiscoveryResult struct {
	IP   string
	Port int
	Err  error
}

type NetPackage struct {
	TY     uint32
	Sender xid.ID
	Data   []byte
}

// a different idea would have the server have a general in queue
// and the client all have individual out queues
type NetClient struct {
	Out chan NetPackage
	In  chan NetPackage
}

type NetManager struct {
	NetType     uint32
	Ip          string
	Port        string
	Address     string
	BlockCount  int16
	LastSent    int16
	UseDiscover bool
	Ctx         context.Context
	ctxCancel   context.CancelFunc
	mu          sync.Mutex
	Clients     map[xid.ID]*NetClient
}

type ProtoFunc func(cli *NetClient, r *NetPackage) error

func Discover(timeout time.Duration) DiscoveryResult {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return DiscoveryResult{Err: err}
	}
	defer conn.Close()
	udpConn := conn.(*net.UDPConn)
	if err := udpConn.SetWriteBuffer(1024); err != nil {
		return DiscoveryResult{Err: err}
	}

	broadcastAddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(broadcastIP, discoveryPort))
	if err != nil {
		return DiscoveryResult{Err: err}
	}
	_, err = udpConn.WriteTo([]byte(discoveryMsg), broadcastAddr)
	if err != nil {
		return DiscoveryResult{Err: err}
	}
	if err := udpConn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return DiscoveryResult{Err: err}
	}
	buf := make([]byte, 514)
	for {
		n, addr, err := udpConn.ReadFrom(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return DiscoveryResult{
					Err: err,
				}
			}
			return DiscoveryResult{Err: err}
		}
		reply := string(buf[:n])
		ss := strings.Split(reply, ":")
		if ss[0] == replyMsg {
			udpAddr := addr.(*net.UDPAddr)
			port, err := strconv.Atoi(ss[1])
			if err != nil {
				return DiscoveryResult{
					Err: err,
				}
			}
			return DiscoveryResult{
				IP:   udpAddr.IP.String(),
				Port: port,
			}
		}
	}
}

func RecieveDiscover(ctx context.Context, port string) {
	pc, err := net.ListenPacket("udp", net.JoinHostPort("", "0"))
	if err != nil {
		log.Println(err)
		return
	}
	defer pc.Close()
	buffer := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n0, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				log.Println(err)
				return
			}
			if string(buffer[:n0]) == discoveryMsg {
				reply := replyMsg + ":" + port
				pc.WriteTo([]byte(reply), addr)
			}
		}
	}
}

func NewNetManager(address, ip, port string) *NetManager {
	c, cancel := context.WithCancel(context.Background())
	return &NetManager{
		Ip:        ip,
		Port:      port,
		Address:   address,
		Ctx:       c,
		ctxCancel: cancel,
		Clients:   make(map[xid.ID]*NetClient),
	}
}

func (n *NetManager) StartServer() {
	if n.UseDiscover {
		go RecieveDiscover(n.Ctx, n.Port)
	}
	go func() {
		ln, err := net.Listen("tcp", net.JoinHostPort(n.Ip, n.Port))
		if err != nil {
			panic(err)
		}
		defer ln.Close()
		for {
			select {
			case <-n.Ctx.Done():
				return
			default:
				conn, err := ln.Accept()
				if err != nil {
					panic(err)
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
		}
	}()
}

func (n *NetManager) CloseClient(id xid.ID) {
	cli, ok := n.Clients[id]
	if ok {
		close(cli.Out)
		delete(n.Clients, id)
	}
}

func (n *NetManager) StopServer() {
	n.ctxCancel()
	clear(n.Clients)
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
		TY:     uint32(TY_NEW_CLIENT),
		Sender: xid.NilID(), //server is nil id
		Data:   id.Bytes(),
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
			if np.TY == uint32(TY_PING) || np.TY == uint32(TY_PONG) {
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

func (n *NetManager) CreateClient() (*NetClient, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(n.Address, n.Port))
	if err != nil {
		return nil, err
	}
	out := make(chan NetPackage, 64)
	in := make(chan NetPackage, 64)
	go RunClient(n.Ctx, conn, out, in)
	return &NetClient{
		In:  in,
		Out: out,
	}, nil
}

func RunClient(ctx context.Context, c net.Conn, out chan NetPackage, in chan<- NetPackage) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	//watcher
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

	//reader
	var wg sync.WaitGroup
	wg.Go(func() {
		defer close(in)
		defer cancel()
		for {
			if err := c.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
				log.Printf("client: set read deadline: %v", err)
				return
			}
			var np NetPackage
			if err := decoder.Decode(&np); err != nil {
				if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
					log.Printf("read: %v", err)
				}
				return
			}
			if np.TY == uint32(TY_PING) {
				out <- NetPackage{
					TY: uint32(TY_PONG),
				}
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

	//writer
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
		}
	}
}

func (n *NetManager) SendToClients(np NetPackage) {
	for _, cli := range n.Clients {
		cli.Out <- np
	}
}

func (n *NetManager) GetClientPackages() []NetPackage {
	nps := make([]NetPackage, 0)
	for id, cli := range n.Clients {
		select {
		case np, ok := <-cli.In:
			if !ok {
				//cli has disconneted
				close(cli.Out)
				delete(n.Clients, id)
			}
			nps = append(nps, np)
		default:
			continue
		}
	}
	return nps
}
