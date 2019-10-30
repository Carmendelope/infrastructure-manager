/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package monitor

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"time"
)

// InstallerMonitor structure with the required clients to read and update states of an install.
type InstallerMonitor struct {
	installerClient   grpc_installer_go.InstallerClient
	clusterClient     grpc_infrastructure_go.ClustersClient
	installerResponse grpc_infrastructure_manager_go.InstallResponse
	callback          func(string, string, string, *grpc_installer_go.InstallResponse, derrors.Error)
}

// NewInstallerMonitor creates a new monitor with a set of clients.
func NewInstallerMonitor(
	installerClient grpc_installer_go.InstallerClient,
	clusterClient grpc_infrastructure_go.ClustersClient,
	installerResponse grpc_infrastructure_manager_go.InstallResponse) *InstallerMonitor {
	return &InstallerMonitor{
		installerClient:   installerClient,
		clusterClient:     clusterClient,
		installerResponse: installerResponse,
		callback:          nil,
	}
}

// RegisterCallback registers a callback function that will be triggered
// when the installation of a cluster finishes.
func (m *InstallerMonitor) RegisterCallback(callback func(installID string, organizationID string, clusterID string, lastResponse *grpc_installer_go.InstallResponse, err derrors.Error)) {
	m.callback = callback
}

// LaunchMonitor periodically monitors the state of an install waiting for it to complete.
func (m *InstallerMonitor) LaunchMonitor() {
	log.Debug().Str("clusterID", m.installerResponse.ClusterId).
		Str("installID", m.installerResponse.InstallId).Msg("Launching installation monitor")

	installID := &grpc_installer_go.InstallId{
		InstallId: m.installerResponse.InstallId,
	}
	exit := false
	remainingFailures := MaxConnFailures
	var status *grpc_installer_go.InstallResponse
	var err error
	for !exit {
		status, err = m.installerClient.CheckProgress(context.Background(), installID)
		if err != nil {
			log.Debug().Str("err", conversions.ToDerror(err).DebugReport()).Msg("error requesting installing status")
			remainingFailures--
			if remainingFailures == 0 {
				log.Warn().Str("installID", installID.InstallId).Msg("Cannot contact installer")
				exit = true
			} else {
				time.Sleep(ConnectRetryDelay)
			}
		} else {
			log.Debug().Str("installID", installID.InstallId).Int64("elapsed", status.ElapsedTime).Msg("processing install progress")
			if status.Error != "" || status.State == grpc_installer_go.InstallProgress_ERROR || status.State == grpc_installer_go.InstallProgress_FINISHED {
				exit = true
			} else {
				time.Sleep(QueryDelay)
			}
		}
	}
	m.notify(status, err)
	log.Debug().Str("clusterID", m.installerResponse.ClusterId).
		Str("installID", m.installerResponse.InstallId).Msg("Install monitor exits")
}

// notify informs the associated callback that the installation has finished.
func (m *InstallerMonitor) notify(lastResponse *grpc_installer_go.InstallResponse, err error) {
	removeInstallRequest := &grpc_installer_go.RemoveInstallRequest{
		InstallId: lastResponse.InstallId,
	}
	_, rErr := m.installerClient.RemoveInstall(context.Background(), removeInstallRequest)
	if rErr != nil {
		log.Error().Str("installID", m.installerResponse.InstallId).
			Str("err", conversions.ToDerror(rErr).DebugReport()).Msg("Cannot remove install from installer")
	}
	var cErr derrors.Error
	if err != nil {
		cErr = conversions.ToDerror(err)
	}
	if m.callback != nil {
		m.callback(m.installerResponse.InstallId, m.installerResponse.OrganizationId, m.installerResponse.ClusterId,
			lastResponse, cErr)
	} else {
		log.Warn().Str("installID", m.installerResponse.InstallId).
			Msg("no callback registered")
	}
}
