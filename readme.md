# CLH Plugin SDK

CLH [(Cloudlog Helper)](https://github.com/SydneyOwl/cloudlog-helper) Plugin Go SDK is a Go SDK for creating CLH plugins. 
It provides basical functionality for communicating with CLH main program.

## Install
```
go get github.com/SydneyOwl/clh-plugin-go-sdk
```

## Quick start

Decode messages can now be configured per-plugin:
- `BATCHED` (default): packed decode frames.
- `REALTIME`: every decode in real-time.
- `BOTH`: receive both.

### Create a client
```go
package main

import (
	"github.com/SydneyOwl/clh-plugin-go-sdk"
	"github.com/davecgh/go-spew/spew"
	"log"
	"time"
)

func main() {
	// Configure plugin information
	config := pluginsdk.PluginConfig{
		Uuid:        "your-plugin-uuid",          // UUID must be unique across all plugins
		Name:        "My Plugin",                 // Plugin name
		Version:     "1.0.0",                     // Plugin version
		Description: "This is my awesome plugin", // Plugin description
		Capabilities: []pluginsdk.PluginCapability{
			pluginsdk.CapabilityWsjtxMessage, // Declare support for WSJT-X messages
			pluginsdk.CapabilityRigData,      // Declare support for radio data
			pluginsdk.CapabilityClhInternalMessage, // Declare support for internal message data
			pluginsdk.CapabilityPipeControl, // Enable control requests (server info, plugin list, etc.)
		},
		Metadata: map[string]string{
			"role": "example",
		},
	}

	// Create client instance (with optional parameters)
	client, err := pluginsdk.NewClient(
		config,
		pluginsdk.WithHeartbeatInterval(3*time.Second), // Custom heartbeat interval
		pluginsdk.WithRawDecodeDelivery(),              // Receive decode in realtime
		pluginsdk.WithWsjtxMessageFilter(
			pluginsdk.MessageType_DECODE,
			pluginsdk.MessageType_STATUS,
		),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	log.Println("Plugin successfully connected to CLH")
	_ = client.RequestServerInfo("req-server-info")

	// then you can receive messages by calling WaitMessage.
	// WaitMessage is blocking, you can receive and process messages in a goroutine.
	// If you call the Close method outside the goroutine, WaitMessage will immediately
	// exit and return an error.
	for{
		mmsg, err := client.WaitMessage()
		if err != nil {
			log.Fatalf("WaitMessage failed: %v", err)
		}

		// you can handle received messages here
		switch v := mmsg.(type) {
		case *pluginsdk.PipeConnectionClosed:
			break
			
		case *pluginsdk.WsjtxMessage:
			spew.Dump(v)

		case *pluginsdk.PackedDecodeMessage:
			spew.Dump(v)

		case *pluginsdk.RigData:
			spew.Dump(v)

		case *pluginsdk.PipeConnectionClosed:
			spew.Dump(v)

		case *pluginsdk.ClhInternalMessage:
			spew.Dump(v)

		case *pluginsdk.PipeControlResponse:
			spew.Dump(v)

		default:
			log.Printf("Unknown message type: %T", v)
		}
    }
}

```

