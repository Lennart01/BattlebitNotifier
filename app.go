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
	ctx       context.Context
	servers   []Server
	gamemodes []string
	regions   []string

	alertLock      sync.RWMutex
	alertConfig    AlertConfig
	serverStates   map[string]string
	alertedServers map[string]bool

	dataLock sync.RWMutex
}

func NewApp() *App {
	return &App{
		serverStates:   make(map[string]string),
		alertedServers: make(map[string]bool),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	go a.fetchStaticData()
	go a.pollServerData()
}

func (a *App) pollServerData() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	a.fetchServerData()
	a.checkAlerts()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.fetchServerData()
			a.checkAlerts()
		}
	}
}

func (a *App) fetchStaticData() {
	servers, err := getData()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Error fetching initial data: %v", err)
		return
	}

	gamemodeSet := make(map[string]bool)
	regionSet := make(map[string]bool)

	for _, server := range servers {
		if server.Gamemode != "" {
			gamemodeSet[server.Gamemode] = true
		}
		if server.Region != "" {
			regionSet[server.Region] = true
		}
	}

	gamemodes := make([]string, 0, len(gamemodeSet))
	for m := range gamemodeSet {
		gamemodes = append(gamemodes, m)
	}
	slices.Sort(gamemodes)

	regions := make([]string, 0, len(regionSet))
	for r := range regionSet {
		regions = append(regions, r)
	}
	slices.Sort(regions)

	a.dataLock.Lock()
	a.gamemodes = gamemodes
	a.regions = regions
	a.servers = servers
	a.dataLock.Unlock()

	runtime.LogInfo(a.ctx, "Fetched static filter lists")
}

func (a *App) fetchServerData() {
	servers, err := getData()
	if err != nil {
		runtime.LogErrorf(a.ctx, "Error fetching server data: %v", err)
		return
	}

	a.dataLock.Lock()
	a.servers = servers
	a.dataLock.Unlock()
}

func (a *App) checkAlerts() {
	a.alertLock.RLock()
	config := a.alertConfig
	a.alertLock.RUnlock()

	if len(config.Maps) == 0 {
		return
	}

	targetMapSet := make(map[string]bool, len(config.Maps))
	for _, m := range config.Maps {
		targetMapSet[m] = true
	}

	targetGamemodeSet := make(map[string]bool, len(config.Gamemodes))
	for _, g := range config.Gamemodes {
		targetGamemodeSet[g] = true
	}

	targetRegionSet := make(map[string]bool, len(config.Regions))
	for _, r := range config.Regions {
		targetRegionSet[r] = true
	}

	a.dataLock.RLock()
	servers := a.servers
	a.dataLock.RUnlock()

	a.alertLock.Lock()
	defer a.alertLock.Unlock()

	for _, server := range servers {
		serverName := server.Name
		currentMap := server.Map

		if serverName == "" {
			continue
		}

		previousMap := a.serverStates[serverName]

		if currentMap != previousMap {
			delete(a.alertedServers, serverName)
		}

		hasBeenAlerted := a.alertedServers[serverName]

		if !hasBeenAlerted &&
			targetMapSet[currentMap] &&
			(len(targetGamemodeSet) == 0 || targetGamemodeSet[server.Gamemode]) &&
			(len(targetRegionSet) == 0 || targetRegionSet[server.Region]) &&
			server.Players >= config.MinPlayers {

			runtime.LogInfof(a.ctx, "Alert triggered: Server '%s' switched to map '%s' with %d players", serverName, currentMap, server.Players)

			dialogOptions := runtime.MessageDialogOptions{
				Type:    runtime.InfoDialog,
				Title:   "BattleBit Alert",
				Message: fmt.Sprintf("Server '%s' is now playing '%s' with %d players!", serverName, currentMap, server.Players),
			}
			_, err := runtime.MessageDialog(a.ctx, dialogOptions)
			if err != nil {
				runtime.LogErrorf(a.ctx, "Error showing message dialog: %v", err)
			}

			a.alertedServers[serverName] = true
		}

		a.serverStates[serverName] = currentMap
	}
}

func (a *App) GetFilterLists() FilterLists {
	a.dataLock.RLock()
	defer a.dataLock.RUnlock()
	return FilterLists{
		Gamemodes: a.gamemodes,
		Regions:   a.regions,
	}
}

func (a *App) SetAlert(config AlertConfig) {
	a.alertLock.Lock()
	defer a.alertLock.Unlock()
	a.alertConfig = config
	a.alertedServers = make(map[string]bool)

	runtime.LogInfof(a.ctx, "Alert set: Maps=%v, Gamemodes=%v, Regions=%v, MinPlayers=%d", config.Maps, config.Gamemodes, config.Regions, config.MinPlayers)
	runtime.EventsEmit(a.ctx, "alertSet")
}

func (a *App) CancelAlert() {
	a.alertLock.Lock()
	defer a.alertLock.Unlock()
	a.alertConfig = AlertConfig{}
	a.alertedServers = make(map[string]bool)
	runtime.LogInfo(a.ctx, "Alert cancelled by user")
	runtime.EventsEmit(a.ctx, "alertCancelled")
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
