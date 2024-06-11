package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/bingoohuang/gg/pkg/man"
	"github.com/bingoohuang/godaemon"
	"github.com/bingoohuang/tproxy/hexdump"
	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
)

type Settings struct {
	Local     []string
	Parent    []string
	Target    []string
	Protocol  string
	LocalPort int
	Delay     time.Duration
	Stat      bool
	Quiet     bool
	UpLimit   float64
	DownLimit float64
}

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
	flag.StringArrayVarP(&settings.Local, "local", "p", []string{":33000"}, "Local ip:port to listen")
	printStrings := flag.BoolP("strings", "S", false, "Print UTF-8 strings after hex dump")
	flag.StringArrayVarP(&settings.Parent, "parent", "P", nil, `Parent address, such as: "23.32.32.19:28008"`)
	flag.StringArrayVarP(&settings.Target, "target", "T", nil, `Target address, such as: "23.32.32.19:28008"，配合 frpc 使用`)
	flag.DurationVarP(&settings.Delay, "delay", "d", 0, "the delay to relay packets")
	flag.StringVarP(&settings.Protocol, "type", "t", "", "The type of protocol, currently support http, http2, grpc, redis and mongodb")
	flag.BoolVarP(&settings.Stat, "stat", "s", false, "Enable statistics")
	daemon := flag.BoolP("daemon", "D", false, "Daemonize")
	flag.BoolVarP(&settings.Quiet, "quiet", "q", false, "Quiet mode, only prints connection open/close and stats, default false")
	flag.Var(NewRateLimitFlag(&settings.UpLimit), "up", "Upward speed limit per second, like 1K")
	flag.Var(NewRateLimitFlag(&settings.DownLimit), "down", "Downward speed limit per second, like 1K")
	flag.Parse()

	if !width.set && (len(settings.Target) > 0 || len(settings.Local) > 1) {
		width.value = 0
	}

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

func NewRateLimitFlag(val *float64) *RateLimitFlag {
	return &RateLimitFlag{Val: val}
}

type RateLimitFlag struct {
	Val *float64
}

func (i *RateLimitFlag) Type() string { return "rateLimit" }

func (i *RateLimitFlag) String() string {
	s := man.Bytes(uint64(*i.Val))
	return s
}

func (i *RateLimitFlag) Set(value string) (err error) {
	val, err := man.ParseBytes(value)
	if err != nil {
		return err
	}

	*i.Val = float64(val)
	return nil
}

func (i *RateLimitFlag) Float64() float64 {
	return *i.Val
}
