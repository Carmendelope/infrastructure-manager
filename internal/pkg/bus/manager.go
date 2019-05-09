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
)

// Structures and operators designed to manipulate the queue operations for the infrastructure ops queue.

type BusManager struct {
    producer *ops.InfrastructureOpsProducer
}

func NewBusManager(client bus.NalejClient, name string) (*BusManager, derrors.Error) {
    producer, err := ops.NewInfrastructureOpsProducer(client, name)
    if err != nil {
        return nil, err
    }
    return &BusManager{producer: producer}, nil
}

func (b BusManager) Send(ctx context.Context, msg proto.Message) derrors.Error {
    return b.producer.Send(ctx, msg)
}