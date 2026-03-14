package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	sdk "github.com/SydneyOwl/clh-plugin-go-sdk"
)

const (
	defaultPipePathUnix    = "/tmp/clh.plugin"
	defaultPipePathWindows = "\\\\.\\pipe\\clh.plugin"
)

type namedValue[T comparable] struct {
	Label string
	Value T
}

var eventTopicOptions = []namedValue[sdk.EnvelopeTopic]{
	{Label: "EVENT_SERVER_STATUS", Value: sdk.EnvelopeTopicEventServerStatus},
	{Label: "EVENT_PLUGIN_LIFECYCLE", Value: sdk.EnvelopeTopicEventPluginLifecycle},
	{Label: "EVENT_WSJTX_MESSAGE", Value: sdk.EnvelopeTopicEventWsjtxMessage},
	{Label: "EVENT_WSJTX_DECODE_REALTIME", Value: sdk.EnvelopeTopicEventWsjtxDecodeRealtime},
	{Label: "EVENT_WSJTX_DECODE_BATCH", Value: sdk.EnvelopeTopicEventWsjtxDecodeBatch},
	{Label: "EVENT_RIG_DATA", Value: sdk.EnvelopeTopicEventRigData},
	{Label: "EVENT_QSO_UPLOAD_STATUS", Value: sdk.EnvelopeTopicEventQsoUploadStatus},
	{Label: "EVENT_QSO_QUEUE_STATUS", Value: sdk.EnvelopeTopicEventQSOQueueStatus},
	{Label: "EVENT_SETTINGS_CHANGED", Value: sdk.EnvelopeTopicEventSettingsChanged},
	{Label: "EVENT_PLUGIN_TELEMETRY", Value: sdk.EnvelopeTopicEventPluginTelemetry},
}

var commandTopicOptions = []namedValue[sdk.EnvelopeTopic]{
	{Label: "COMMAND_SHOW_MAIN_WINDOW", Value: sdk.EnvelopeTopicCommandShowMainWindow},
	{Label: "COMMAND_HIDE_MAIN_WINDOW", Value: sdk.EnvelopeTopicCommandHideMainWindow},
	{Label: "COMMAND_OPEN_WINDOW", Value: sdk.EnvelopeTopicCommandOpenWindow},
	{Label: "COMMAND_SEND_NOTIFICATION", Value: sdk.EnvelopeTopicCommandSendNotification},
	{Label: "COMMAND_TOGGLE_UDP_SERVER", Value: sdk.EnvelopeTopicCommandToggleUDPServer},
	{Label: "COMMAND_TOGGLE_RIG_BACKEND", Value: sdk.EnvelopeTopicCommandToggleRigBackend},
	{Label: "COMMAND_SWITCH_RIG_BACKEND", Value: sdk.EnvelopeTopicCommandSwitchRigBackend},
	{Label: "COMMAND_UPLOAD_EXTERNAL_QSO", Value: sdk.EnvelopeTopicCommandUploadExternalQSO},
	{Label: "COMMAND_TRIGGER_QSO_REUPLOAD", Value: sdk.EnvelopeTopicCommandTriggerQSOReupload},
	{Label: "COMMAND_UPDATE_SETTINGS", Value: sdk.EnvelopeTopicCommandUpdateSettings},
	{Label: "COMMAND_SUBSCRIBE_EVENTS", Value: sdk.EnvelopeTopicCommandSubscribeEvents},
}

var queryTopicOptions = []namedValue[sdk.EnvelopeTopic]{
	{Label: "QUERY_SERVER_INFO", Value: sdk.EnvelopeTopicQueryServerInfo},
	{Label: "QUERY_CONNECTED_PLUGINS", Value: sdk.EnvelopeTopicQueryConnectedPlugins},
	{Label: "QUERY_RUNTIME_SNAPSHOT", Value: sdk.EnvelopeTopicQueryRuntimeSnapshot},
	{Label: "QUERY_RIG_SNAPSHOT", Value: sdk.EnvelopeTopicQueryRigSnapshot},
	{Label: "QUERY_UDP_SNAPSHOT", Value: sdk.EnvelopeTopicQueryUDPSnapshot},
	{Label: "QUERY_QSO_QUEUE_SNAPSHOT", Value: sdk.EnvelopeTopicQueryQSOQueueSnapshot},
	{Label: "QUERY_SETTINGS_SNAPSHOT", Value: sdk.EnvelopeTopicQuerySettingsSnapshot},
	{Label: "QUERY_PLUGIN_TELEMETRY", Value: sdk.EnvelopeTopicQueryPluginTelemetry},
}

var rawTopicOptions = append(append([]namedValue[sdk.EnvelopeTopic]{}, queryTopicOptions...), commandTopicOptions...)

var openWindowOptions = []namedValue[sdk.ControllableWindow]{
	{Label: "SettingsWindow", Value: sdk.WindowSettings},
	{Label: "AboutWindow", Value: sdk.WindowAbout},
	{Label: "QSOAssistantWindow", Value: sdk.WindowQSOAssistant},
	{Label: "StationStatisticWindow", Value: sdk.WindowStationStats},
	{Label: "PolarChartWindow", Value: sdk.WindowPolarChart},
}

var notificationLevels = []namedValue[sdk.NotificationLevel]{
	{Label: "INFO", Value: sdk.NotificationLevelInfo},
	{Label: "SUCCESS", Value: sdk.NotificationLevelSuccess},
	{Label: "WARNING", Value: sdk.NotificationLevelWarning},
	{Label: "ERROR", Value: sdk.NotificationLevelError},
}

var rigBackendOptions = []namedValue[sdk.RigBackend]{
	{Label: "Hamlib", Value: sdk.RigBackendHamlib},
	{Label: "FLRig", Value: sdk.RigBackendFLRig},
	{Label: "OmniRig", Value: sdk.RigBackendOmniRig},
}

type demoUI struct {
	app fyne.App
	win fyne.Window

	statusData binding.String
	resultData binding.String
	logData    binding.String

	logMu sync.Mutex

	clientMu sync.RWMutex
	client   *sdk.Client

	waitMu     sync.Mutex
	waitCancel context.CancelFunc

	pipeEntry           *widget.Entry
	uuidEntry           *widget.Entry
	nameEntry           *widget.Entry
	versionEntry        *widget.Entry
	descriptionEntry    *widget.Entry
	heartbeatSecEntry   *widget.Entry
	timeoutSecEntry     *widget.Entry
	waitBufferSizeEntry *widget.Entry

	eventTopicChecks   *widget.CheckGroup
	telemetryUUIDEntry *widget.Entry

	openWindowSelect   *widget.Select
	openWindowAsDialog *widget.Check
	notificationLevel  *widget.Select
	notificationTitle  *widget.Entry
	notificationBody   *widget.Entry
	rigToggleSelect    *widget.Select
	rigBackendSelect   *widget.Select
	udpToggleSelect    *widget.Select
	qsoAttrsEntry      *widget.Entry
	externalAdifEntry  *widget.Entry
	settingsPatchEntry *widget.Entry

	rawKindSelect  *widget.Select
	rawTopicSelect *widget.Select
	rawAttrsEntry  *widget.Entry
}

func main() {
	a := app.NewWithID("clh.plugin.sdk.fyne.demo")
	w := a.NewWindow("CLH Plugin SDK Demo (Fyne)")
	w.Resize(fyne.NewSize(1600, 980))

	ui := newDemoUI(a, w)
	w.SetContent(ui.build())
	w.SetCloseIntercept(func() {
		ui.shutdown()
		w.Close()
	})

	w.ShowAndRun()
}

func newDemoUI(a fyne.App, w fyne.Window) *demoUI {
	ui := &demoUI{
		app:        a,
		win:        w,
		statusData: binding.NewString(),
		resultData: binding.NewString(),
		logData:    binding.NewString(),
	}

	ui.initWidgets()
	_ = ui.statusData.Set("Disconnected")
	_ = ui.resultData.Set("Ready")
	_ = ui.logData.Set("")
	return ui
}

func (d *demoUI) initWidgets() {
	d.pipeEntry = widget.NewEntry()
	d.pipeEntry.SetText(defaultPipePathUnix)
	if runtime.GOOS == "windows" {
		d.pipeEntry.SetText(defaultPipePathWindows)
	}

	d.uuidEntry = widget.NewEntry()
	d.uuidEntry.SetText(fmt.Sprintf("fyne-demo-%d", time.Now().Unix()))

	d.nameEntry = widget.NewEntry()
	d.nameEntry.SetText("fyne-sdk-demo")

	d.versionEntry = widget.NewEntry()
	d.versionEntry.SetText("0.1.0")

	d.descriptionEntry = widget.NewEntry()
	d.descriptionEntry.SetText("Fyne demo for CLH plugin Go SDK")

	d.heartbeatSecEntry = widget.NewEntry()
	d.heartbeatSecEntry.SetText("5")

	d.timeoutSecEntry = widget.NewEntry()
	d.timeoutSecEntry.SetText("8")

	d.waitBufferSizeEntry = widget.NewEntry()
	d.waitBufferSizeEntry.SetText("256")

	d.telemetryUUIDEntry = widget.NewEntry()
	d.telemetryUUIDEntry.SetPlaceHolder("empty = self")

	d.eventTopicChecks = widget.NewCheckGroup(optionLabels(eventTopicOptions), nil)
	d.eventTopicChecks.SetSelected(optionLabels(eventTopicOptions))

	d.openWindowSelect = widget.NewSelect(optionLabels(openWindowOptions), nil)
	d.openWindowSelect.SetSelected(openWindowOptions[0].Label)

	d.openWindowAsDialog = widget.NewCheck("Open as dialog", nil)
	d.openWindowAsDialog.SetChecked(false)

	d.notificationLevel = widget.NewSelect(optionLabels(notificationLevels), nil)
	d.notificationLevel.SetSelected("INFO")

	d.notificationTitle = widget.NewEntry()
	d.notificationTitle.SetText("SDK Demo")
	d.notificationBody = widget.NewMultiLineEntry()
	d.notificationBody.SetMinRowsVisible(2)
	d.notificationBody.SetText("Hello from Fyne demo")

	d.rigToggleSelect = widget.NewSelect([]string{"Auto", "Enable", "Disable"}, nil)
	d.rigToggleSelect.SetSelected("Auto")

	d.rigBackendSelect = widget.NewSelect(optionLabels(rigBackendOptions), nil)
	d.rigBackendSelect.SetSelected(rigBackendOptions[0].Label)

	d.udpToggleSelect = widget.NewSelect([]string{"Auto", "Enable", "Disable"}, nil)
	d.udpToggleSelect.SetSelected("Auto")

	d.qsoAttrsEntry = widget.NewMultiLineEntry()
	d.qsoAttrsEntry.SetMinRowsVisible(4)
	d.qsoAttrsEntry.SetPlaceHolder("key=value per line\nExample:\nqsoIds=1")

	d.externalAdifEntry = widget.NewMultiLineEntry()
	d.externalAdifEntry.SetMinRowsVisible(6)
	d.externalAdifEntry.SetPlaceHolder("<call:5>BA1ABC <gridsquare:4>OL12 <mode:3>FT8 <rst_sent:3>-17 <rst_rcvd:3>-13 <qso_date:8>20240930 <time_on:6>024231 <qso_date_off:8>20240930 <time_off:6>024314 <band:2>6m <freq:9>50.314044 <station_callsign:6>BA2ABC <my_gridsquare:4>OL34 <eor>")

	d.settingsPatchEntry = widget.NewMultiLineEntry()
	d.settingsPatchEntry.SetMinRowsVisible(6)
	d.settingsPatchEntry.SetPlaceHolder("key=value per line\nExample:\nudp.enable_udp_server=true")

	d.rawKindSelect = widget.NewSelect([]string{"query", "command"}, nil)
	d.rawKindSelect.SetSelected("query")

	d.rawTopicSelect = widget.NewSelect(optionLabels(rawTopicOptions), nil)
	d.rawTopicSelect.SetSelected(rawTopicOptions[0].Label)

	d.rawAttrsEntry = widget.NewMultiLineEntry()
	d.rawAttrsEntry.SetMinRowsVisible(4)
	d.rawAttrsEntry.SetPlaceHolder("key=value per line")
}

func (d *demoUI) build() fyne.CanvasObject {
	top := d.buildTopBar()

	leftTabs := container.NewAppTabs(
		container.NewTabItem("Connection", d.buildConnectionTab()),
		container.NewTabItem("Query", d.buildQueryTab()),
		container.NewTabItem("Command", d.buildCommandTab()),
		container.NewTabItem("Raw", d.buildRawTab()),
	)

	resultEntry := widget.NewMultiLineEntry()
	resultEntry.Bind(d.resultData)
	resultEntry.SetMinRowsVisible(22)
	resultEntry.Wrapping = fyne.TextWrapWord

	logEntry := widget.NewMultiLineEntry()
	logEntry.Bind(d.logData)
	logEntry.SetMinRowsVisible(22)
	logEntry.Wrapping = fyne.TextWrapWord

	rightTabs := container.NewAppTabs(
		container.NewTabItem("Result", container.NewVScroll(resultEntry)),
		container.NewTabItem("Logs", container.NewVScroll(logEntry)),
	)

	split := container.NewHSplit(leftTabs, rightTabs)
	split.Offset = 0.57

	return container.NewBorder(top, nil, nil, nil, split)
}

func (d *demoUI) buildTopBar() fyne.CanvasObject {
	connectBtn := widget.NewButton("Connect", func() { d.connect() })
	disconnectBtn := widget.NewButton("Disconnect", func() { d.disconnect() })
	startWaitBtn := widget.NewButton("Start WaitMessage", func() { d.startWaitLoop() })
	stopWaitBtn := widget.NewButton("Stop WaitMessage", func() { d.stopWaitLoop() })
	clearLogBtn := widget.NewButton("Clear Logs", func() { _ = d.logData.Set("") })
	statusLabel := widget.NewLabelWithData(d.statusData)

	return container.NewVBox(
		container.NewHBox(
			connectBtn,
			disconnectBtn,
			startWaitBtn,
			stopWaitBtn,
			clearLogBtn,
			widget.NewSeparator(),
			widget.NewLabel("Status:"),
			statusLabel,
		),
		widget.NewSeparator(),
	)
}

func (d *demoUI) buildConnectionTab() fyne.CanvasObject {
	form := widget.NewForm(
		widget.NewFormItem("Pipe Path", d.pipeEntry),
		widget.NewFormItem("Plugin UUID", d.uuidEntry),
		widget.NewFormItem("Plugin Name", d.nameEntry),
		widget.NewFormItem("Plugin Version", d.versionEntry),
		widget.NewFormItem("Description", d.descriptionEntry),
		widget.NewFormItem("Heartbeat Seconds", d.heartbeatSecEntry),
		widget.NewFormItem("Request Timeout Seconds", d.timeoutSecEntry),
		widget.NewFormItem("Wait Buffer Size", d.waitBufferSizeEntry),
		widget.NewFormItem("Telemetry Query UUID", d.telemetryUUIDEntry),
	)

	eventCard := widget.NewCard(
		"Event Subscription",
		"Select event topics",
		container.NewVBox(d.eventTopicChecks),
	)

	content := container.NewVBox(form, widget.NewSeparator(), eventCard)
	return container.NewVScroll(content)
}

func (d *demoUI) buildQueryTab() fyne.CanvasObject {
	buttons := []fyne.CanvasObject{
		widget.NewButton("Query ServerInfo", func() {
			d.runClientCall("QueryServerInfo", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryServerInfo(ctx)
			})
		}),
		widget.NewButton("Query ConnectedPlugins", func() {
			d.runClientCall("QueryConnectedPlugins", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryConnectedPlugins(ctx)
			})
		}),
		widget.NewButton("Query RuntimeSnapshot", func() {
			d.runClientCall("QueryRuntimeSnapshot", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryRuntimeSnapshot(ctx)
			})
		}),
		widget.NewButton("Query RigSnapshot", func() {
			d.runClientCall("QueryRigSnapshot", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryRigSnapshot(ctx)
			})
		}),
		widget.NewButton("Query UDPSnapshot", func() {
			d.runClientCall("QueryUDPSnapshot", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryUDPSnapshot(ctx)
			})
		}),
		widget.NewButton("Query QSOQueueSnapshot", func() {
			d.runClientCall("QueryQSOQueueSnapshot", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryQSOQueueSnapshot(ctx)
			})
		}),
		widget.NewButton("Query SettingsSnapshot", func() {
			d.runClientCall("QuerySettingsSnapshot", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QuerySettingsSnapshot(ctx)
			})
		}),
		widget.NewButton("Query PluginTelemetry", func() {
			d.runClientCall("QueryPluginTelemetry", func(ctx context.Context, c *sdk.Client) (any, error) {
				return c.QueryPluginTelemetry(ctx, strings.TrimSpace(d.telemetryUUIDEntry.Text))
			})
		}),
	}

	grid := container.NewGridWithColumns(3, buttons...)
	return container.NewVScroll(container.NewVBox(grid))
}

func (d *demoUI) buildCommandTab() fyne.CanvasObject {
	subscriptionCard := widget.NewCard(
		"Subscription Commands",
		"Update event subscription",
		container.NewVBox(
			widget.NewButton("Subscribe Events", func() { d.commandSubscribeEvents() }),
		),
	)

	windowCard := widget.NewCard(
		"UI Commands",
		"Main window and open specific window",
		container.NewVBox(
			container.NewGridWithColumns(2,
				widget.NewButton("Show Main Window", func() {
					d.runClientCall("ShowMainWindow", func(ctx context.Context, c *sdk.Client) (any, error) {
						return nil, c.ShowMainWindow(ctx)
					})
				}),
				widget.NewButton("Hide Main Window", func() {
					d.runClientCall("HideMainWindow", func(ctx context.Context, c *sdk.Client) (any, error) {
						return nil, c.HideMainWindow(ctx)
					})
				}),
			),
			container.NewHBox(widget.NewLabel("Window"), d.openWindowSelect, d.openWindowAsDialog),
			widget.NewButton("Open Selected Window", func() { d.commandOpenWindow() }),
		),
	)

	notificationCard := widget.NewCard(
		"Notification Command",
		"Send in-app notification",
		container.NewVBox(
			container.NewHBox(widget.NewLabel("Level"), d.notificationLevel),
			container.NewHBox(widget.NewLabel("Title"), d.notificationTitle),
			d.notificationBody,
			widget.NewButton("Send Notification", func() { d.commandSendNotification() }),
		),
	)

	rigUdpCard := widget.NewCard(
		"Rig + UDP Commands",
		"Rig backend controls (full SDK wrappers) and UDP toggle",
		container.NewVBox(
			container.NewGridWithColumns(3,
				widget.NewButton("Toggle Rig Backend", func() {
					d.commandToggleRigBackend()
				}),
				widget.NewButton("Switch Rig Backend", func() {
					d.commandSwitchRigBackend()
				}),
			),
			container.NewHBox(widget.NewLabel("Rig Target"), d.rigToggleSelect),
			container.NewHBox(widget.NewLabel("Rig Backend"), d.rigBackendSelect),
			container.NewHBox(widget.NewLabel("UDP Target"), d.udpToggleSelect),
			widget.NewButton("Toggle UDP Server", func() { d.commandToggleUDP() }),
		),
	)

	qsoSettingsCard := widget.NewCard(
		"QSO + Settings Commands",
		"Trigger QSO reupload, upload external ADIF, and update settings patch",
		container.NewVBox(
			widget.NewLabel("QSO Reupload Attributes (key=value per line)"),
			d.qsoAttrsEntry,
			widget.NewButton("Trigger QSO Reupload", func() {
				d.commandTriggerQSOReupload()
			}),
			widget.NewSeparator(),
			widget.NewLabel("External QSO ADIF (COMMAND_UPLOAD_EXTERNAL_QSO)"),
			d.externalAdifEntry,
			widget.NewButton("Upload External QSO", func() {
				d.commandUploadExternalQSO()
			}),
			widget.NewSeparator(),
			widget.NewLabel("Settings Patch (key=value per line)"),
			d.settingsPatchEntry,
			widget.NewButton("Update Settings", func() {
				d.commandUpdateSettings()
			}),
		),
	)

	return container.NewVScroll(container.NewVBox(
		subscriptionCard,
		windowCard,
		notificationCard,
		rigUdpCard,
		qsoSettingsCard,
	))
}

func (d *demoUI) buildRawTab() fyne.CanvasObject {
	sendBtn := widget.NewButton("Send Raw Request", func() { d.sendRawRequest() })
	return container.NewVScroll(container.NewVBox(
		widget.NewCard(
			"Raw Query/Command",
			"Use RawQuery / RawCommand (payload omitted in this demo, attributes only)",
			container.NewVBox(
				container.NewGridWithColumns(2,
					container.NewHBox(widget.NewLabel("Kind"), d.rawKindSelect),
					container.NewHBox(widget.NewLabel("Topic"), d.rawTopicSelect),
				),
				widget.NewLabel("Attributes (key=value per line)"),
				d.rawAttrsEntry,
				sendBtn,
			),
		),
	))
}

func (d *demoUI) connect() {
	d.runTask("Connect", func() (any, error) {
		heartbeatSeconds, err := parsePositiveInt(d.heartbeatSecEntry.Text, "heartbeat seconds")
		if err != nil {
			return nil, err
		}
		timeoutSeconds, err := parsePositiveInt(d.timeoutSecEntry.Text, "request timeout seconds")
		if err != nil {
			return nil, err
		}
		waitBuffer, err := parsePositiveInt(d.waitBufferSizeEntry.Text, "wait buffer size")
		if err != nil {
			return nil, err
		}

		eventSub, err := d.currentEventSubscription()
		if err != nil {
			return nil, err
		}

		manifest := sdk.PluginManifest{
			UUID:              strings.TrimSpace(d.uuidEntry.Text),
			Name:              strings.TrimSpace(d.nameEntry.Text),
			Version:           strings.TrimSpace(d.versionEntry.Text),
			Description:       strings.TrimSpace(d.descriptionEntry.Text),
			EventSubscription: &eventSub,
			Metadata: map[string]string{
				"demo":   "fyne",
				"source": "examples/fyne-demo",
			},
		}

		client, err := sdk.NewClient(
			manifest,
			sdk.WithPipePath(strings.TrimSpace(d.pipeEntry.Text)),
			sdk.WithHeartbeatInterval(time.Duration(heartbeatSeconds)*time.Second),
			sdk.WithRequestTimeout(time.Duration(timeoutSeconds)*time.Second),
			sdk.WithWaitBufferSize(waitBuffer),
			sdk.WithMessageHandler(d.onCallbackMessage),
		)
		if err != nil {
			return nil, err
		}

		if old := d.getClient(); old != nil {
			_ = old.Close(context.Background())
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
		defer cancel()

		resp, err := client.Connect(ctx)
		if err != nil {
			return nil, err
		}

		d.setClient(client)
		_ = d.statusData.Set(fmt.Sprintf("Connected (%s, %s)", resp.InstanceID, resp.ServerInfo.Version))
		return resp, nil
	})
}

func (d *demoUI) disconnect() {
	d.runTask("Disconnect", func() (any, error) {
		d.stopWaitLoop()
		client := d.getClient()
		if client == nil {
			_ = d.statusData.Set("Disconnected")
			return "no active client", nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		err := client.Close(ctx)
		d.setClient(nil)
		_ = d.statusData.Set("Disconnected")
		return "closed", err
	})
}

func (d *demoUI) shutdown() {
	d.stopWaitLoop()
	client := d.getClient()
	if client == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = client.Close(ctx)
	d.setClient(nil)
}

func (d *demoUI) startWaitLoop() {
	client := d.getClient()
	if client == nil {
		d.appendLog("wait loop start failed: client not connected")
		return
	}

	d.waitMu.Lock()
	if d.waitCancel != nil {
		d.waitMu.Unlock()
		d.appendLog("wait loop already running")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	d.waitCancel = cancel
	d.waitMu.Unlock()

	d.appendLog("wait loop started")

	go func(c *sdk.Client) {
		defer func() {
			d.waitMu.Lock()
			d.waitCancel = nil
			d.waitMu.Unlock()
			d.appendLog("wait loop stopped")
		}()
		for {
			msg, err := c.WaitMessage(ctx)
			if err != nil {
				return
			}
			d.appendLog("wait <- %s", summarizeMessage(msg))
			d.showResult("WaitMessage", msg)
		}
	}(client)
}

func (d *demoUI) stopWaitLoop() {
	d.waitMu.Lock()
	cancel := d.waitCancel
	d.waitCancel = nil
	d.waitMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (d *demoUI) onCallbackMessage(msg sdk.Message) {
	d.appendLog("callback <- %s", summarizeMessage(msg))
	d.showResult("Callback Message", msg)
}

func (d *demoUI) runClientCall(name string, fn func(ctx context.Context, c *sdk.Client) (any, error)) {
	client := d.getClient()
	if client == nil {
		d.appendLog("%s failed: client not connected", name)
		return
	}

	d.runTask(name, func() (any, error) {
		ctx, cancel := context.WithTimeout(context.Background(), d.requestTimeout())
		defer cancel()
		return fn(ctx, client)
	})
}

func (d *demoUI) commandSubscribeEvents() {
	sub, err := d.currentEventSubscription()
	if err != nil {
		d.appendLog("SubscribeEvents input error: %v", err)
		return
	}
	d.runClientCall("SubscribeEvents", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.SubscribeEvents(ctx, sub)
	})
}

func (d *demoUI) commandOpenWindow() {
	window, ok := valueByLabel(d.openWindowSelect.Selected, openWindowOptions)
	if !ok {
		d.appendLog("OpenWindow input error: invalid window")
		return
	}
	asDialog := d.openWindowAsDialog.Checked
	d.runClientCall("OpenWindow", func(ctx context.Context, c *sdk.Client) (any, error) {
		return nil, c.OpenWindow(ctx, window, asDialog)
	})
}

func (d *demoUI) commandSendNotification() {
	level, ok := valueByLabel(d.notificationLevel.Selected, notificationLevels)
	if !ok {
		d.appendLog("SendNotification input error: invalid level")
		return
	}
	cmd := sdk.NotificationCommand{
		Level:   level,
		Title:   strings.TrimSpace(d.notificationTitle.Text),
		Message: strings.TrimSpace(d.notificationBody.Text),
	}
	d.runClientCall("SendNotification", func(ctx context.Context, c *sdk.Client) (any, error) {
		return nil, c.SendNotification(ctx, cmd)
	})
}

func (d *demoUI) commandToggleUDP() {
	var enabled *bool
	switch d.udpToggleSelect.Selected {
	case "Enable":
		v := true
		enabled = &v
	case "Disable":
		v := false
		enabled = &v
	}
	d.runClientCall("ToggleUDPServer", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.ToggleUDPServer(ctx, enabled)
	})
}

func (d *demoUI) commandToggleRigBackend() {
	var enabled *bool
	switch d.rigToggleSelect.Selected {
	case "Enable":
		v := true
		enabled = &v
	case "Disable":
		v := false
		enabled = &v
	}
	d.runClientCall("ToggleRigBackend", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.ToggleRigBackend(ctx, enabled)
	})
}

func (d *demoUI) commandSwitchRigBackend() {
	backend, ok := valueByLabel(d.rigBackendSelect.Selected, rigBackendOptions)
	if !ok {
		d.appendLog("SwitchRigBackend input error: invalid backend")
		return
	}
	d.runClientCall("SwitchRigBackend", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.SwitchRigBackend(ctx, backend)
	})
}

func (d *demoUI) commandTriggerQSOReupload() {
	attrs, err := parseKeyValueLines(d.qsoAttrsEntry.Text)
	if err != nil {
		d.appendLog("TriggerQSOReupload input error: %v", err)
		return
	}
	d.runClientCall("TriggerQSOReupload", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.TriggerQSOReupload(ctx, attrs)
	})
}

func (d *demoUI) commandUploadExternalQSO() {
	adif := strings.TrimSpace(d.externalAdifEntry.Text)
	if adif == "" {
		d.appendLog("UploadExternalQSO input error: adif text is empty")
		return
	}
	d.runClientCall("UploadExternalQSO", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.UploadExternalQSO(ctx, adif)
	})
}

func (d *demoUI) commandUpdateSettings() {
	values, err := parseKeyValueLines(d.settingsPatchEntry.Text)
	if err != nil {
		d.appendLog("UpdateSettings input error: %v", err)
		return
	}
	d.runClientCall("UpdateSettings", func(ctx context.Context, c *sdk.Client) (any, error) {
		return c.UpdateSettings(ctx, sdk.SettingsPatch{Values: values})
	})
}

func (d *demoUI) sendRawRequest() {
	topic, ok := valueByLabel(d.rawTopicSelect.Selected, rawTopicOptions)
	if !ok {
		d.appendLog("Raw request input error: invalid topic")
		return
	}

	attrs, err := parseKeyValueLines(d.rawAttrsEntry.Text)
	if err != nil {
		d.appendLog("Raw request input error: %v", err)
		return
	}

	kind := strings.ToLower(strings.TrimSpace(d.rawKindSelect.Selected))
	switch kind {
	case "query":
		d.runClientCall("RawQuery", func(ctx context.Context, c *sdk.Client) (any, error) {
			return c.RawQuery(ctx, topic, attrs, nil)
		})
	case "command":
		d.runClientCall("RawCommand", func(ctx context.Context, c *sdk.Client) (any, error) {
			return c.RawCommand(ctx, topic, attrs, nil, nil)
		})
	default:
		d.appendLog("Raw request input error: unknown kind %q", kind)
	}
}

func (d *demoUI) currentEventSubscription() (sdk.EventSubscription, error) {
	sub := sdk.EventSubscription{
		Topics: selectedValues(d.eventTopicChecks.Selected, eventTopicOptions),
	}
	return sub, nil
}

func (d *demoUI) runTask(name string, fn func() (any, error)) {
	go func() {
		d.appendLog("%s: start", name)
		result, err := fn()
		if err != nil {
			d.appendLog("%s: error: %v", name, err)
			d.showResult(name+" Error", map[string]any{
				"error": err.Error(),
			})
			return
		}
		d.appendLog("%s: success", name)
		if result != nil {
			d.showResult(name+" Result", result)
		}
	}()
}

func (d *demoUI) showResult(title string, value any) {
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		body = []byte(fmt.Sprintf("%+v", value))
	}
	_ = d.resultData.Set(fmt.Sprintf("%s\n\n%s", title, string(body)))
}

func (d *demoUI) appendLog(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	logLine := fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), line)

	d.logMu.Lock()
	defer d.logMu.Unlock()

	existing, _ := d.logData.Get()
	updated := existing + logLine
	const maxSize = 128 * 1024
	if len(updated) > maxSize {
		updated = updated[len(updated)-maxSize:]
		if idx := strings.Index(updated, "\n"); idx > 0 {
			updated = updated[idx+1:]
		}
	}
	_ = d.logData.Set(updated)
}

func (d *demoUI) requestTimeout() time.Duration {
	value, err := parsePositiveInt(d.timeoutSecEntry.Text, "request timeout seconds")
	if err != nil {
		return 8 * time.Second
	}
	return time.Duration(value) * time.Second
}

func (d *demoUI) setClient(client *sdk.Client) {
	d.clientMu.Lock()
	d.client = client
	d.clientMu.Unlock()
}

func (d *demoUI) getClient() *sdk.Client {
	d.clientMu.RLock()
	defer d.clientMu.RUnlock()
	return d.client
}

func summarizeMessage(msg sdk.Message) string {
	switch msg.Kind {
	case sdk.InboundKindRigData:
		if msg.RigData == nil {
			return "rig_data(nil)"
		}
		return fmt.Sprintf("rig_data provider=%s freq=%d", msg.RigData.Provider, msg.RigData.Frequency)
	case sdk.InboundKindCLHInternal:
		return "clh_internal"
	case sdk.InboundKindEnvelope:
		if msg.Envelope == nil {
			return "envelope(nil)"
		}
		return fmt.Sprintf(
			"envelope kind=%d topic=%d success=%t error_code=%s",
			msg.Envelope.Kind,
			msg.Envelope.Topic,
			msg.Envelope.Success,
			msg.Envelope.ErrorCode,
		)
	case sdk.InboundKindConnectionClosed:
		return "connection_closed"
	case sdk.InboundKindUnknown:
		if msg.Unknown == nil {
			return "unknown"
		}
		return fmt.Sprintf("unknown type_url=%s", msg.Unknown.TypeURL)
	default:
		return string(msg.Kind)
	}
}

func parseKeyValueLines(text string) (map[string]string, error) {
	result := map[string]string{}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("line %d must be key=value", i+1)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("line %d key is empty", i+1)
		}
		result[key] = value
	}
	return result, nil
}

func parsePositiveInt(text string, fieldName string) (int, error) {
	value, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil {
		return 0, fmt.Errorf("%s parse failed: %w", fieldName, err)
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be > 0", fieldName)
	}
	return value, nil
}

func optionLabels[T comparable](values []namedValue[T]) []string {
	out := make([]string, 0, len(values))
	for _, item := range values {
		out = append(out, item.Label)
	}
	return out
}

func selectedValues[T comparable](selected []string, options []namedValue[T]) []T {
	index := map[string]T{}
	for _, item := range options {
		index[item.Label] = item.Value
	}

	out := make([]T, 0, len(selected))
	for _, label := range selected {
		value, ok := index[label]
		if ok {
			out = append(out, value)
		}
	}
	return out
}

func valueByLabel[T comparable](label string, options []namedValue[T]) (T, bool) {
	var zero T
	for _, item := range options {
		if item.Label == label {
			return item.Value, true
		}
	}
	return zero, false
}
