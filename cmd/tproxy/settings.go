package main

import "time"

type Settings struct {
	Parent    []string
	Local     []string
	Protocol  string
	LocalPort int
	Delay     time.Duration
	Stat      bool
	Quiet     bool
}

func saveSettings(local, parent []string, delay time.Duration,
	protocol string, stat, quiet bool,
) {
	settings.Local = local
	settings.Parent = parent
	settings.Delay = delay
	settings.Protocol = protocol
	settings.Stat = stat
	settings.Quiet = quiet
}
