package main

import (
	"fmt"
	"os"

	"github.com/bingoohuang/tproxy/hexdump"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

var settings Settings

func main() {
	var (
		localPort    = flag.IntP("port", "p", 0, "Local port to listen on, default to pick a random port")
		width        = flag.IntP("width", "w", 32, "Number of bytes in each hex dump row (use 0 to turn off).")
		printStrings = flag.BoolP("strings", "S", true, "Print UTF-8 strings after hex dump")
		localHost    = flag.StringP("listen", "l", "localhost", "Local address to listen on")
		remote       = flag.StringP("remote", "r", "", "Remote address (host:port) to connect")
		delay        = flag.DurationP("delay", "d", 0, "the delay to relay packets")
		protocol     = flag.StringP("type", "t", "", "The type of protocol, currently support http2, grpc, redis and mongodb")
		enableStats  = flag.BoolP("stat", "s", false, "Enable statistics")
		quiet        = flag.BoolP("quiet", "q", false,
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

	var dumper func(data []byte) string
	if *width > 0 {
		dumper = hexdump.Config{Width: *width, PrintStrings: *printStrings}.Dump
	}

	if err := startListener(dumper); err != nil {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Failed to start listener: %v", err))
		os.Exit(1)
	}
}
