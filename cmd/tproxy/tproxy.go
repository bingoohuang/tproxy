package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bingoohuang/tproxy/hexdump"
	"github.com/fatih/color"
)

var settings Settings

func main() {
	var (
		localPort    = flag.Int("p", 0, "Local port to listen on, default to pick a random port")
		width        = flag.Int("width", 32, "Number of bytes in each hex dump row.")
		printStrings = flag.Bool("strings", true, "Print UTF-8 strings after hex dump")
		localHost    = flag.String("l", "localhost", "Local address to listen on")
		remote       = flag.String("r", "", "Remote address (host:port) to connect")
		delay        = flag.Duration("d", 0, "the delay to relay packets")
		protocol     = flag.String("t", "", "The type of protocol, currently support http2, grpc, redis and mongodb")
		enableStats  = flag.Bool("s", false, "Enable statistics")
		quiet        = flag.Bool("q", false,
			"Quiet mode, only prints connection open/close and stats, default false")
	)

	if len(os.Args) <= 1 {
		flag.Usage()
		return
	}

	flag.Parse()
	saveSettings(*localHost, *localPort, *remote, *delay, *protocol, *enableStats, *quiet)

	if len(settings.Remote) == 0 {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Remote target required"))
		flag.PrintDefaults()
		os.Exit(1)
	}

	dumper := hexdump.Config{Width: *width, PrintStrings: *printStrings}
	if err := startListener(dumper.Dump); err != nil {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Failed to start listener: %v", err))
		os.Exit(1)
	}
}