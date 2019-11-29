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

package server

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-application-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
	"github.com/nalej/grpc-provisioner-go"
	"github.com/nalej/infrastructure-manager/internal/pkg/bus"
	"github.com/nalej/infrastructure-manager/internal/pkg/server/infrastructure"
	"github.com/nalej/nalej-bus/pkg/bus/pulsar-comcast"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

// Service structure with the configuration and the gRPC server.
type Service struct {
	Configuration Config
	Server        *grpc.Server
}

// NewService creates a new system model service.
func NewService(conf Config) *Service {
	return &Service{
		conf,
		grpc.NewServer(),
	}
}

// Clients structure with the gRPC clients for remote services.
type Clients struct {
	ClusterClient     grpc_infrastructure_go.ClustersClient
	NodesClient       grpc_infrastructure_go.NodesClient
	InstallerClient   grpc_installer_go.InstallerClient
	ProvisionerClient grpc_provisioner_go.ProvisionClient
	ScalerClient      grpc_provisioner_go.ScaleClient
	AppClient         grpc_application_go.ApplicationsClient
}

// GetClients creates the required connections with the remote clients.
func (s *Service) GetClients() (*Clients, derrors.Error) {
	smConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the system model")
	}
	insConn, err := grpc.Dial(s.Configuration.InstallerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the installer")
	}
	provConn, err := grpc.Dial(s.Configuration.ProvisionerAddress, grpc.WithInsecure())
	if err != nil {
		return nil, derrors.AsError(err, "cannot create connection with the provisioner")
	}
	cClient := grpc_infrastructure_go.NewClustersClient(smConn)
	nClient := grpc_infrastructure_go.NewNodesClient(smConn)
	iClient := grpc_installer_go.NewInstallerClient(insConn)
	pClient := grpc_provisioner_go.NewProvisionClient(provConn)
	scClient := grpc_provisioner_go.NewScaleClient(provConn)
	appClient := grpc_application_go.NewApplicationsClient(smConn)

	return &Clients{cClient, nClient, iClient,
		pClient, scClient, appClient}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() error {
	cErr := s.Configuration.Validate()
	if cErr != nil {
		log.Fatal().Str("err", cErr.DebugReport()).Msg("invalid configuration")
	}
	s.Configuration.Print()
	clients, cErr := s.GetClients()
	if cErr != nil {
		log.Fatal().Str("err", cErr.DebugReport()).Msg("Cannot create clients")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		log.Fatal().Errs("failed to listen: %v", []error{err})
	}

	// Create a bus manager
	log.Info().Msg("instantiate bus manager...")
	queueClient := pulsar_comcast.NewClient(s.Configuration.QueueAddress, nil)

	busManager, err := bus.NewBusManager(queueClient, "InfrastructureManager")
	if err != nil {
		log.Panic().Err(err).Msg("impossible to create bus manager instance")
		return err
	}
	log.Info().Msg("done")

	// Create handlers
	manager := infrastructure.NewManager(
		s.Configuration.TempDir,
		clients.ClusterClient, clients.NodesClient, clients.InstallerClient,
		clients.ProvisionerClient, clients.ScalerClient, clients.AppClient, busManager)
	handler := infrastructure.NewHandler(manager)

	grpc_infrastructure_manager_go.RegisterInfrastructureManagerServer(s.Server, handler)

	if s.Configuration.Debug {
		log.Info().Msg("Enabling gRPC server reflection")
		// Register reflection service on gRPC server.
		reflection.Register(s.Server)
	}

	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := s.Server.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
	return nil
}
