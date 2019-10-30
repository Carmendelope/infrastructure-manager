package monitor

import (
	"time"
)

// MaxConnFailures contains the number of communication failures allowed until the monitor exists.
const (
	MaxConnFailures = 5
	DefaultTimeout  = 2 * time.Minute
)

// QueryDelay contains the polling interval to check the progress on the provisioner.
const QueryDelay = time.Second * 15

// ConnectRetryDelay contains the polling interval to retry the connection with the provisioner
const ConnectRetryDelay = time.Second * 30
