package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bingoohuang/godaemon"
	"github.com/bingoohuang/tproxy/hexdump"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

var settings Settings

// -- int Value
type intValue struct {
	value int
	set   bool
}

func newIntValue(val int) *intValue { return &intValue{value: val, set: false} }

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	i.value = int(v)
	i.set = true
	return err
}

func (i *intValue) Type() string   { return "int" }
func (i *intValue) String() string { return strconv.Itoa(i.value) }

func main() {
	width := newIntValue(32)

	flag.VarP(width, "width", "w", "Number of bytes in each hex dump row (use 0 to turn off).")
	local := flag.StringArrayP("local", "p", []string{":33000"}, "Local ip:port to listen")
	printStrings := flag.BoolP("strings", "S", true, "Print UTF-8 strings after hex dump")
	parent := flag.StringArrayP("parent", "P", nil, `Parent address, such as: "23.32.32.19:28008"`)
	target := flag.StringArrayP("target", "T", nil, `Target address, such as: "23.32.32.19:28008"，配合 frpc 使用`)
	delay := flag.DurationP("delay", "d", 0, "the delay to relay packets")
	protocol := flag.StringP("type", "t", "", "The type of protocol, currently support http2, grpc, redis and mongodb")
	enableStats := flag.BoolP("stat", "s", false, "Enable statistics")
	daemon := flag.BoolP("daemon", "D", false, "Daemonize")
	quiet := flag.BoolP("quiet", "q", false,
		"Quiet mode, only prints connection open/close and stats, default false")

	flag.Parse()

	if !width.set && (len(*target) > 0 || len(*local) > 1) {
		width.value = 0
	}

	saveSettings(*local, *parent, *target, *delay, *protocol, *enableStats, *quiet)

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

	godaemon.Daemonize(godaemon.WithDaemon(*daemon))

	var dumper func(data []byte) string
	if width.value > 0 {
		dumper = hexdump.Config{Width: width.value, PrintStrings: *printStrings}.Dump
	}

	startListener(dumper)
}
