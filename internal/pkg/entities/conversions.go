/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-installer-go"
)

func StateToStatus(state grpc_installer_go.InstallProgress) grpc_infrastructure_go.InfraStatus {
	var newStatus grpc_infrastructure_go.InfraStatus
	switch state {
	case grpc_installer_go.InstallProgress_REGISTERED:
		newStatus = grpc_infrastructure_go.InfraStatus_INSTALLING
	case grpc_installer_go.InstallProgress_IN_PROGRESS:
		newStatus = grpc_infrastructure_go.InfraStatus_INSTALLING
	case grpc_installer_go.InstallProgress_FINISHED:
		newStatus = grpc_infrastructure_go.InfraStatus_RUNNING
	case grpc_installer_go.InstallProgress_ERROR:
		newStatus = grpc_infrastructure_go.InfraStatus_ERROR
	}
	return newStatus
}
