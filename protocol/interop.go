package protocol

import (
	"fmt"
	"io"
	"sync"

	"github.com/bingoohuang/tproxy/display"
)

const (
	ServerSide = "SERVER"
	ClientSide = "CLIENT"

	bufferSize     = 1 << 20
	grpcProtocol   = "grpc"
	http2Protocol  = "http2"
	httpProtocol   = "http"
	redisProtocol  = "redis"
	mysqlProtocol  = "mysql"
	oracleProtocol = "oracle"
	mongoProtocol  = "mongo"
	mqttProtocol   = "mqtt"
)

type (
	HexDumper func(data []byte) string
	Interop   interface {
		Dump(r io.Reader, source string, id int, quiet bool)
	}
)

func CreateInterop(protocol string, hexDumper HexDumper, printLock *sync.Mutex) Interop {
	switch protocol {
	case grpcProtocol:
		return &http2Interop{
			explainer: new(grpcExplainer),
			hexDumper: hexDumper,
		}
	case httpProtocol:
		return newHttpInterop()
	case http2Protocol:
		return new(http2Interop)
		
	case mysqlProtocol:
		return new(mysqlInterop)
	case oracleProtocol:
		return new(oracleInterop)
	case redisProtocol:
		return new(redisInterop)
	case mongoProtocol:
		return new(mongoInterop)
	case mqttProtocol:
		return new(mqttInterop)
	default:
		return &defaultInterop{
			hexDumper: hexDumper,
			printLock: printLock,
		}
	}
}

type defaultInterop struct {
	hexDumper HexDumper
	printLock *sync.Mutex
}

func (d defaultInterop) Dump(r io.Reader, source string, id int, quiet bool) {
	data := make([]byte, bufferSize)
	for {
		n, err := r.Read(data)
		if n > 0 && !quiet {
			d.print(source, id, data, n)
		}
		if err != nil && err != io.EOF {
			fmt.Printf("unable to read data %v", err)
			break
		}
		if n == 0 {
			break
		}
	}
}

func (d defaultInterop) print(source string, id int, data []byte, n int) {
	if d.hexDumper == nil {
		return
	}

	d.printLock.Lock()
	defer d.printLock.Unlock()

	display.PrintfWithTime("from %s [%d]:\n", source, id)
	fmt.Println(d.hexDumper(data[:n]))
}
