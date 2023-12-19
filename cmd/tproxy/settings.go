package main

import "time"

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

func saveSettings(local, parent, target []string, delay time.Duration,
	protocol string, stat, quiet bool, upLimit, downLimit float64,
) {
	settings.Local = local
	settings.Parent = parent
	settings.Target = target
	settings.Delay = delay
	settings.Protocol = protocol
	settings.Stat = stat
	settings.Quiet = quiet
	settings.UpLimit = upLimit
	settings.DownLimit = downLimit
}
