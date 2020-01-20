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
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"time"
)

// DecomissionerMonitor structure with the required clients to read and update states of a decomission.
type DecomissionerMonitor struct {
	decomissionerClient grpc_provisioner_go.DecomissionClient
	clusterId           string
	requestId           string
	callback            func(string, *grpc_common_go.OpResponse, derrors.Error)
}

// NewDecomissionerMonitor creates a new monitor with a set of clients.
func NewDecomissionerMonitor(
	decomissionerClient grpc_provisioner_go.DecomissionClient,
	clusterId string,
	requestId string) *DecomissionerMonitor {
	return &DecomissionerMonitor{
		decomissionerClient: decomissionerClient,
		clusterId:           clusterId,
		requestId:           requestId,
		callback:            nil,
	}
}

// RegisterCallback registers a callback function that will be triggered
// when the decomission of a cluster finishes.
func (m *DecomissionerMonitor) RegisterCallback(callback func(clusterID string, lastResponse *grpc_common_go.OpResponse, err derrors.Error)) {
	m.callback = callback
}

// LaunchMonitor periodically monitors the state of a decomission waiting for it to complete.
func (m *DecomissionerMonitor) LaunchMonitor() {
	log.Debug().Str("clusterID", m.clusterId).
		Str("requestID", m.requestId).Msg("Launching decomission monitor")

	requestID := &grpc_common_go.RequestId{
		RequestId: m.requestId,
	}
	exit := false
	remainingFailures := MaxConnFailures
	var status *grpc_common_go.OpResponse
	var err error
	for !exit {
		status, err = m.decomissionerClient.CheckProgress(context.Background(), requestID)
		if err != nil {
			log.Debug().Str("err", conversions.ToDerror(err).DebugReport()).Msg("error requesting decomissioning status")
			remainingFailures--
			if remainingFailures == 0 {
				log.Warn().Str("requestID", requestID.RequestId).Msg("Cannot contact provisioner")
				exit = true
			} else {
				time.Sleep(ConnectRetryDelay)
			}
		} else {
			log.Debug().Str("requestID", requestID.RequestId).Int64("elapsed", status.ElapsedTime).Str("state", status.Status.String()).Msg("processing decomission progress")
			if status.Error != "" || status.Status == grpc_common_go.OpStatus_FAILED || status.Status == grpc_common_go.OpStatus_CANCELED || status.Status == grpc_common_go.OpStatus_SUCCESS {
				exit = true
			} else {
				time.Sleep(QueryDelay)
			}
		}
	}
	m.notify(status, err)
	log.Debug().Str("clusterID", m.clusterId).
		Str("requestID", m.requestId).Msg("Decomission monitor exits")
}

// notify informs the associated callback that the installation has finished.
func (m *DecomissionerMonitor) notify(lastResponse *grpc_common_go.OpResponse, err error) {
	requestID := &grpc_common_go.RequestId{
		RequestId: lastResponse.GetRequestId(),
	}
	_, rErr := m.decomissionerClient.RemoveDecomission(context.Background(), requestID)
	if rErr != nil {
		log.Error().Str("requestID", requestID.RequestId).
			Str("err", conversions.ToDerror(rErr).DebugReport()).Msg("Cannot remove decomission from provisioner")
	}
	var cErr derrors.Error
	if err != nil {
		cErr = conversions.ToDerror(err)
	}
	if m.callback != nil {
		m.callback(m.clusterId, lastResponse, cErr)
	} else {
		log.Warn().Str("requestID", requestID.RequestId).
			Msg("no callback registered")
	}
}
