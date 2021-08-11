package main

import (
	"github.com/daviddengcn/go-colortext"

	"github.com/spyre-project/spyre"

	"fmt"
)

var logo = [...]string{
	`  ___ ___  __ _________ `,
	` (_-</ _ \/ // / __/ -_)`,
	`/___/ .__/\_, /_/  \__/ `,
	`   /_/   /___/          `,
}

func displayLogo() {
	ct.Foreground(ct.Blue, true)
	fmt.Println(logo[0])
	fmt.Print(logo[1])
	ct.ResetColor()
	fmt.Printf("          version %s\n", spyre.Version)
	ct.Foreground(ct.Magenta, true)
	fmt.Println(logo[2])
	ct.Foreground(ct.Blue, true)
	fmt.Print(logo[3])
	ct.ResetColor()
	fmt.Println("          https://github.com/spyre-project/spyre")
	fmt.Println()
}
