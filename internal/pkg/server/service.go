/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package server

import (
	"fmt"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-manager-go"
	"github.com/nalej/grpc-installer-go"
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
	Server * grpc.Server
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
	ClusterClient grpc_infrastructure_go.ClustersClient
	NodesClient grpc_infrastructure_go.NodesClient
	InstallerClient grpc_installer_go.InstallerClient
}

// GetClients creates the required connections with the remote clients.
func (s * Service) GetClients() (* Clients, derrors.Error) {
	smConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the system model")
	}
	insConn, err := grpc.Dial(s.Configuration.InstallerAddress, grpc.WithInsecure())
	if err != nil{
		return nil, derrors.AsError(err, "cannot create connection with the installer")
	}
	cClient := grpc_infrastructure_go.NewClustersClient(smConn)
	nClient := grpc_infrastructure_go.NewNodesClient(smConn)
	iClient := grpc_installer_go.NewInstallerClient(insConn)

	return &Clients{cClient, nClient, iClient}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() error {
	cErr := s.Configuration.Validate()
	if cErr != nil{
		log.Fatal().Str("err", cErr.DebugReport()).Msg("invalid configuration")
	}
	s.Configuration.Print()
	clients, cErr := s.GetClients()
	if cErr != nil{
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
	manager := infrastructure.NewManager(s.Configuration.TempDir, clients.ClusterClient, clients.NodesClient, clients.InstallerClient, busManager)
	handler := infrastructure.NewHandler(manager)

	grpc_infrastructure_manager_go.RegisterInfrastructureManagerServer(s.Server, handler)

	// Register reflection service on gRPC server.
	reflection.Register(s.Server)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := s.Server.Serve(lis); err != nil {
		log.Fatal().Errs("failed to serve: %v", []error{err})
	}
	return nil
}