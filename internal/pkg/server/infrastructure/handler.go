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

package infrastructure

import (
	"context"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/infrastructure-manager/internal/pkg/entities"
	"github.com/satori/go.uuid"
)

type Handler struct {
	Manager Manager
}

func NewHandler(manager Manager) *Handler {
	return &Handler{manager}
}

// ProvisionAndInstallCluster provisions a new kubernetes cluster and then installs it
func (h *Handler) ProvisionAndInstallCluster(ctx context.Context, provisionRequest *grpc_provisioner_go.ProvisionClusterRequest) (*grpc_infrastructure_manager_go.ProvisionerResponse, error) {
	err := entities.ValidProvisionClusterRequest(provisionRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	provisionRequest.RequestId = uuid.NewV4().String()
	return h.Manager.ProvisionAndInstallCluster(provisionRequest)
}

// InstallCluster installs a new cluster into the system.
func (h *Handler) InstallCluster(ctx context.Context, installRequest *grpc_installer_go.InstallRequest) (*grpc_infrastructure_manager_go.InstallResponse, error) {
	err := entities.ValidInstallRequest(installRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	installRequest.InstallId = uuid.NewV4().String()
	return h.Manager.InstallCluster(installRequest)
}

// GetCluster retrieves the cluster information.
func (h *Handler) GetCluster(ctx context.Context, clusterID *grpc_infrastructure_go.ClusterId) (*grpc_infrastructure_go.Cluster, error) {
	err := entities.ValidClusterId(clusterID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.GetCluster(clusterID)
}

// ListClusters obtains a list of the clusters in the organization.
func (h *Handler) ListClusters(ctx context.Context, organizationID *grpc_organization_go.OrganizationId) (*grpc_infrastructure_go.ClusterList, error) {
	err := entities.ValidOrganizationId(organizationID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.ListClusters(organizationID)
}

// UpdateCluster allows the user to update the information of a cluster.
func (h *Handler) UpdateCluster(ctx context.Context, request *grpc_infrastructure_go.UpdateClusterRequest) (*grpc_infrastructure_go.Cluster, error) {
	err := entities.ValidUpdateClusterRequest(request)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.UpdateCluster(request)
}

// DrainCluster reschedules the services deployed in a given cluster.
func (h *Handler) DrainCluster(ctx context.Context, clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error) {
	err := entities.ValidClusterId(clusterID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.DrainCluster(clusterID)
}

// CordonCluster blocks the deployment of new services in a given cluster.
func (h *Handler) CordonCluster(ctx context.Context, clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error) {
	err := entities.ValidClusterId(clusterID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.CordonCluster(clusterID)
}

// UncordonCluster unblocks the deployment of new services in a given cluster.
func (h *Handler) UncordonCluster(ctx context.Context, clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error) {
	err := entities.ValidClusterId(clusterID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.UncordonCluster(clusterID)
}

// RemoveCluster removes a cluster from an organization. Notice that removing a cluster implies draining the cluster
// of running applications.
func (h *Handler) RemoveCluster(ctx context.Context, removeClusterRequest *grpc_infrastructure_go.RemoveClusterRequest) (*grpc_common_go.Success, error) {
	err := entities.ValidRemoveClusterRequest(removeClusterRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.RemoveCluster(removeClusterRequest)
}

// UpdateNode allows the user to update the information of a node.
func (h *Handler) UpdateNode(ctx context.Context, request *grpc_infrastructure_go.UpdateNodeRequest) (*grpc_infrastructure_go.Node, error) {
	err := entities.ValidUpdateNodeRequest(request)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.UpdateNode(request)
}

// ListNodes obtains a list of nodes in a cluster.
func (h *Handler) ListNodes(ctx context.Context, clusterID *grpc_infrastructure_go.ClusterId) (*grpc_infrastructure_go.NodeList, error) {
	err := entities.ValidClusterId(clusterID)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.ListNodes(clusterID)
}

// RemoveNodes removes a set of nodes from the system.
func (h *Handler) RemoveNodes(ctx context.Context, removeNodesRequest *grpc_infrastructure_go.RemoveNodesRequest) (*grpc_common_go.Success, error) {
	err := entities.ValidRemoveNodesRequest(removeNodesRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	return h.Manager.RemoveNodes(removeNodesRequest)
}
