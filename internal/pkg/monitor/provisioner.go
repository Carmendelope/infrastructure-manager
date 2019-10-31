/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package monitor

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/rs/zerolog/log"
	"time"
)

// ProvisionerMonitor structure with the required clients to read and update states of a provision.
type ProvisionerMonitor struct {
	provisionerClient   grpc_provisioner_go.ProvisionClient
	clusterClient       grpc_infrastructure_go.ClustersClient
	provisionerResponse grpc_infrastructure_manager_go.ProvisionerResponse
	callback            func(string, string, string, *grpc_provisioner_go.ProvisionClusterResponse, derrors.Error)
}

// NewProvisionerMonitor creates a new monitor with a set of clients.
func NewProvisionerMonitor(
	provisionerClient grpc_provisioner_go.ProvisionClient,
	clusterClient grpc_infrastructure_go.ClustersClient,
	provisionerResponse grpc_infrastructure_manager_go.ProvisionerResponse) *ProvisionerMonitor {
	return &ProvisionerMonitor{
		provisionerClient:   provisionerClient,
		clusterClient:       clusterClient,
		provisionerResponse: provisionerResponse,
		callback:            nil,
	}
}

// RegisterCallback registers a callback function that will be triggered
// when the provision of a cluster finishes.
func (m *ProvisionerMonitor) RegisterCallback(callback func(requestID string, organizationID string, clusterID string, lastResponse *grpc_provisioner_go.ProvisionClusterResponse, err derrors.Error)) {
	m.callback = callback
}

// LaunchMonitor periodically monitors the state of a provision waiting for it to complete.
func (m *ProvisionerMonitor) LaunchMonitor() {
	log.Debug().Str("clusterID", m.provisionerResponse.ClusterId).
		Str("requestID", m.provisionerResponse.RequestId).Msg("Launching provision monitor")

	requestID := &grpc_common_go.RequestId{
		RequestId: m.provisionerResponse.RequestId,
	}
	exit := false
	remainingFailures := MaxConnFailures
	var status *grpc_provisioner_go.ProvisionClusterResponse
	var err error
	for !exit {
		status, err = m.provisionerClient.CheckProgress(context.Background(), requestID)
		if err != nil {
			log.Debug().Str("err", conversions.ToDerror(err).DebugReport()).Msg("error requesting provisioning status")
			remainingFailures--
			if remainingFailures == 0 {
				log.Warn().Str("requestID", requestID.RequestId).Msg("Cannot contact provisioner")
				exit = true
			} else {
				time.Sleep(ConnectRetryDelay)
			}
		} else {
			log.Debug().Str("requestID", requestID.RequestId).Int64("elapsed", status.ElapsedTime).Str("state", status.State.String()).Msg("processing provision progress")
			if status.Error != "" || status.State == grpc_provisioner_go.ProvisionProgress_ERROR || status.State == grpc_provisioner_go.ProvisionProgress_FINISHED {
				exit = true
			} else {
				time.Sleep(QueryDelay)
			}
		}
	}
	m.notify(status, err)
	log.Debug().Str("clusterID", m.provisionerResponse.ClusterId).
		Str("requestID", m.provisionerResponse.RequestId).Msg("Provision monitor exits")
}

// notify informs the associated callback that the installation has finished.
func (m *ProvisionerMonitor) notify(lastResponse *grpc_provisioner_go.ProvisionClusterResponse, err error) {
	requestID := &grpc_common_go.RequestId{
		RequestId: m.provisionerResponse.RequestId,
	}
	_, rErr := m.provisionerClient.RemoveProvision(context.Background(), requestID)
	if rErr != nil {
		log.Error().Str("requestID", requestID.RequestId).
			Str("err", conversions.ToDerror(rErr).DebugReport()).Msg("Cannot remove provision from provisioner")
	}
	var cErr derrors.Error
	if err != nil {
		cErr = conversions.ToDerror(err)
	}
	if m.callback != nil {
		m.callback(requestID.RequestId, m.provisionerResponse.OrganizationId, m.provisionerResponse.ClusterId,
			lastResponse, cErr)
	} else {
		log.Warn().Str("installID", requestID.RequestId).
			Msg("no callback registered")
	}
}
