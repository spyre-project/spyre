package main

import (
	"github.com/daviddengcn/go-colortext"

	"github.com/spyre-project/spyre"
	"github.com/spyre-project/spyre/config"

	"fmt"
)

var logo = [...]string{
	`  ___ ___  __ _________ `,
	` (_-</ _ \/ // / __/ -_)`,
	`/___/ .__/\_, /_/  \__/ `,
	`   /_/   /___/          `,
}

const spacer = "     "

func displayLogo() {
	ct.Foreground(ct.Blue, true)
	fmt.Println(logo[0])
	fmt.Print(logo[1])

	ct.ResetColor()
	fmt.Print(spacer)
	fmt.Print("version")
	ct.Foreground(ct.Green, false)
	fmt.Printf(" %s\n", spyre.Version)

	ct.Foreground(ct.Magenta, true)
	fmt.Print(logo[2])
	if config.Global.RulesetMarker != "" {
		ct.ResetColor()
		fmt.Print(spacer)
		fmt.Print("ruleset")
		ct.Foreground(ct.Yellow, false)
		fmt.Printf(" %s\n", config.Global.RulesetMarker)
	} else {
		fmt.Println()
	}
	ct.Foreground(ct.Blue, true)
	fmt.Print(logo[3])
	ct.ResetColor()
	fmt.Println(spacer + "https://github.com/spyre-project/spyre")
	fmt.Println()
}
