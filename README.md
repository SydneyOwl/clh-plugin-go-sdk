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

// 3) Trigger reupload by qsoIds (qsoIds is required, use ;;; as separator for multiple IDs)
_, _ = client.TriggerQSOReupload(ctx, map[string]string{
	"qsoIds": "your-qso-uuid",
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


## Cheatsheet
• Feature-to-API Cheat Sheet (Cloudlog Helper Plugin System)

### 1) Connect / lifecycle

| Feature | Go SDK | C# SDK | Notes |                                                                                                                                                                                             
  |---|---|---|---|                                                                                                                                                                                                                 
| Create client | NewClient(PluginManifest, ...options) | new ClhClient(PluginManifest, ClientOptions) | Uuid/Name/Version are required |
| Connect/register | Connect(ctx) | ConnectAsync() | Registers plugin with CLH |                                                                                                                                                  
| Heartbeat | Auto via WithHeartbeatInterval(...) | Auto via ClientOptions.HeartbeatInterval | Keepalive timeout is enforced by CLH |                                                                                             
| Graceful close | Close(ctx) | CloseAsync() | Sends deregister request |                                                                                                                                                         
| Inbound callback | WithMessageHandler(func(Message){...}) | ClientOptions.OnMessage / MessageReceived | Push mode |                                                                                                             
| Inbound pull | WaitMessage(ctx) | WaitMessageAsync(...) | Pull mode |                                                                                                                                                           

### 2) Query APIs (read state)

| Feature | Protocol topic | Go SDK | C# SDK |                                                                                                                                                                                    
  |---|---|---|---|                                                                                                                                                                                                                 
| Server info | QueryServerInfo | QueryServerInfo(ctx) | QueryServerInfoAsync() |                                                                                                                                                 
| Connected plugins | QueryConnectedPlugins | QueryConnectedPlugins(ctx) | QueryConnectedPluginsAsync() |                                                                                                                         
| Runtime snapshot (all-in-one) | QueryRuntimeSnapshot | QueryRuntimeSnapshot(ctx) | QueryRuntimeSnapshotAsync() |                                                                                                                
| Rig snapshot | QueryRigSnapshot | QueryRigSnapshot(ctx) | QueryRigSnapshotAsync() |                                                                                                                                             
| UDP snapshot | QueryUdpSnapshot | QueryUDPSnapshot(ctx) | QueryUdpSnapshotAsync() |                                                                                                                                             
| QSO queue snapshot | QueryQsoQueueSnapshot | QueryQSOQueueSnapshot(ctx) | QueryQsoQueueSnapshotAsync() |                                                                                                                        
| Settings snapshot | QuerySettingsSnapshot | QuerySettingsSnapshot(ctx) | QuerySettingsSnapshotAsync() |                                                                                                                         
| Plugin telemetry | QueryPluginTelemetry | QueryPluginTelemetry(ctx, pluginUUID) | QueryPluginTelemetryAsync(pluginUuid?) |                                                                                                      

### 3) Command APIs (control/actions)

| Feature | Protocol topic | Go SDK | C# SDK | Required input |                                                                                                                                                                   
  |---|---|---|---|---|                                                                                                                                                                                                             
| Subscribe event topics | CommandSubscribeEvents | SubscribeEvents(ctx, sub) | SubscribeEventsAsync(sub) | topics[] |                                                                                                            
| Show main window | CommandShowMainWindow | ShowMainWindow(ctx) | ShowMainWindowAsync() | - |                                                                                                                                    
| Hide main window | CommandHideMainWindow | HideMainWindow(ctx) | HideMainWindowAsync() | - |                                                                                                                                    
| Open specific window | CommandOpenWindow | OpenWindow(ctx, window, asDialog) | OpenWindowAsync(window, asDialog) | window, asDialog |                                                                                           
| Send in-app notification | CommandSendNotification | SendNotification(ctx, NotificationCommand) | SendNotificationAsync(command) | level, message |                                                                             
| Toggle UDP server | CommandToggleUdpServer | ToggleUDPServer(ctx, enabled*) | ToggleUdpServerAsync(enabled?) | optional enabled |                                                                                               
| Toggle rig backend polling | CommandToggleRigBackend | ToggleRigBackend(ctx, enabled*) | ToggleRigBackendAsync(enabled?) | optional enabled |                                                                                   
| Switch rig backend | CommandSwitchRigBackend | SwitchRigBackend(ctx, backend) | SwitchRigBackendAsync(backend) | Hamlib/FLRig/OmniRig |                                                                                         
| Upload external QSO(s) via ADIF | CommandUploadExternalQso | UploadExternalQSO(ctx, adifLogs) | UploadExternalQsoAsync(adifLogs) | attribute adifLogs |                                                                         
| Trigger QSO reupload | CommandTriggerQsoReupload | TriggerQSOReupload(ctx, attrs) | TriggerQsoReuploadAsync(...) | qsoIds (use ;;; separator) |
| Update settings | CommandUpdateSettings | UpdateSettings(ctx, patch) | UpdateSettingsAsync(patch) | SettingsPatch.Values |                                                                                                      
| Raw request escape hatch | any topic | RawQuery, RawCommand | RawQueryAsync, RawCommandAsync | advanced use |                                                                                                                   

### 4) Event topics you can subscribe to

| Topic | Payload type |                                                                                                                                                                                                          
  |---|---|                                                                                                                                                                                                                         
| EventServerStatus | ClhServerStatusChanged |                                                                                                                                                                                    
| EventPluginLifecycle | ClhPluginLifecycleChanged |                                                                                                                                                                              
| EventWsjtxMessage | WsjtxMessage |                                                                                                                                                                                              
| EventWsjtxDecodeRealtime | WsjtxMessage (decode payload) |                                                                                                                                                                      
| EventWsjtxDecodeBatch | PackedDecodeMessage |                                                                                                                                                                                   
| EventRigData | RigData |                                                                                                                                                                                                        
| EventQsoUploadStatus | ClhQSOUploadStatusChanged |                                                                                                                                                                              
| EventQsoQueueStatus | ClhQsoQueueStatusChanged |                                                                                                                                                                                
| EventSettingsChanged | ClhSettingsChanged |
| EventPluginTelemetry | ClhPluginTelemetryChanged |      


## License

This sdk is licensed under `The Unlicense`