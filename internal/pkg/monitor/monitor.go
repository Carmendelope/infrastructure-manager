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
