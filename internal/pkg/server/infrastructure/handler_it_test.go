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

/*
RUN_INTEGRATION_TEST=true
IT_SM_ADDRESS=localhost:8800
IT_INSTALLER_ADDRESS=localhost:8900
IT_K8S_KUBECONFIG=/Users/daniel/.kube/config
*/

package infrastructure

import (
	"context"
	"fmt"
	grpc_application_go "github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-connectivity-manager-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/grpc-utils/pkg/test"
	"github.com/nalej/infrastructure-manager/internal/pkg/utils"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"io/ioutil"
	"os"
	"time"
)

func createOrganization(name string, orgClient grpc_organization_go.OrganizationsClient) *grpc_organization_go.Organization {
	toAdd := &grpc_organization_go.AddOrganizationRequest{
		Name: name,
	}
	added, err := orgClient.AddOrganization(context.Background(), toAdd)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(added).ToNot(gomega.BeNil())
	return added
}

// Check that a cluster exists and contains nodes.
func checkClusterAndNodes(organizationID string, clusterID string, systemModelAddress string) {
	smConn, err := grpc.Dial(systemModelAddress, grpc.WithInsecure())
	gomega.Expect(err).To(gomega.Succeed())
	clusterClient := grpc_infrastructure_go.NewClustersClient(smConn)
	cID := &grpc_infrastructure_go.ClusterId{
		OrganizationId: organizationID,
		ClusterId:      clusterID,
	}
	cluster, err := clusterClient.GetCluster(context.Background(), cID)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(cluster.ClusterStatus).Should(gomega.Equal(grpc_connectivity_manager_go.ClusterStatus_ONLINE))

	nodesClient := grpc_infrastructure_go.NewNodesClient(smConn)
	nodes, err := nodesClient.ListNodes(context.Background(), cID)
	gomega.Expect(err).To(gomega.Succeed())
	gomega.Expect(len(nodes.Nodes)).Should(gomega.Equal(1))
}

var _ = ginkgo.Describe("Infrastructure", func() {

	if !utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	var (
		systemModelAddress = os.Getenv("IT_SM_ADDRESS")
		installerAddress   = os.Getenv("IT_INSTALLER_ADDRESS")
		provisionerAddress = os.Getenv("IT_PROVISIONER_ADDRESS")
		kubeConfigFile     = os.Getenv("IT_K8S_KUBECONFIG")
	)

	if systemModelAddress == "" || kubeConfigFile == "" || installerAddress == "" {
		ginkgo.Fail("missing environment variables")
	}

	// gRPC server
	var server *grpc.Server
	// gRPC test listener
	var listener *bufconn.Listener

	// Clients
	var orgClient grpc_organization_go.OrganizationsClient
	var clusterClient grpc_infrastructure_go.ClustersClient
	var nodesClient grpc_infrastructure_go.NodesClient
	var installerClient grpc_installer_go.InstallerClient
	var provisionerClient grpc_provisioner_go.ProvisionClient
	var scaleClient grpc_provisioner_go.ScaleClient
	var appClient grpc_application_go.ApplicationsClient

	var smConn *grpc.ClientConn
	var instConn *grpc.ClientConn
	var provConn *grpc.ClientConn
	var client grpc_infrastructure_manager_go.InfrastructureManagerClient

	// Temp dir
	var tempDir string

	// Target organization.
	var targetOrganization *grpc_organization_go.Organization
	var kubeConfigRaw string

	const maxWait = 30

	ginkgo.BeforeSuite(func() {
		listener = test.GetDefaultListener()
		server = grpc.NewServer()

		smConn = utils.GetConnection(systemModelAddress)
		orgClient = grpc_organization_go.NewOrganizationsClient(smConn)
		clusterClient = grpc_infrastructure_go.NewClustersClient(smConn)
		nodesClient = grpc_infrastructure_go.NewNodesClient(smConn)
		instConn = utils.GetConnection(installerAddress)
		installerClient = grpc_installer_go.NewInstallerClient(instConn)
		provConn = utils.GetConnection(provisionerAddress)
		provisionerClient = grpc_provisioner_go.NewProvisionClient(provConn)
		scaleClient = grpc_provisioner_go.NewScaleClient(provConn)
		appClient = grpc_application_go.NewApplicationsClient(smConn)

		conn, err := test.GetConn(*listener)
		gomega.Expect(err).To(gomega.Succeed())

		manager := NewManager(tempDir, clusterClient, nodesClient, installerClient, provisionerClient, scaleClient, appClient, nil)
		handler := NewHandler(manager)
		grpc_infrastructure_manager_go.RegisterInfrastructureManagerServer(server, handler)
		test.LaunchServer(server, listener)

		client = grpc_infrastructure_manager_go.NewInfrastructureManagerClient(conn)
		targetOrganization = createOrganization(fmt.Sprintf("testOrg-%d", ginkgo.GinkgoRandomSeed()), orgClient)

		raw, err := ioutil.ReadFile(kubeConfigFile)
		gomega.Expect(err).To(gomega.Succeed())
		kubeConfigRaw = string(raw)

		created, err := ioutil.TempDir("", "handlerIT")
		gomega.Expect(err).To(gomega.Succeed())
		tempDir = created
	})

	ginkgo.AfterSuite(func() {
		server.Stop()
		listener.Close()
		smConn.Close()
		err := os.RemoveAll(tempDir)
		gomega.Expect(err).To(gomega.Succeed())
	})

	ginkgo.Context("with an existing kubernetes cluster", func() {
		ginkgo.It("should be able to install the nalej components", func() {

			installRequest := &grpc_installer_go.InstallRequest{
				OrganizationId:    targetOrganization.OrganizationId,
				ClusterId:         uuid.NewV4().String(),
				ClusterType:       grpc_infrastructure_go.ClusterType_KUBERNETES,
				InstallBaseSystem: false,
				KubeConfigRaw:     kubeConfigRaw,
				Username:          "",
				PrivateKey:        "",
				Nodes:             nil,
			}
			installResponse, err := client.InstallCluster(context.Background(), installRequest)
			gomega.Expect(err).To(gomega.Succeed())
			gomega.Expect(installResponse.RequestId).ShouldNot(gomega.BeEmpty())
			gomega.Expect(installResponse.Error).Should(gomega.BeEmpty())

			log.Debug().Interface("response", installResponse).Msg("Installation in progress")
			// Wait for the install to finish
			clusterID := &grpc_infrastructure_go.ClusterId{
				OrganizationId: targetOrganization.OrganizationId,
				ClusterId:      installRequest.ClusterId,
			}

			finished := false
			for i := 0; i < maxWait && !finished; i++ {
				updated, err := client.GetCluster(context.Background(), clusterID)
				gomega.Expect(err).To(gomega.Succeed())
				log.Debug().Interface("updated", updated).Msg("cluster status")
				finished = updated.ClusterStatus == grpc_connectivity_manager_go.ClusterStatus_ONLINE
				if !finished {
					time.Sleep(time.Second * 1)
				}
			}

			checkClusterAndNodes(targetOrganization.OrganizationId, installRequest.ClusterId, systemModelAddress)
		})
	})

	ginkgo.PContext("with a basic OS cluster", func() {
		ginkgo.PIt("should be able to install the nalej components", func() {

		})
	})

	ginkgo.PContext("with a single node cluster", func() {
		ginkgo.PIt("should be able to install the nalej components", func() {

		})
	})

})
