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
 */

package entities

import (
	grpc_common_go "github.com/nalej/grpc-common-go"
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

func InstallStateToNodeState(state grpc_installer_go.InstallProgress) grpc_infrastructure_go.NodeState {
	var newState grpc_infrastructure_go.NodeState
	switch state {
	case grpc_installer_go.InstallProgress_REGISTERED:
		newState = grpc_infrastructure_go.NodeState_ASSIGNED
	case grpc_installer_go.InstallProgress_IN_PROGRESS:
		newState = grpc_infrastructure_go.NodeState_ASSIGNED
	case grpc_installer_go.InstallProgress_FINISHED:
		newState = grpc_infrastructure_go.NodeState_ASSIGNED
	case grpc_installer_go.InstallProgress_ERROR:
		newState = grpc_infrastructure_go.NodeState_UNREGISTERED
	}
	return newState
}

func OpStatusToNodeState(status grpc_common_go.OpStatus) grpc_infrastructure_go.NodeState {
	var newState grpc_infrastructure_go.NodeState
	switch status {
	case grpc_common_go.OpStatus_INIT:
		newState = grpc_infrastructure_go.NodeState_UNASSIGNED
	case grpc_common_go.OpStatus_SCHEDULED:
		newState = grpc_infrastructure_go.NodeState_UNASSIGNED
	case grpc_common_go.OpStatus_INPROGRESS:
		newState = grpc_infrastructure_go.NodeState_UNASSIGNED
	case grpc_common_go.OpStatus_SUCCESS:
		newState = grpc_infrastructure_go.NodeState_ASSIGNED
	case grpc_common_go.OpStatus_FAILED:
		newState = grpc_infrastructure_go.NodeState_UNREGISTERED
	}
	return newState
}
