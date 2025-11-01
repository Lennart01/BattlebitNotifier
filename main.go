package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

type Server struct {
	Map          string `json:"Map"`
	MapSize      string `json:"MapSize"`
	Gamemode     string `json:"Gamemode"`
	Region       string `json:"Region"`
	Players      int    `json:"Players"`
	QueuePlayers int    `json:"QueuePlayers"`
	MaxPlayers   int    `json:"MaxPlayers"`
	Hz           int    `json:"Hz"`
	DayNight     string `json:"DayNight"`
	IsOfficial   bool   `json:"IsOfficial"`
	HasPassword  bool   `json:"HasPassword"`
	AntiCheat    string `json:"AntiCheat"`
	Build        string `json:"Build"`
}

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "BattleBit Server Alerter",
		Width:  600,
		Height: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
