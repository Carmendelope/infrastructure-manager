/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-conductor-go"
	"github.com/nalej/grpc-connectivity-manager-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
	"github.com/nalej/infrastructure-manager/internal/pkg/bus"
	"github.com/nalej/infrastructure-manager/internal/pkg/entities"
	"github.com/nalej/infrastructure-manager/internal/pkg/monitor"
	"github.com/nalej/infrastructure-manager/internal/pkg/server/discovery/k8s"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"os"
	"time"
)

const (
	// Default timeout
	DefaultTimeout = 2 * time.Minute
	// Standard timeout for operations done in this manager
	InfrastructureManagerTimeout = time.Second * 5
)

// Manager structure with the remote clients required to coordinate infrastructure operations.
type Manager struct {
	tempPath          string
	clusterClient     grpc_infrastructure_go.ClustersClient
	nodesClient       grpc_infrastructure_go.NodesClient
	installerClient   grpc_installer_go.InstallerClient
	provisionerClient grpc_provisioner_go.ProvisionClient
	busManager        *bus.BusManager
}

// NewManager creates a new manager.
func NewManager(
	tempDir string,
	clusterClient grpc_infrastructure_go.ClustersClient,
	nodesClient grpc_infrastructure_go.NodesClient,
	installerClient grpc_installer_go.InstallerClient,
	provisionerClient grpc_provisioner_go.ProvisionClient,
	busManager *bus.BusManager) Manager {
	return Manager{
		tempPath:          tempDir,
		clusterClient:     clusterClient,
		nodesClient:       nodesClient,
		installerClient:   installerClient,
		provisionerClient: provisionerClient,
		busManager:        busManager,
	}
}

// writeTempFile writes a content to a temporal file
func (m *Manager) writeTempFile(content string, prefix string) (*string, derrors.Error) {
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
func (m *Manager) addClusterToSM(requestID string, organizationID string, cluster entities.Cluster, clusterState grpc_infrastructure_go.ClusterState) (*grpc_infrastructure_go.Cluster, derrors.Error) {
	toAdd := &grpc_infrastructure_go.AddClusterRequest{
		RequestId:            requestID,
		OrganizationId:       organizationID,
		Name:                 cluster.Name,
		Hostname:             cluster.Hostname,
		ControlPlaneHostname: cluster.ControlPlaneHostname,
	}
	log.Debug().Str("name", toAdd.Name).Msg("Adding cluster to SM")
	clusterAdded, err := m.clusterClient.AddCluster(context.Background(), toAdd)
	if err != nil {
		return nil, conversions.ToDerror(err)
	}
	err = m.updateClusterState(organizationID, clusterAdded.ClusterId, clusterState)

	for _, n := range cluster.Nodes {
		nodeToAdd := &grpc_infrastructure_go.AddNodeRequest{
			RequestId:      requestID,
			OrganizationId: organizationID,
			Ip:             n.IP,
			Labels:         n.Labels,
		}
		log.Debug().Str("IP", nodeToAdd.Ip).Msg("Adding node to SM")
		addedNode, err := m.nodesClient.AddNode(context.Background(), nodeToAdd)
		if err != nil {
			return nil, conversions.ToDerror(err)
		}
		attachReq := &grpc_infrastructure_go.AttachNodeRequest{
			RequestId:      requestID,
			OrganizationId: organizationID,
			ClusterId:      clusterAdded.ClusterId,
			NodeId:         addedNode.NodeId,
		}
		log.Debug().Str("nodeId", attachReq.NodeId).Str("clusterID", attachReq.ClusterId).Msg("Attaching node to cluster")
		_, err = m.nodesClient.AttachNode(context.Background(), attachReq)
		if err != nil {
			return nil, conversions.ToDerror(err)
		}
	}
	return clusterAdded, nil
}

func (m *Manager) discoverCluster(requestID string, kubeConfig string, hostname string) (*entities.Cluster, derrors.Error) {
	// Store the kubeconfig file in a temporal path.
	tempFile, err := m.writeTempFile(kubeConfig, requestID)
	defer os.Remove(*tempFile)
	if err != nil {
		return nil, err
	}
	dh := k8s.NewDiscoveryHelper(*tempFile)
	err = dh.Connect()
	if err != nil {
		return nil, err
	}
	discovered, err := dh.Discover()
	if err != nil {
		return nil, err
	}
	discovered.Hostname = hostname
	log.Debug().Str("KubernetesVersion", discovered.KubernetesVersion).
		Int("numNodes", len(discovered.Nodes)).
		Str("ControlPlaneHostname", discovered.ControlPlaneHostname).
		Str("hostname", discovered.Hostname).Msg("cluster has been discovered")

	return discovered, nil
}

// getOrCreateProvisionedCluster retrieves the target cluster from system model, or triggers the discovery of an existing cluster depending
// on the request parameters.
func (m *Manager) getOrCreateProvisionedCluster(installRequest *grpc_installer_go.InstallRequest) (*grpc_infrastructure_go.Cluster, derrors.Error) {
	var result *grpc_infrastructure_go.Cluster
	if installRequest.ClusterId == "" {
		log.Debug().Str("requestID", installRequest.InstallId).Msg("Discovering cluster")
		// Discover cluster
		discovered, err := m.discoverCluster(installRequest.InstallId, installRequest.KubeConfigRaw, installRequest.Hostname)
		if err != nil {
			return nil, err
		}
		added, err := m.addClusterToSM(installRequest.InstallId, installRequest.OrganizationId, *discovered, grpc_infrastructure_go.ClusterState_PROVISIONED)
		if err != nil {
			return nil, err
		}
		result = added
	} else {
		log.Debug().Str("requestID", installRequest.InstallId).Str("clusterID", installRequest.ClusterId).Msg("Retrieving existing cluster")
		clusterID := &grpc_infrastructure_go.ClusterId{
			OrganizationId: installRequest.OrganizationId,
			ClusterId:      installRequest.ClusterId,
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

// UpdateClusterState updates the state of a cluster in system model. The update is also sent to the bus
// so that other components of the system can react to events such as new cluster becoming available.
func (m *Manager) updateClusterState(organizationID string, clusterID string, newState grpc_infrastructure_go.ClusterState) derrors.Error {
	updateRequest := &grpc_infrastructure_go.UpdateClusterRequest{
		OrganizationId:     organizationID,
		ClusterId:          clusterID,
		UpdateClusterState: true,
		State:              newState,
	}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()
	_, err := m.clusterClient.UpdateCluster(ctx, updateRequest)
	if err != nil {
		return derrors.AsError(err, "cannot update cluster state")
	}

	// if correct send it to the bus
	ctxBus, cancelBus := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancelBus()
	errBus := m.busManager.SendEvents(ctxBus, updateRequest)
	if errBus != nil {
		log.Error().Err(errBus).Msg("error in the bus when sending an update cluster request")
		return errBus
	}
	return nil
}

func (m *Manager) ProvisionAndInstallCluster(provisionRequest *grpc_provisioner_go.ProvisionClusterRequest) (*grpc_infrastructure_manager_go.ProvisionerResponse, error) {
	log.Debug().Interface("request", provisionRequest).Msg("ProvisionAndInstallCluster")
	log.Debug().Str("platform", provisionRequest.TargetPlatform.String()).Msg("Target platform")
	log.Debug().Str("cluster_name", provisionRequest.ClusterName).Msg("Cluster name")

	cluster, err := m.addClusterToSM(provisionRequest.RequestId, provisionRequest.OrganizationId, entities.Cluster{}, grpc_infrastructure_go.ClusterState_PROVISIONING)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	provisionRequest.ClusterId = cluster.ClusterId

	log.Debug().Str("clusterID", provisionRequest.ClusterId).Msg("provisioning cluster")
	provisionerResponse, pErr := m.provisionerClient.ProvisionCluster(context.Background(), provisionRequest)
	if pErr != nil {
		return nil, pErr
	}
	log.Debug().Str("clusterID", provisionRequest.ClusterId).Msg("cluster is being provisioned")
	provisionResponse := &grpc_infrastructure_manager_go.ProvisionerResponse{
		RequestId:      provisionerResponse.RequestId,
		OrganizationId: provisionRequest.OrganizationId,
		ClusterId:      provisionRequest.ClusterId,
		State:          provisionerResponse.State,
		Error:          provisionerResponse.Error,
	}
	mon := monitor.NewProvisionerMonitor(m.provisionerClient, m.clusterClient, *provisionResponse)
	mon.RegisterCallback(m.provisionCallback)
	go mon.LaunchMonitor()
	return provisionResponse, nil
}

func (m *Manager) provisionCallback(requestID string, organizationID string, clusterID string,
	lastResponse *grpc_provisioner_go.ProvisionClusterResponse, err derrors.Error) {
	log.Debug().Str("requestID", requestID).Msg("provisioner callback received")
	if err != nil {
		log.Error().Str("err", err.DebugReport()).Msg("error callback received")
	}
	if lastResponse == nil {
		return
	}

	newState := grpc_infrastructure_go.ClusterState_PROVISIONED
	if err != nil || lastResponse.State == grpc_provisioner_go.ProvisionProgress_ERROR {
		newState = grpc_infrastructure_go.ClusterState_FAILURE
		log.Warn().Str("requestID", requestID).Str("organizationID", organizationID).Str("clusterID", clusterID).Msg("Provision failed")
	}
	err = m.updateClusterState(organizationID, clusterID, newState)
	if err != nil {
		log.Error().Msg("unable to update cluster state after provision")
	}

	discovered, err := m.discoverCluster(requestID, lastResponse.RawKubeConfig, lastResponse.Hostname)
	if err != nil {
		log.Error().Msg("unable to discover cluster")
	}
	clusterUpdate := &grpc_infrastructure_go.UpdateClusterRequest{
		OrganizationId:             organizationID,
		ClusterId:                  clusterID,
		UpdateHostname:             true,
		Hostname:                   lastResponse.Hostname,
		UpdateControlPlaneHostname: true,
		ControlPlaneHostname:       discovered.ControlPlaneHostname,
	}
	_, updErr := m.UpdateCluster(clusterUpdate)
	if updErr != nil {
		derrors.NewInternalError("error updating discovered cluster", updErr)
	}

	installRequest := &grpc_installer_go.InstallRequest{
		InstallId:         requestID,
		OrganizationId:    organizationID,
		ClusterId:         clusterID,
		ClusterType:       grpc_infrastructure_go.ClusterType_KUBERNETES,
		InstallBaseSystem: false,
		KubeConfigRaw:     lastResponse.RawKubeConfig,
		Hostname:          lastResponse.Hostname,
		TargetPlatform:    grpc_installer_go.Platform_AZURE,
		StaticIpAddresses: lastResponse.StaticIpAddresses,
	}
	_, icErr := m.InstallCluster(installRequest)
	if icErr != nil {
		derrors.NewInternalError("error creating install request after provision", icErr)
	}
	return
}

func (m *Manager) InstallCluster(installRequest *grpc_installer_go.InstallRequest) (*grpc_infrastructure_manager_go.InstallResponse, error) {
	log.Debug().Interface("request", installRequest).Msg("InstallCluster")
	log.Debug().Str("platform", installRequest.TargetPlatform.String()).Msg("Target platform")
	log.Debug().Str("hostname", installRequest.Hostname).Msg("Public App cluster hostname")
	cluster, err := m.getOrCreateProvisionedCluster(installRequest)
	if err != nil {
		return nil, conversions.ToGRPCError(err)
	}
	if installRequest.InstallBaseSystem {
		return nil, derrors.NewUnimplementedError("InstallBaseSystem not supported")
	}
	installRequest.ClusterId = cluster.ClusterId
	if cluster.State != grpc_infrastructure_go.ClusterState_PROVISIONED {
		return nil, derrors.NewInvalidArgumentError("selected cluster is not ready for install")
	}
	err = m.updateClusterState(installRequest.OrganizationId, installRequest.ClusterId, grpc_infrastructure_go.ClusterState_INSTALL_IN_PROGRESS)
	if err != nil {
		log.Error().Str("trace", err.DebugReport()).Msg("cannot update cluster state")
		return nil, err
	}
	log.Debug().Str("clusterID", installRequest.ClusterId).Msg("installing cluster")
	installerResponse, iErr := m.installerClient.InstallCluster(context.Background(), installRequest)
	if iErr != nil {
		return nil, iErr
	}
	log.Debug().Interface("state", installerResponse.State).Msg("cluster is being installed")
	installResponse := &grpc_infrastructure_manager_go.InstallResponse{
		InstallId:      installerResponse.InstallId,
		OrganizationId: installRequest.OrganizationId,
		ClusterId:      installRequest.ClusterId,
		State:          installerResponse.State,
		Error:          installerResponse.Error,
	}
	mon := monitor.NewInstallerMonitor(m.installerClient, m.clusterClient, *installResponse)
	mon.RegisterCallback(m.installCallback)
	go mon.LaunchMonitor()
	return installResponse, nil
}

func (m *Manager) installCallback(
	installID string, organizationID string, clusterID string,
	lastResponse *grpc_installer_go.InstallResponse, err derrors.Error) {
	log.Debug().Str("installID", installID).Msg("installer callback received")
	if err != nil {
		log.Error().Str("err", err.DebugReport()).Msg("error callback received")
	}
	if lastResponse == nil {
		return
	}

	newState := grpc_infrastructure_go.ClusterState_INSTALLED
	if err != nil || lastResponse.State == grpc_installer_go.InstallProgress_ERROR {
		newState = grpc_infrastructure_go.ClusterState_FAILURE
		log.Warn().Str("installID", installID).Str("organizationID", organizationID).Str("clusterID", clusterID).Msg("installation failed")
	}
	err = m.updateClusterState(organizationID, clusterID, newState)
	if err != nil {
		log.Error().Msg("unable to update cluster state after install")
	}
	// TODO Refactor to update nodes accordingly.
	// Get the list of nodes
	cID := &grpc_infrastructure_go.ClusterId{
		OrganizationId: organizationID,
		ClusterId:      clusterID,
	}
	nodes, nErr := m.nodesClient.ListNodes(context.Background(), cID)
	if err != nil {
		log.Error().Str("err", conversions.ToDerror(nErr).DebugReport()).Msg("cannot obtain the list of nodes in the cluster")
		return
	}

	for _, n := range nodes.Nodes {
		updateNodeRequest := &grpc_infrastructure_go.UpdateNodeRequest{
			OrganizationId: organizationID,
			NodeId:         n.NodeId,
			UpdateStatus:   true,
			Status:         n.Status,
			UpdateState:    true,
			State:          entities.InstallStateToNodeState(lastResponse.State),
		}
		_, updateErr := m.nodesClient.UpdateNode(context.Background(), updateNodeRequest)
		if updateErr != nil {
			log.Error().Str("err", conversions.ToDerror(updateErr).DebugReport()).Msg("cannot update the node status")
			return
		}
		log.Debug().Str("organizationID", organizationID).Str("nodeId", n.NodeId).Interface("newStatus", n.Status).Msg("Node status updated")
	}
}

// GetCluster retrieves the cluster information.
func (m *Manager) GetCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_infrastructure_go.Cluster, error) {
	return m.clusterClient.GetCluster(context.Background(), clusterID)
}

// ListClusters obtains a list of the clusters in the organization.
func (m *Manager) ListClusters(organizationID *grpc_organization_go.OrganizationId) (*grpc_infrastructure_go.ClusterList, error) {
	return m.clusterClient.ListClusters(context.Background(), organizationID)
}

// UpdateCluster allows the user to update the information of a cluster.
func (m *Manager) UpdateCluster(request *grpc_infrastructure_go.UpdateClusterRequest) (*grpc_infrastructure_go.Cluster, error) {
	// update system model
	ctx, cancel := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancel()

	updateResult, updateErr := m.clusterClient.UpdateCluster(ctx, request)
	if updateErr != nil {
		return nil, updateErr
	}
	// if correct send it to the bus
	ctxBus, cancelBus := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancelBus()
	errBus := m.busManager.SendEvents(ctxBus, request)
	if errBus != nil {
		log.Error().Err(errBus).Msg("error in the bus when sending an update cluster request")
		return nil, errBus
	}
	return updateResult, nil

}

// DrainCluster reschedules the services deployed in a given cluster.
func (m *Manager) DrainCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error) {
	// Check this cluster is cordoned
	ctx, cancel := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancel()
	targetCluster, err := m.clusterClient.GetCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("status", targetCluster.ClusterStatus.String()).Msg("cluster status")
	if targetCluster.ClusterStatus != grpc_connectivity_manager_go.ClusterStatus_OFFLINE_CORDON && targetCluster.ClusterStatus != grpc_connectivity_manager_go.ClusterStatus_ONLINE_CORDON {
		err := errors.New(fmt.Sprintf("cluster %s must be cordoned before draining", targetCluster.ClusterId))
		return nil, err
	}

	// send drain operation to the common bus
	ctxDrain, cancelDrain := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancelDrain()
	msg := &grpc_conductor_go.DrainClusterRequest{ClusterId: clusterID}
	err = m.busManager.SendOps(ctxDrain, msg)
	if err != nil {
		log.Error().Err(err).Msg("error in the bus when sending a drain cluster request")
		return nil, err
	}

	return &grpc_common_go.Success{}, nil
}

// CordonCluster blocks the deployment of new services in a given cluster.
func (m *Manager) CordonCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error) {
	ctx, cancel := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancel()
	succ, err := m.clusterClient.CordonCluster(ctx, clusterID)
	ctxe, cancele := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancele()
	errBus := m.busManager.SendEvents(ctxe, &grpc_infrastructure_go.SetClusterStatusRequest{ClusterId: clusterID, Cordon: true})
	if errBus != nil {
		log.Error().Err(errBus).Msg("error sending set cluster request to queue")
	}
	return succ, err
}

// CordonCluster unblocks the deployment of new services in a given cluster.
func (m *Manager) UncordonCluster(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_common_go.Success, error) {
	ctx, cancel := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancel()
	succ, err := m.clusterClient.UncordonCluster(ctx, clusterID)
	ctxe, cancele := context.WithTimeout(context.Background(), InfrastructureManagerTimeout)
	defer cancele()
	errBus := m.busManager.SendEvents(ctxe, &grpc_infrastructure_go.SetClusterStatusRequest{ClusterId: clusterID, Cordon: false})
	if errBus != nil {
		log.Error().Err(errBus).Msg("error sending set cluster request to queue")
	}
	return succ, err
}

// RemoveCluster removes a cluster from an organization. Notice that removing a cluster implies draining the cluster
// of running applications.
func (m *Manager) RemoveCluster(removeClusterRequest *grpc_infrastructure_go.RemoveClusterRequest) (*grpc_common_go.Success, error) {
	return nil, derrors.NewUnimplementedError("RemoveCluster is not implemented yet")
}

// UpdateNode allows the user to update the information of a node.
func (m *Manager) UpdateNode(request *grpc_infrastructure_go.UpdateNodeRequest) (*grpc_infrastructure_go.Node, error) {
	updated, err := m.nodesClient.UpdateNode(context.Background(), request)
	if err != nil {
		return nil, err
	}
	// TODO Update the labels in Kubernetes. A new proto should be added in the app cluster api to pass that information
	log.Warn().Str("organizationId", updated.OrganizationId).
		Str("nodeId", updated.NodeId).
		Str("clusterId", updated.ClusterId).
		Msg("node labels have not been updated in kubernetes")
	return updated, err
}

// ListNodes obtains a list of nodes in a cluster.
func (m *Manager) ListNodes(clusterID *grpc_infrastructure_go.ClusterId) (*grpc_infrastructure_go.NodeList, error) {
	return m.nodesClient.ListNodes(context.Background(), clusterID)
}

// RemoveNodes removes a set of nodes from the system.
func (m *Manager) RemoveNodes(removeNodesRequest *grpc_infrastructure_go.RemoveNodesRequest) (*grpc_common_go.Success, error) {
	return nil, derrors.NewUnimplementedError("RemoveNodes is not implemented yet")
}
