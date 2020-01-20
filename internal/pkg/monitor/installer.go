/*
 * Copyright 2020 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package monitor

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"time"
)

// InstallerMonitor structure with the required clients to read and update states of an install.
type InstallerMonitor struct {
	clusterID           string
	installerClient     grpc_installer_go.InstallerClient
	clusterClient       grpc_infrastructure_go.ClustersClient
	installerResponse   grpc_common_go.OpResponse
	callback            func(string, string, string, *grpc_common_go.OpResponse, derrors.Error)
	decomissionCallback *DecomissionCallback
}

// DecomissionCallback is a structure to handle the callback function and required parameters to execute it
type DecomissionCallback struct {
	Callback func(*grpc_provisioner_go.DecomissionClusterRequest)
	Request  *grpc_provisioner_go.DecomissionClusterRequest
}

// NewInstallerMonitor creates a new monitor with a set of clients.
func NewInstallerMonitor(
	clusterID string,
	installerClient grpc_installer_go.InstallerClient,
	clusterClient grpc_infrastructure_go.ClustersClient,
	installerResponse grpc_common_go.OpResponse) *InstallerMonitor {
	return &InstallerMonitor{
		clusterID:         clusterID,
		installerClient:   installerClient,
		clusterClient:     clusterClient,
		installerResponse: installerResponse,
		callback:          nil,
	}
}

// RegisterCallback registers a callback function that will be triggered
// when the installation of a cluster finishes.
func (m *InstallerMonitor) RegisterCallback(callback func(installID string, organizationID string, clusterID string, lastResponse *grpc_common_go.OpResponse, err derrors.Error)) {
	m.callback = callback
}

func (m *InstallerMonitor) RegisterDecomissionCallback(callback *DecomissionCallback) {
	m.decomissionCallback = callback
}

// LaunchMonitor periodically monitors the state of an install waiting for it to complete.
func (m *InstallerMonitor) LaunchMonitor() {
	log.Debug().Str("requestID", m.installerResponse.RequestId).
		Str("OrganizationID", m.installerResponse.OrganizationId).Str("clusterID", m.clusterID).
		Msg("Launching installer monitor")

	requestID := &grpc_common_go.RequestId{
		RequestId: m.installerResponse.RequestId,
	}
	exit := false
	remainingFailures := MaxConnFailures
	var response *grpc_common_go.OpResponse
	var err error
	for !exit {
		response, err = m.installerClient.CheckProgress(context.Background(), requestID)
		if err != nil {
			log.Debug().Str("err", conversions.ToDerror(err).DebugReport()).Msg("error requesting installing status")
			remainingFailures--
			if remainingFailures == 0 {
				log.Warn().Str("requestID", requestID.RequestId).Msg("Cannot contact installer")
				exit = true
			} else {
				time.Sleep(ConnectRetryDelay)
			}
		} else {
			log.Debug().Str("requestID", response.RequestId).Int64("elapsed", response.ElapsedTime).Msg("processing operation progress")
			if response.Error != "" || response.Status == grpc_common_go.OpStatus_FAILED || response.Status == grpc_common_go.OpStatus_SUCCESS {
				exit = true
			} else {
				time.Sleep(QueryDelay)
			}
		}
	}
	m.notify(response, err)
	log.Debug().Str("requestID", response.RequestId).Str("organizationID", response.OrganizationId).
		Str("clusterID", m.clusterID).Msg("Installer monitor exits")
}

// notify informs the associated callback that the installation has finished.
func (m *InstallerMonitor) notify(lastResponse *grpc_common_go.OpResponse, err error) {
	removeInstallRequest := &grpc_common_go.RequestId{
		RequestId: lastResponse.RequestId,
	}
	_, rErr := m.installerClient.RemoveInstall(context.Background(), removeInstallRequest)
	if rErr != nil {
		log.Error().Str("requestID", m.installerResponse.RequestId).
			Str("err", conversions.ToDerror(rErr).DebugReport()).Msg("Cannot remove operation from installer")
	}
	var cErr derrors.Error
	if err != nil {
		cErr = conversions.ToDerror(err)
	}
	if m.callback != nil {
		m.callback(m.installerResponse.RequestId,
			m.installerResponse.OrganizationId, m.clusterID,
			lastResponse, cErr)
	} else {
		log.Warn().Str("requestID", m.installerResponse.RequestId).
			Msg("no callback registered")
	}
	if m.decomissionCallback != nil {
		log.Debug().Msg("Launch decommission callback")
		m.decomissionCallback.Callback(m.decomissionCallback.Request)
	}
}
