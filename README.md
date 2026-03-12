# CLH Plugin Go SDK (v20260309)

Cloudlog Helper plugin SDK for Go, built on `clh-proto/gen/go/v20260312`.

## Features

- Option-based configuration
- Configurable heartbeat interval (`WithHeartbeatInterval`)
- Message callback (`WithMessageHandler`)
- Blocking receive (`WaitMessage(ctx)`)
- Event stream is envelope-only (`PipeEnvelope` with typed `Payload`)
- Clean structs only in public API (no protobuf structs returned)
- Full plugin control/query wrapper methods for current protocol topics

## Quick Start

```go
package main

import (
	"context"
	"log"
	"time"

	sdk "github.com/SydneyOwl/clh-plugin-go-sdk"
)

func main() {
	client, err := sdk.NewClient(
		sdk.PluginManifest{
			UUID:    "your-plugin-uuid",
			Name:    "demo-plugin",
			Version: "1.0.0",
		},
		sdk.WithHeartbeatInterval(3*time.Second),
		sdk.WithMessageHandler(func(msg sdk.Message) {
			log.Printf("callback kind=%s", msg.Kind)
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	reg, err := client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("connected to instance=%s", reg.InstanceID)

	// Blocking receive
	go func() {
		for {
			msg, waitErr := client.WaitMessage(ctx)
			if waitErr != nil {
				return
			}
			log.Printf("wait message kind=%s", msg.Kind)
		}
	}()

	info, err := client.QueryServerInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("server version=%s", info.Version)

	_ = client.Close(context.Background())
}
```

## Supported Methods

### Query

- `QueryServerInfo`
- `QueryConnectedPlugins`
- `QueryRuntimeSnapshot`
- `QueryRigSnapshot`
- `QueryUDPSnapshot`
- `QueryQSOQueueSnapshot`
- `QuerySettingsSnapshot`
- `QueryPluginTelemetry`

### Command

- `SubscribeEvents`
- `ShowMainWindow`
- `HideMainWindow`
- `OpenWindow`
- `SendNotification`
- `ToggleUDPServer`
- `StartRigBackend`
- `StopRigBackend`
- `RestartRigBackend`
- `TriggerQSOReupload`
- `UpdateSettings`

### Generic

- `RawQuery`
- `RawCommand`
