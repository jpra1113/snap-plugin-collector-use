/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/aasssddd/snap-plugin-lib-go/v1/plugin"
	"github.com/hyperpilotio/snap-plugin-collector-use/use"
	"google.golang.org/grpc"
)

const (
	pluginName     = "use"
	pluginVersion  = 1
	maxMessageSize = 100 << 20
)

// plugin bootstrap
func main() {
	plugin.StartCollector(
		use.NewUseCollector(),
		pluginName,
		pluginVersion,
		plugin.GRPCServerOptions(grpc.MaxMsgSize(maxMessageSize)),
		plugin.GRPCServerOptions(grpc.MaxSendMsgSize(maxMessageSize)),
		plugin.GRPCServerOptions(grpc.MaxRecvMsgSize(maxMessageSize)),
	)
}
