package main

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed Logo.png
var logoBytes []byte

var appIcon fyne.Resource = fyne.NewStaticResource("Logo.png", logoBytes)
