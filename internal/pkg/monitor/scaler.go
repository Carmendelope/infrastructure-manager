/*
 * Copyright 2019 Nalej
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
 *
 */

package monitor

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"time"
)

// ScalerMonitor structure with the client required to monitor a scale request.
type ScalerMonitor struct {
	scaleClient         grpc_provisioner_go.ScaleClient
	provisionerResponse grpc_infrastructure_manager_go.ProvisionerResponse
	callback            func(string, string, string, *grpc_provisioner_go.ScaleClusterResponse, derrors.Error)
}

// NewScalerMonitor creates a new monitor with a set of clients.
func NewScalerMonitor(
	scaleClient grpc_provisioner_go.ScaleClient,
	provisionerResponse grpc_infrastructure_manager_go.ProvisionerResponse) *ScalerMonitor {
	return &ScalerMonitor{
		scaleClient:         scaleClient,
		provisionerResponse: provisionerResponse,
		callback:            nil,
	}
}

// RegisterCallback registers a callback function that will be triggered
// when the scale of a cluster finishes.
func (m *ScalerMonitor) RegisterCallback(callback func(requestID string, organizationID string, clusterID string, lastResponse *grpc_provisioner_go.ScaleClusterResponse, err derrors.Error)) {
	m.callback = callback
}

// LaunchMonitor periodically monitors the state of a scaling operation waiting for it to complete.
func (m *ScalerMonitor) LaunchMonitor() {
	log.Debug().Str("clusterID", m.provisionerResponse.ClusterId).
		Str("requestID", m.provisionerResponse.RequestId).Msg("Launching scaler monitor")

	requestID := &grpc_common_go.RequestId{
		RequestId: m.provisionerResponse.RequestId,
	}
	exit := false
	remainingFailures := MaxConnFailures
	var status *grpc_provisioner_go.ScaleClusterResponse
	var err error
	for !exit {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		status, err = m.scaleClient.CheckProgress(ctx, requestID)
		if err != nil {
			log.Debug().Str("err", conversions.ToDerror(err).DebugReport()).Msg("error requesting scaling status")
			remainingFailures--
			if remainingFailures == 0 {
				log.Warn().Str("requestID", requestID.RequestId).Msg("Cannot contact provisioner component")
				exit = true
			} else {
				time.Sleep(ConnectRetryDelay)
			}
		} else {
			log.Debug().Str("requestID", requestID.RequestId).Int64("elapsed", status.ElapsedTime).Str("state", status.State.String()).Msg("processing scaling progress")
			if status.Error != "" || status.State == grpc_provisioner_go.ProvisionProgress_ERROR || status.State == grpc_provisioner_go.ProvisionProgress_FINISHED {
				exit = true
			} else {
				time.Sleep(QueryDelay)
			}
		}
	}
	m.notify(status, err)
	log.Debug().Str("clusterID", m.provisionerResponse.ClusterId).
		Str("requestID", m.provisionerResponse.RequestId).Msg("Scale monitor exits")
}

// notify informs the associated callback that the scaling has finished.
func (m *ScalerMonitor) notify(lastResponse *grpc_provisioner_go.ScaleClusterResponse, err error) {
	requestID := &grpc_common_go.RequestId{
		RequestId: m.provisionerResponse.RequestId,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	_, rErr := m.scaleClient.RemoveScale(ctx, requestID)
	if rErr != nil {
		log.Error().Str("requestID", requestID.RequestId).
			Str("err", conversions.ToDerror(rErr).DebugReport()).Msg("Cannot remove scale operation from provisioner")
	}
	var cErr derrors.Error
	if err != nil {
		cErr = conversions.ToDerror(err)
	}
	if m.callback != nil {
		m.callback(requestID.RequestId, m.provisionerResponse.OrganizationId, m.provisionerResponse.ClusterId,
			lastResponse, cErr)
	} else {
		log.Warn().Str("requestID", requestID.RequestId).
			Msg("no callback registered")
	}
}
