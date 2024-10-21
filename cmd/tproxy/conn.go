package main

import (
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/bingoohuang/ngg/goup/shapeio"
	"github.com/bingoohuang/tproxy/display"
	"github.com/bingoohuang/tproxy/protocol"
	"github.com/fatih/color"
)

const (
	useOfClosedConn = "use of closed network connection"
	statInterval    = time.Second * 5
)

var (
	errClientCanceled = errors.New("client canceled")
	stat              Stater
)

type PairedConnection struct {
	cliConn   net.Conn
	svrConn   net.Conn
	stopChan  chan struct{}
	hexDumper protocol.HexDumper
	id        int
	once      sync.Once

	printLock sync.Mutex
}

func NewPairedConnection(id int, cliConn net.Conn, hexDumper protocol.HexDumper) *PairedConnection {
	return &PairedConnection{
		id:        id,
		cliConn:   cliConn,
		stopChan:  make(chan struct{}),
		hexDumper: hexDumper,
	}
}

type IdleConn struct {
	net.Conn
	idleTimeout time.Duration
}

func (c *IdleConn) Read(b []byte) (int, error) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.idleTimeout)); err != nil {
		log.Printf("SetReadDeadline error: %v", err)
	}

	n, err := c.Conn.Read(b)
	if err != nil {
		var e net.Error
		if errors.As(err, &e) && e.Timeout() { // 如果是超时错误或者连接关闭错误，则退出
			log.Printf("Connection idle timeout %s, closing %s", c.idleTimeout, c.Conn.RemoteAddr())
			err = io.EOF
		}
	}

	return n, err

}

func (c *PairedConnection) copyData(dst io.Writer, srcConn net.Conn, tag string, limit float64, idleTimeout time.Duration) {
	if idleTimeout > 0 {
		srcConn = &IdleConn{Conn: srcConn, idleTimeout: idleTimeout}
	}

	var src io.Reader = srcConn
	if limit > 0 {
		src = shapeio.NewReader(src, shapeio.WithRateLimit(limit))
	}

	if _, e := io.Copy(dst, src); e != nil && !errors.Is(e, io.EOF) {
		var netOpError *net.OpError
		if errors.As(e, &netOpError) && netOpError.Err.Error() != useOfClosedConn {
			reason := netOpError.Unwrap().Error()
			display.PrintlnWithTime(color.HiRedString("[%d] %s error, %s", c.id, tag, reason))
		}
	}
}

func (c *PairedConnection) handleClientMessage() {
	// client closed also trigger server close.
	defer c.stop()

	r, w := io.Pipe()
	defer r.Close()

	tee := io.MultiWriter(c.svrConn, w)
	interop := protocol.CreateInterop(settings.Protocol, c.hexDumper, &c.printLock)
	go interop.Dump(r, protocol.ClientSide, c.id, settings.Quiet)
	c.copyData(tee, c.cliConn, protocol.ClientSide, settings.UpLimit, settings.IdleTimeout)
}

func (c *PairedConnection) handleServerMessage() {
	// server closed also trigger client close.
	defer c.stop()

	r, w := io.Pipe()
	defer r.Close()

	tee := io.MultiWriter(newDelayedWriter(c.cliConn, settings.Delay, c.stopChan), w)
	interop := protocol.CreateInterop(settings.Protocol, c.hexDumper, &c.printLock)
	go interop.Dump(r, protocol.ServerSide, c.id, settings.Quiet)
	c.copyData(tee, c.svrConn, protocol.ServerSide, settings.DownLimit, settings.IdleTimeout)
}

func expandAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Panicf("invalid %s, should [host]:port", addr)
	}

	if host == "" {
		host = "127.0.0.1"
	}
	if l := port[0]; 'a' <= l && l <= 'z' || 'A' <= l && l <= 'Z' {
		h := sha1.New()
		h.Write([]byte(port))
		sum := h.Sum(nil)
		portNum := binary.BigEndian.Uint16(sum[:2])
		port = fmt.Sprintf("%d", portNum)
	}

	return host + ":" + port
}

func (c *PairedConnection) process(parent, target string) {
	parentExpand := expandAddr(parent)
	conn, err := net.Dial("tcp", parentExpand)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("[x][%d] Couldn't connect to server: %v", c.id, err))
		return
	}

	display.PrintlnWithTime(color.HiGreenString("[%d] Connected to server: %s", c.id, conn.RemoteAddr()))

	if target != "" {
		targetExpand := expandAddr(target)
		_, _ = conn.Write([]byte("TARGET " + targetExpand + ";"))
	}

	stat.AddConn(strconv.Itoa(c.id), conn.(*net.TCPConn))
	c.svrConn = conn

	go c.handleServerMessage()
	go c.handleClientMessage()
}

func (c *PairedConnection) stop() {
	c.once.Do(func() {
		close(c.stopChan)
		stat.DelConn(strconv.Itoa(c.id))

		if c.cliConn != nil {
			display.PrintlnWithTime(color.HiBlueString("[%d] Client connection closed", c.id))
			c.cliConn.Close()
		}
		if c.svrConn != nil {
			display.PrintlnWithTime(color.HiBlueString("[%d] Server connection closed", c.id))
			c.svrConn.Close()
		}
	})
}

func startListener(hexDumper protocol.HexDumper) {
	stat = NewStater(NewConnCounter(), NewStatPrinter(statInterval))
	go stat.Start()

	var wg sync.WaitGroup
	for i, local := range settings.Local {
		wg.Add(1)

		target := ""
		if len(settings.Target) > i {
			target = settings.Target[i]
		}

		go func(local, parent, target string) {
			defer wg.Done()

			if err := startListenerSingle(hexDumper, local, parent, target); err != nil {
				fmt.Fprintln(os.Stderr, color.HiRedString("[x] Failed to start listener: %v", err))
			}
		}(local, settings.Parent[i], target)

	}

	wg.Wait()
}

func startListenerSingle(hexDumper protocol.HexDumper, local, parent, target string) error {
	if _, ok := ParsePort((local)); ok {
		local = "127.0.0.1:" + local
	}
	if _, ok := ParsePort((parent)); ok {
		parent = "127.0.0.1:" + parent
	}

	conn, err := net.Listen("tcp", local)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer conn.Close()

	display.PrintfWithTime("Listening on %s...\n", conn.Addr().String())

	var connIndex int
	for {
		cliConn, err := conn.Accept()
		if err != nil {
			return fmt.Errorf("server: accept: %w", err)
		}

		connIndex++
		display.PrintlnWithTime(color.HiGreenString("[%d] Accepted from: %s",
			connIndex, cliConn.RemoteAddr()))

		pconn := NewPairedConnection(connIndex, cliConn, hexDumper)
		go pconn.process(parent, target)
	}
}

func ParsePort(portStr string) (int, bool) {
	// 将字符串转换为整数
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, false
	}

	// 检查端口是否在有效范围内
	if port >= 0 && port <= 65535 {
		return port, true
	}
	return 0, false
}
