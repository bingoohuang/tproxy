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
		local        = flag.StringArrayP("local", "p", []string{":33000"}, "Local ip:port to listen")
		width        = flag.IntP("width", "w", 32, "Number of bytes in each hex dump row (use 0 to turn off).")
		printStrings = flag.BoolP("strings", "S", true, "Print UTF-8 strings after hex dump")
		parent       = flag.StringArrayP("parent", "P", nil, "Parent address, such as: \"23.32.32.19:28008\"")
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
	saveSettings(*local, *parent, *delay, *protocol, *enableStats, *quiet)

	if len(settings.Parent) == 0 {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Parent target required"))
		flag.PrintDefaults()
		os.Exit(1)
	}

	if len(settings.Local) != len(settings.Parent) {
		fmt.Fprintln(os.Stderr, color.HiRedString("[x] Local/Parent mismatched"))
		flag.PrintDefaults()
		os.Exit(1)
	}

	var dumper func(data []byte) string
	if *width > 0 {
		dumper = hexdump.Config{Width: *width, PrintStrings: *printStrings}.Dump
	}

	startListener(dumper)
}
