/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

package infrastructure

import (
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"testing"
)

func TestInfrastructurePackage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Infrastructure package suite")
}