/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package infrastructure

import (
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/infrastructure-manager/internal/pkg/entities"
	"github.com/nalej/infrastructure-manager/internal/pkg/monitor"
	"github.com/nalej/infrastructure-manager/internal/pkg/server/discovery/k8s"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
)

// Manager structure with the remote clients required to coordinate infrastructure operations.
type Manager struct {
	tempPath string
	clusterClient grpc_infrastructure_go.ClustersClient
	nodesClient grpc_infrastructure_go.NodesClient
	installerClient grpc_installer_go.InstallerClient
}

// NewManager creates a new manager.
func NewManager(
	tempDir string,
	clusterClient grpc_infrastructure_go.ClustersClient,
	nodesClient grpc_infrastructure_go.NodesClient,
	installerClient grpc_installer_go.InstallerClient) Manager {
	return Manager{
		tempPath: tempDir,
		clusterClient:   clusterClient,
		nodesClient:     nodesClient,
		installerClient: installerClient,
	}
}

// writeTempFile writes a content to a temporal file
func (m * Manager) writeTempFile(content string, prefix string) (*string, derrors.Error) {
	tmpfile, err := ioutil.TempFile(m.tempPath, prefix)
	if err != nil {
		return nil, derrors.AsError(err, "cannot create temporal file")
	}
	_, err = tmpfile.Write([]byte(content))
	if err != nil {
		return nil, derrors.AsError(err, "cannot write temporal file")
	}
	err = tmpfile.Close()
	if err != nil {
		return nil, derrors.AsError(err, "cannot close temporal file")
	}
	tmpName := tmpfile.Name()
	return &tmpName, nil
}

// addClusterToSM adds the newly discovered cluster to the system model.
func (m * Manager) addClusterToSM(installID string, organizationID string, cluster entities.Cluster) (*grpc_infrastructure_go.Cluster, derrors.Error) {
	toAdd := &grpc_infrastructure_go.AddClusterRequest{
		RequestId:            installID,
		OrganizationId:       organizationID,
		Name:                 cluster.Name,
		Description:          cluster.Description,
		Hostname:             cluster.Hostname,
		ControlPlaneHostname: cluster.ControlPlaneHostname,
	}
	log.Debug().Str("name", toAdd.Name).Msg("Adding cluster to SM")
	clusterAdded, err := m.clusterClient.AddCluster(context.Background(), toAdd)
	if err != nil {
		return nil, conversions.ToDerror(err)
	}
	for _, n := range cluster.Nodes{
		nodeToAdd := &grpc_infrastructure_go.AddNodeRequest{
			RequestId:            installID,
			OrganizationId:       organizationID,
			Ip:                   n.IP,
			Labels:               n.Labels,
		}
		log.Debug().Str("IP", nodeToAdd.Ip).Msg("Adding node to SM")
		addedNode, err := m.nodesClient.AddNode(context.Background(), nodeToAdd)
		if err != nil {
			return nil, conversions.ToDerror(err)
		}
		attachReq := &grpc_infrastructure_go.AttachNodeRequest{
			RequestId:            installID,
			OrganizationId:       organizationID,
			ClusterId:            clusterAdded.ClusterId,
			NodeId:               addedNode.NodeId,
		}
		log.Debug().Str("nodeId", attachReq.NodeId).Str("clusterID", attachReq.ClusterId).Msg("Attaching node to cluster")
		_, err = m.nodesClient.AttachNode(context.Background(), attachReq)
		if err != nil {
			return nil, conversions.ToDerror(err)
		}
	}
	return clusterAdded, nil
}

// discoverAndAddCluster triggers the discovery of an existing Kubernetes cluster using a KubeConfig file.
func (m * Manager) discoverAndAddCluster(installRequest *grpc_installer_go.InstallRequest) (*grpc_infrastructure_go.Cluster, derrors.Error){
	// Store the kubeconfig file in a temporal path.
	tempFile, err := m.writeTempFile(installRequest.KubeConfigRaw, installRequest.InstallId)
	defer os.Remove(*tempFile)
	if err != nil {
		return nil, err
	}
	dh := k8s.NewDiscoveryHelper(*tempFile)
	err = dh.Connect()
	if err != nil{
		return nil, err
	}
	discovered, err := dh.Discover()
	if err != nil {
		return nil, err
	}
	discovered.Hostname = installRequest.Hostname
	log.Debug().Str("KubernetesVersion", discovered.KubernetesVersion).
		Int("numNodes", len(discovered.Nodes)).
		Str("ControlPlaneHostname", discovered.ControlPlaneHostname).
		Str("hostname", discovered.Hostname).Msg("cluster has been discovered")
	// Add cluster and nodes to the system model.
	return m.addClusterToSM(installRequest.InstallId, installRequest.OrganizationId, *discovered)
}

// getOrCreateCluster retrieves the target cluster from system model, or triggers the discovery of an existing cluster depending
// on the request parameters.
func (m * Manager) getOrCreateCluster(installRequest *grpc_installer_go.InstallRequest) (*grpc_infrastructure_go.Cluster, derrors.Error){
	var result *grpc_infrastructure_go.Cluster
	if installRequest.ClusterId == "" {
		log.Debug().Str("installId", installRequest.InstallId).Msg("Discovering cluster")
		// Discover cluster
		added, err := m.discoverAndAddCluster(installRequest)
		if err != nil{
			return nil, err
		}
		result = added
	}else{
		log.Debug().Str("installId", installRequest.InstallId).Str("clusterID", installRequest.ClusterId).Msg("Retrieving existing cluster")
		clusterID := &grpc_infrastructure_go.ClusterId{
			OrganizationId:       installRequest.OrganizationId,
			ClusterId:            installRequest.ClusterId,
		}
		retrieved, err := m.clusterClient.GetCluster(context.Background(), clusterID)
		if err != nil {
			return nil, conversions.ToDerror(err)
		}
		result = retrieved
	}
	if result == nil {
		return nil, derrors.NewInternalError("cannot discover or get existing cluster")
	}
	log.Debug().Str("clusterID", result.ClusterId).Msg("target cluster found")
	return result, nil
}

func (m * Manager) InstallCluster(installRequest *grpc_installer_go.InstallRequest) (*grpc_infrastructure_manager_go.InstallResponse, error) {
	cluster, err := m.getOrCreateCluster(installRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	if installRequest.InstallBaseSystem {
		return nil, derrors.NewUnimplementedError("InstallBaseSystem not supported")
	}
	installRequest.ClusterId = cluster.ClusterId

	log.Debug().Str("clusterID", installRequest.ClusterId).Msg("installing cluster")
	installerResponse, iErr := m.installerClient.InstallCluster(context.Background(), installRequest)
	if iErr != nil {
		return nil, iErr
	}
	log.Debug().Interface("state", installerResponse.State).Msg("cluster is being installed")
	installResponse := &grpc_infrastructure_manager_go.InstallResponse{
		InstallId:            installerResponse.InstallId,
		OrganizationId:       installRequest.OrganizationId,
		ClusterId:            installRequest.ClusterId,
		State:                installerResponse.State,
		Error:                installerResponse.Error,
	}
	mon := monitor.NewMonitor(m.installerClient, m.clusterClient, *installResponse)
	mon.RegisterCallback(m.installCallback)
	go mon.LaunchMonitor()
	return installResponse, nil
}

func (m * Manager) installCallback(
		installID string, organizationID string, clusterID string,
		lastResponse * grpc_installer_go.InstallResponse, err derrors.Error){
	log.Debug().Str("installID", installID).Msg("callback received")
	if err != nil {
		log.Error().Str("err", err.DebugReport()).Msg("error callback received")
	}
	if lastResponse == nil {
		return
	}

	var newStatus = entities.StateToStatus(lastResponse.State)
	updateClusterRequest := &grpc_infrastructure_go.UpdateClusterRequest{
		OrganizationId:       organizationID,
		ClusterId:            clusterID,
		UpdateStatus:         true,
		Status:               newStatus,
	}
	_, cErr := m.clusterClient.UpdateCluster(context.Background(), updateClusterRequest)
	if cErr != nil {
		log.Error().Str("err", conversions.ToDerror(err).DebugReport()).Msg("cannot update system model")
	}
}

// GetCluster retrieves the cluster information.
func (m * Manager) GetCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_infrastructure_go.Cluster, error){
	return m.clusterClient.GetCluster(context.Background(), clusterID)
}

// ListClusters obtains a list of the clusters in the organization.
func (m * Manager) ListClusters(organizationID *grpc_organization_go.OrganizationId) (*grpc_infrastructure_go.ClusterList, error){
	return m.clusterClient.ListClusters(context.Background(), organizationID)
}

// DrainCluster reschedules the services deployed in a given cluster.
func (m * Manager) DrainCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error){
	return nil, derrors.NewUnimplementedError("DrainCluster is not implemented yet")
}

// CordonCluster blocks the deployment of new services in a given cluster.
func (m * Manager) CordonCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error){
	return nil, derrors.NewUnimplementedError("CordonCluster is not implemented yet")
}

// RemoveCluster removes a cluster from an organization. Notice that removing a cluster implies draining the cluster
// of running applications.
func (m * Manager) RemoveCluster(removeClusterRequest *grpc_infrastructure_go.RemoveClusterRequest) (*grpc_common_go.Success, error){
	return nil, derrors.NewUnimplementedError("RemoveCluster is not implemented yet")
}

// ListNodes obtains a list of nodes in a cluster.
func (m * Manager) ListNodes(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_infrastructure_go.NodeList, error){
	return m.nodesClient.ListNodes(context.Background(), clusterID)
}

// RemoveNodes removes a set of nodes from the system.
func (m * Manager) RemoveNodes(removeNodesRequest *grpc_infrastructure_go.RemoveNodesRequest) (*grpc_common_go.Success, error){
	return nil, derrors.NewUnimplementedError("RemoveNodes is not implemented yet")
}