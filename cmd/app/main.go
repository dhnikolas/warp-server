package main

import (
	"context"
	"fmt"
	"github.com/jroimartin/gocui"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"syscall"
	"time"
	"warp-server/api"
	"warp-server/internal/config"
	"warp-server/internal/controllers"
	repoVPN "warp-server/internal/repositories/cisco"
	"warp-server/internal/repositories/packetfilter"
	"warp-server/internal/repositories/sshkeys"
	"warp-server/internal/repositories/sshtunnel"
	"warp-server/internal/services/fw"
	"warp-server/internal/services/tunnel"
	"warp-server/internal/services/vpn"
	"warp-server/internal/ui"
	cl "warp-server/pkg/controlloop"
	"warp-server/pkg/log"
)

func main() {

	currentUser, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Cannot get current user: %v", err))
	}
	if currentUser.Gid == "0" || currentUser.Uid == "0" {
		panic(fmt.Sprintf("Don't run from root"))
	}

	l := &ui.LogWriter{Logs: make(chan string, 100)}
	log.SetOutput(l)
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("Error loading config: %s ", err))
	}

	newResource := cl.NewResource("newResource")
	mc := &api.MainConfig{
		Resource: *newResource,
		Spec:     api.MainConfigSpec{},
	}

	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	privateKeyPath := filepath.Join(sshDir, "id_rsa_warp")
	publicKeyPath := filepath.Join(sshDir, "id_rsa_warp.pub")
	sshKeysRepo := sshkeys.NewRepository(cfg.LocalUsername, cfg.LocalHost, sshDir)
	sshTunnelRepo := sshtunnel.NewRepository(cfg.LocalUsername, cfg.LocalHost, cfg.TunnelAddress)

	vpnService := vpn.NewService(repoVPN.NewRepository(cfg.CiscoHost, cfg.CiscoUsername, cfg.CiscoPassword))
	fwService := fw.NewService(packetfilter.NewRepository(cfg.LocalPassword))
	tunnelService := tunnel.NewService(publicKeyPath, privateKeyPath, sshKeysRepo, sshTunnelRepo)

	conditionChannel := make(chan []cl.Condition, 100)

	mainController := controllers.NewMainReconcile(
		conditionChannel,
		vpnService,
		fwService,
		tunnelService,
	)

	mainLoop := cl.New[*api.MainConfig](mainController, cl.WithLogger(log.NewLogger()))
	mainLoop.Storage.Add(mc)
	mainLoop.Run()
	log.Info().Msg("Main", "Start run main loop")

	ctxExit, cancel := context.WithCancel(context.Background())

	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		panic(fmt.Errorf("Error loading gui: %s ", err))
	}
	go func() {
		defer cancel()

		err := ui.CreateTUI(cancel, g, l, conditionChannel)
		if err != nil {
			fmt.Println("CreateTUI error:", err)
		}
	}()

	go func() {
		defer cancel()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
	}()
	log.Info().Msg("Main", "Warp-server started!")
	<-ctxExit.Done()

	log.Info().Msg("Main", "Stopping warp-server...")
	mainLoop.Stop()
	log.Info().Msg("Main", "warp-server stopped")
	time.Sleep(time.Second * 2)
	g.Close()

	fmt.Println("Application stopped")
}
