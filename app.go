package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	servers  []Server
	mapNames []string

	alertLock  sync.RWMutex
	alertMaps  []string
	minPlayers int

	dataLock sync.RWMutex
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.fetchDataAndMaps()
	go a.pollServerData()
}

func (a *App) pollServerData() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.fetchDataAndMaps()
			a.checkAlert()
		}
	}
}

func (a *App) fetchDataAndMaps() {
	servers, err := getData()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Error fetching data: %v", err)
		return
	}

	newMapNames := extractMapNames(servers)

	a.dataLock.Lock()
	a.servers = servers
	needsUpdate := !slices.Equal(a.mapNames, newMapNames)
	if needsUpdate {
		a.mapNames = newMapNames
	}
	a.dataLock.Unlock()

	if needsUpdate {
		runtime.EventsEmit(a.ctx, "mapsUpdated", newMapNames)
	}
}

func (a *App) checkAlert() {
	a.alertLock.RLock()
	targetMaps := a.alertMaps
	targetPlayers := a.minPlayers
	a.alertLock.RUnlock()

	if len(targetMaps) == 0 || targetPlayers == 0 {
		return
	}

	targetMapSet := make(map[string]bool, len(targetMaps))
	for _, m := range targetMaps {
		targetMapSet[m] = true
	}

	a.dataLock.RLock()
	defer a.dataLock.RUnlock()

	totalPlayers := 0
	serverCount := 0
	for _, server := range a.servers {
		if targetMapSet[server.Map] {
			totalPlayers += server.Players
			serverCount++
		}
	}

	if totalPlayers >= targetPlayers {
		runtime.LogInfof(a.ctx, "Alert triggered: Maps=%v, Players=%d", targetMaps, totalPlayers)

		dialogOptions := runtime.MessageDialogOptions{
			Type:    runtime.InfoDialog,
			Title:   "BattleBit Alert",
			Message: fmt.Sprintf("%d players found on your selected maps (%d servers)!", totalPlayers, serverCount),
		}
		_, err := runtime.MessageDialog(a.ctx, dialogOptions)
		if err != nil {
			runtime.LogErrorf(a.ctx, "Error showing message dialog: %v", err)
		}

		a.alertLock.Lock()
		a.alertMaps = nil
		a.minPlayers = 0
		a.alertLock.Unlock()

		runtime.EventsEmit(a.ctx, "alertTriggered")
	}
}

func (a *App) GetMapList() []string {
	a.dataLock.RLock()
	defer a.dataLock.RUnlock()
	return a.mapNames
}

func (a *App) SetAlert(mapNames []string, minPlayers int) {
	a.alertLock.Lock()
	defer a.alertLock.Unlock()
	a.alertMaps = mapNames
	a.minPlayers = minPlayers

	runtime.LogInfof(a.ctx, "Alert set: Maps=%v, MinPlayers=%d", mapNames, minPlayers)
}

func (a *App) CancelAlert() {
	a.alertLock.Lock()
	defer a.alertLock.Unlock()
	a.alertMaps = nil
	a.minPlayers = 0
	runtime.LogInfo(a.ctx, "Alert cancelled by user")
}

func extractMapNames(servers []Server) []string {
	mapSet := make(map[string]bool)
	for _, server := range servers {
		if server.Map != "" {
			mapSet[server.Map] = true
		}
	}

	maps := make([]string, 0, len(mapSet))
	for m := range mapSet {
		maps = append(maps, m)
	}
	slices.Sort(maps)
	return maps
}

func getData() ([]Server, error) {
	url := "https://publicapi.battlebit.cloud/Servers/GetServerList"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	bom := []byte{0xEF, 0xBB, 0xBF}
	bodyBytes = bytes.TrimPrefix(bodyBytes, bom)

	var servers []Server
	err = json.Unmarshal(bodyBytes, &servers)
	if err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}

	return servers, nil
}
