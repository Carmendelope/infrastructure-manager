/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package bus

import (
    "context"
    "github.com/golang/protobuf/proto"
    "github.com/nalej/derrors"
    "github.com/nalej/nalej-bus/pkg/bus"
    "github.com/nalej/nalej-bus/pkg/queue/infrastructure/ops"
    "github.com/nalej/nalej-bus/pkg/queue/infrastructure/events"
)

// Structures and operators designed to manipulate the queue operations for the infrastructure ops queue.

type BusManager struct {
    producerOps *ops.InfrastructureOpsProducer
    producerEvents *events.InfrastructureEventsProducer
}

func NewBusManager(client bus.NalejClient, name string) (*BusManager, derrors.Error) {
    producerOps, err := ops.NewInfrastructureOpsProducer(client, name)
    if err != nil {
        return nil, err
    }
    producerEvents, err := events.NewInfrastructureEventsProducer(client,name)
    if err != nil {
        return nil, err
    }
    return &BusManager{producerOps: producerOps, producerEvents: producerEvents}, nil
}

// Send a new operation
func (b BusManager) SendOps(ctx context.Context, msg proto.Message) derrors.Error {
    return b.producerOps.Send(ctx, msg)
}

// Send a new event
func (b BusManager) SendEvents(ctx context.Context, msg proto.Message) derrors.Error {
    return b.producerEvents.Send(ctx, msg)
}