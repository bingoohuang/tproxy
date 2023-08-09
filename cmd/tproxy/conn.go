package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

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

func (c *PairedConnection) copyData(dst io.Writer, src io.Reader, tag string) {
	_, e := io.Copy(dst, src)
	if e != nil && !errors.Is(e, io.EOF) {
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
	tee := io.MultiWriter(c.svrConn, w)
	interop := protocol.CreateInterop(settings.Protocol, c.hexDumper, &c.printLock)
	go interop.Dump(r, protocol.ClientSide, c.id, settings.Quiet)
	c.copyData(tee, c.cliConn, protocol.ClientSide)
}

func (c *PairedConnection) handleServerMessage() {
	// server closed also trigger client close.
	defer c.stop()

	r, w := io.Pipe()
	tee := io.MultiWriter(newDelayedWriter(c.cliConn, settings.Delay, c.stopChan), w)
	interop := protocol.CreateInterop(settings.Protocol, c.hexDumper, &c.printLock)
	go interop.Dump(r, protocol.ServerSide, c.id, settings.Quiet)
	c.copyData(tee, c.svrConn, protocol.ServerSide)
}

func (c *PairedConnection) process() {
	defer c.stop()

	conn, err := net.Dial("tcp", settings.Remote)
	if err != nil {
		display.PrintlnWithTime(color.HiRedString("[x][%d] Couldn't connect to server: %v", c.id, err))
		return
	}

	display.PrintlnWithTime(color.HiGreenString("[%d] Connected to server: %s", c.id, conn.RemoteAddr()))

	stat.AddConn(strconv.Itoa(c.id), conn.(*net.TCPConn))
	c.svrConn = conn
	go c.handleServerMessage()

	c.handleClientMessage()
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

func startListener(hexDumper protocol.HexDumper) error {
	stat = NewStater(NewConnCounter(), NewStatPrinter(statInterval))
	go stat.Start()

	conn, err := net.Listen("tcp", fmt.Sprintf("%s:%d", settings.LocalHost, settings.LocalPort))
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
		go pconn.process()
	}
}
