# Fyne Demo for CLH Plugin Go SDK

This example is a full-featured manual test tool for:

- Connect / Disconnect
- Heartbeat option
- Callback message stream
- Blocking `WaitMessage` stream
- All SDK query APIs
- All SDK command APIs
- Raw query/command APIs

## Run

```bash
cd src/CloudlogHelper/clh-proto/pluginsdk/go/v20260309/examples/fyne-demo
go mod tidy
go run .
```

## Notes

- Start Cloudlog Helper first and make sure plugin pipe service is enabled.

