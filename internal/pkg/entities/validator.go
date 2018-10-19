/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package entities

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-organization-go"
)

const emptyRequestId = "request_id cannot be empty"
const emptyOrganizationId = "organization_id cannot be empty"
const emptyClusterId = "cluster_id cannot be empty"
const emptyEmail = "email cannot be empty"


func ValidOrganizationId(organizationID *grpc_organization_go.OrganizationId) derrors.Error {
	if organizationID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	return nil
}

func ValidClusterId(clusterID *grpc_infrastructure_go.ClusterId) derrors.Error {
	if clusterID.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if clusterID.ClusterId == "" {
		return derrors.NewInvalidArgumentError(emptyClusterId)
	}
	return nil
}

func ValidInstallRequest(installRequest *grpc_installer_go.InstallRequest) derrors.Error {
	if installRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if installRequest.ClusterId != "" {
		return derrors.NewInvalidArgumentError("cluster_id must be nil, and set by this component")
	}
	if installRequest.InstallId !=  ""{
		return derrors.NewInvalidArgumentError("install_id must be nil, and set by this component")
	}

	authFound := false

	if installRequest.Username != "" {
		if installRequest.PrivateKey == "" {
			return derrors.NewInvalidArgumentError("expecting PrivateKey with Username")
		}
		if len(installRequest.Nodes) == 0 {
			return derrors.NewInvalidArgumentError("expecting Nodes with Username")
		}
		authFound = true
	}
	if installRequest.KubeConfigRaw != "" {
		if installRequest.Username != "" {
			return derrors.NewInvalidArgumentError("expecting KubeConfigRaw without Username")
		}
		if installRequest.PrivateKey != "" {
			return derrors.NewInvalidArgumentError("expecting KubeConfigRaw without PrivateKey")
		}
		if len(installRequest.Nodes) > 0 {
			return derrors.NewInvalidArgumentError("expecting KubeConfigRaw without Nodes")
		}
		authFound = true
	}
	if ! authFound {
		return derrors.NewInvalidArgumentError("expecting KubeConfigRaw or Username, PrivateKey and Nodes")
	}
	return nil

}

func ValidRemoveClusterRequest(removeClusterRequest *grpc_infrastructure_go.RemoveClusterRequest) derrors.Error {
	return derrors.NewUnimplementedError("ValidRemoveClusterRequest")
}

func ValidRemoveNodesRequest(removeNodesRequest *grpc_infrastructure_go.RemoveNodesRequest) derrors.Error {
	if removeNodesRequest.RequestId == "" {
		return derrors.NewInvalidArgumentError(emptyRequestId)
	}
	if removeNodesRequest.OrganizationId == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if len(removeNodesRequest.Nodes) == 0 {
		return derrors.NewInvalidArgumentError("nodes must not be empty")
	}
	return nil
}


