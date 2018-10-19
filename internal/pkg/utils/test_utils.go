/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package utils

import (
	"github.com/onsi/gomega"
	"google.golang.org/grpc"
	"os"
)

// RunIntegrationTests checks whether integration tests should be executed.
func RunIntegrationTests() bool {
	var runIntegration = os.Getenv("RUN_INTEGRATION_TEST")
	return runIntegration == "true"
}

func GetConnection(address string) (* grpc.ClientConn) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	gomega.Expect(err).To(gomega.Succeed())
	return conn
}