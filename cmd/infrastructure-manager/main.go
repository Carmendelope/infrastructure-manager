/*
 * Copyright (C) 2018 Nalej - All Rights Reserved
 */

// This is an example of an executable command.

package main

import (
	"github.com/nalej/infrastructure-manager/cmd/infrastructure-manager/commands"
	"github.com/nalej/infrastructure-manager/version"
)

var MainVersion string

var MainCommit string

func main() {
	version.AppVersion = MainVersion
	version.Commit = MainCommit
	commands.Execute()
}
