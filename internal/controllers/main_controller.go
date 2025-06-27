package controllers

import (
	"context"
	"time"
	"warp-server/api"
	"warp-server/internal/config"
	"warp-server/internal/dto"
	"warp-server/internal/services/fw"
	"warp-server/internal/services/tunnel"
	"warp-server/internal/services/vpn"
	cl "warp-server/pkg/controlloop"
	"warp-server/pkg/log"
)

func NewMainReconcile(
	cfg *config.Config,
	conditionsChan chan []cl.Condition,
	vpnService *vpn.Service,
	fwService *fw.Service,
	tunnelService *tunnel.Service,
) *MainReconcile {
	return &MainReconcile{
		cfg:            cfg,
		conditionsChan: conditionsChan,
		vpnService:     vpnService,
		fwService:      fwService,
		tunnelService:  tunnelService,
	}
}

type MainReconcile struct {
	cfg            *config.Config
	conditionsChan chan []cl.Condition
	vpnService     *vpn.Service
	fwService      *fw.Service
	tunnelService  *tunnel.Service
}

func (r *MainReconcile) Reconcile(ctx context.Context, object *api.MainConfig) (cl.Result, error) {
	config := object // TODO: rename? Variable 'config' collides with imported package name

	defer func() {
		r.conditionsChan <- config.GetConditions()
	}()

	if config.GetKillTimestamp() != "" {
		return r.reconcileKill(ctx, config)
	}

	return r.reconcileNormal(ctx, config)
}

func (r *MainReconcile) reconcileNormal(ctx context.Context, mc *api.MainConfig) (cl.Result, error) {
	vpnState, _, err := r.vpnService.GetState()
	if err != nil {
		mc.MarkFalse(api.VPNConnectedCondition, api.VPNConnectionStateFailedReason, err.Error())
		return cl.Result{}, err
	}
	if vpnState != dto.VPNStateConnected {
		log.Info().Msg("Main", "Connecting to VPN...")
		err = r.vpnService.Connect()
		if err != nil {
			mc.MarkFalse(api.VPNConnectedCondition, api.VPNConnectionFailedReason, err.Error())
			return cl.Result{}, err
		}
		log.Info().Msg("Main", "Connecting to VPN success!")
		return cl.Result{RequeueAfter: time.Second}, nil
	}
	mc.MarkTrue(api.VPNConnectedCondition)
	if r.cfg.VpnOnly {
		// imitate success condition
		mc.MarkTrue(api.PFDisabledCondition)
		mc.MarkTrue(api.SSHKeysInstalledCondition)
		mc.MarkTrue(api.TunnelEnabledCondition)

		return cl.Result{RequeueAfter: time.Second * 20}, nil
	}
	err = r.fwService.Disable()
	if err != nil {
		mc.MarkFalse(api.PFDisabledCondition, api.PFDisabledFailedReason, err.Error())
		return cl.Result{}, err
	}
	mc.MarkTrue(api.PFDisabledCondition)
	err = r.tunnelService.SetupSSHKey()
	if err != nil {
		mc.MarkFalse(api.SSHKeysInstalledCondition, api.SSHKeysFailedReason, err.Error())
		return cl.Result{}, err
	}
	mc.MarkTrue(api.SSHKeysInstalledCondition)
	err = r.tunnelService.StartTunnel()
	if err != nil {
		log.Info().Err(err)
		mc.MarkFalse(api.TunnelEnabledCondition, api.TunnelInitializationFailedReason, err.Error())
		return cl.Result{}, err
	}
	mc.MarkTrue(api.TunnelEnabledCondition)
	return cl.Result{RequeueAfter: time.Second * 20}, nil
}

func (r *MainReconcile) reconcileKill(ctx context.Context, mc *api.MainConfig) (cl.Result, error) {
	log.Info().Msg("Main", "Reconcile Kill")

	pid, ok, err := r.tunnelService.GetTunnelPID()
	if err != nil {
		mc.MarkFalse(api.TunnelEnabledCondition, api.TunnelDisablingFailedReason, err.Error())
		return cl.Result{}, err
	}
	log.Info().Msg("Main", "Reconcile Kill PID", pid)

	if ok {
		err = r.tunnelService.StopTunnel(ctx, pid)
		if err != nil {
			mc.MarkFalse(api.TunnelEnabledCondition, api.TunnelDisablingFailedReason, err.Error())
			return cl.Result{}, err
		}
		log.Info().Msg("Main", "Reconcile Kill STOP TUNNEL", pid)
	}

	vpnState, _, err := r.vpnService.GetState()
	if err != nil {
		return cl.Result{}, err
	}

	if vpnState == dto.VPNStateConnected {
		err = r.vpnService.Disconnect()
		if err != nil {
			mc.MarkFalse(api.VPNConnectedCondition, api.VPNDisconnectFailedReason, err.Error())
			return cl.Result{}, err
		}
		log.Info().Msg("Main", "Reconcile Kill STOP VPN")
	}

	log.Info().Msg("Main", "Reconcile Kill STOPPED")

	return cl.Result{}, nil
}
