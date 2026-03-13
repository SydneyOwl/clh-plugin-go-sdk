# CLH Plugin Go SDK

Cloudlog Helper plugin SDK for Go, built on `clh-proto/gen/go/v20260312`.

## Install

```bash
go get github.com/SydneyOwl/clh-plugin-go-sdk
```

## What this SDK provides

- Plugin registration / heartbeat / graceful deregistration
- Typed query & command wrappers for current CLH plugin protocol topics
- Dual inbound mode: callback (`WithMessageHandler`) + blocking read (`WaitMessage`)
- Public API returns clean Go structs (no protobuf structs exposed)
- `RawQuery` / `RawCommand` for advanced custom requests

## Quick start

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
		sdk.WithRequestTimeout(8*time.Second),
		sdk.WithMessageHandler(func(msg sdk.Message) {
			log.Printf("callback message: kind=%s", msg.Kind)
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
	log.Printf("connected: instance=%s version=%s", reg.InstanceID, reg.ServerInfo.Version)

	info, err := client.QueryServerInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("plugins=%d uptime=%d", info.ConnectedPluginCount, info.UptimeSec)

	_ = client.Close(context.Background())
}
```

## Core options

- `WithPipePath(path)` for custom named pipe path
- `WithHeartbeatInterval(d)` to configure heartbeat interval
- `WithRequestTimeout(d)` to configure default request timeout
- `WithWaitBufferSize(n)` to tune `WaitMessage` buffer
- `WithMessageHandler(fn)` to receive async callbacks

## API coverage

### Query wrappers

- `QueryServerInfo`
- `QueryConnectedPlugins`
- `QueryRuntimeSnapshot`
- `QueryRigSnapshot`
- `QueryUDPSnapshot`
- `QueryQSOQueueSnapshot`
- `QuerySettingsSnapshot`
- `QueryPluginTelemetry`

### Command wrappers

- `SubscribeEvents`
- `ShowMainWindow`
- `HideMainWindow`
- `OpenWindow`
- `SendNotification`
- `ToggleUDPServer`
- `ToggleRigBackend`
- `SwitchRigBackend`
- `UploadExternalQSO`
- `StartRigBackend`
- `StopRigBackend`
- `RestartRigBackend`
- `TriggerQSOReupload`
- `UpdateSettings`

### Generic wrappers

- `RawQuery`
- `RawCommand`

## Command examples

```go
ctx := context.Background()

// 1) Subscribe to event topics
_, _ = client.SubscribeEvents(ctx, sdk.EventSubscription{
	Topics: []sdk.EnvelopeTopic{
		sdk.EnvelopeTopicEventServerStatus,
		sdk.EnvelopeTopicEventPluginLifecycle,
		sdk.EnvelopeTopicEventWsjtxDecodeBatch,
	},
})

// 2) Upload external ADIF QSO (maps to COMMAND_UPLOAD_EXTERNAL_QSO)
_, _ = client.UploadExternalQSO(ctx, "<CALL:6>BH1XYZ <MODE:3>FT8 <BAND:3>20M <EOR>")

// 3) Trigger reupload by qsoId (qsoId is required)
_, _ = client.TriggerQSOReupload(ctx, map[string]string{
	"qsoId": "your-qso-uuid",
})

// 4) Update settings patch
_, _ = client.UpdateSettings(ctx, sdk.SettingsPatch{
	Values: map[string]string{
		"udp.enable_udp_server": "true",
	},
})
```

## Message handling model

- `WithMessageHandler` receives async `sdk.Message` callback
- `WaitMessage(ctx)` allows pull-style consumption
- Inbound payloads are converted to typed models when possible
- Unknown payloads are mapped to `UnknownMessage` (type URL + raw bytes)

## Error model

- Transport/state errors: `ErrNotConnected`, `ErrClientClosed`, context timeout/cancel
- Remote response errors: `*RemoteError` (`Topic`, `Code`, `Message`, `CorrelationID`)

## Demo app

A full Fyne demo is included at:

- `examples/fyne-demo`
