package clhplugin

import "time"

type EnvelopeKind int32

const (
	EnvelopeKindUnspecified EnvelopeKind = 0
	EnvelopeKindEvent       EnvelopeKind = 1
	EnvelopeKindQuery       EnvelopeKind = 2
	EnvelopeKindCommand     EnvelopeKind = 3
	EnvelopeKindResponse    EnvelopeKind = 4
)

type EnvelopeTopic int32

const (
	EnvelopeTopicUnspecified               EnvelopeTopic = 0
	EnvelopeTopicEventServerStatus         EnvelopeTopic = 1
	EnvelopeTopicEventPluginLifecycle      EnvelopeTopic = 2
	EnvelopeTopicEventWsjtxMessage         EnvelopeTopic = 3
	EnvelopeTopicEventWsjtxDecodeRealtime  EnvelopeTopic = 4
	EnvelopeTopicEventWsjtxDecodeBatch     EnvelopeTopic = 5
	EnvelopeTopicEventRigData              EnvelopeTopic = 6
	EnvelopeTopicEventQsoUploadStatus      EnvelopeTopic = 7
	EnvelopeTopicEventQSOQueueStatus       EnvelopeTopic = 10
	EnvelopeTopicEventSettingsChanged      EnvelopeTopic = 11
	EnvelopeTopicEventPluginTelemetry      EnvelopeTopic = 12
	EnvelopeTopicQueryServerInfo           EnvelopeTopic = 100
	EnvelopeTopicQueryConnectedPlugins     EnvelopeTopic = 101
	EnvelopeTopicQueryRuntimeSnapshot      EnvelopeTopic = 102
	EnvelopeTopicQueryRigSnapshot          EnvelopeTopic = 103
	EnvelopeTopicQueryUDPSnapshot          EnvelopeTopic = 104
	EnvelopeTopicQueryQSOQueueSnapshot     EnvelopeTopic = 105
	EnvelopeTopicQuerySettingsSnapshot     EnvelopeTopic = 106
	EnvelopeTopicQueryPluginTelemetry      EnvelopeTopic = 107
	EnvelopeTopicCommandShowMainWindow     EnvelopeTopic = 200
	EnvelopeTopicCommandHideMainWindow     EnvelopeTopic = 201
	EnvelopeTopicCommandOpenWindow         EnvelopeTopic = 202
	EnvelopeTopicCommandSendNotification   EnvelopeTopic = 203
	EnvelopeTopicCommandToggleUDPServer    EnvelopeTopic = 204
	EnvelopeTopicCommandToggleRigBackend   EnvelopeTopic = 205
	EnvelopeTopicCommandSwitchRigBackend   EnvelopeTopic = 206
	EnvelopeTopicCommandUploadExternalQSO  EnvelopeTopic = 207
	EnvelopeTopicCommandTriggerQSOReupload EnvelopeTopic = 208
	EnvelopeTopicCommandUpdateSettings     EnvelopeTopic = 209
	EnvelopeTopicCommandSubscribeEvents    EnvelopeTopic = 210
)

type NotificationLevel int32

const (
	NotificationLevelUnspecified NotificationLevel = 0
	NotificationLevelInfo        NotificationLevel = 1
	NotificationLevelSuccess     NotificationLevel = 2
	NotificationLevelWarning     NotificationLevel = 3
	NotificationLevelError       NotificationLevel = 4
)

type WsjtxMessageType int32

const (
	WsjtxMessageTypeHeartbeat           WsjtxMessageType = 0
	WsjtxMessageTypeStatus              WsjtxMessageType = 1
	WsjtxMessageTypeDecode              WsjtxMessageType = 2
	WsjtxMessageTypeClear               WsjtxMessageType = 3
	WsjtxMessageTypeReply               WsjtxMessageType = 4
	WsjtxMessageTypeQSOLogged           WsjtxMessageType = 5
	WsjtxMessageTypeClose               WsjtxMessageType = 6
	WsjtxMessageTypeReplay              WsjtxMessageType = 7
	WsjtxMessageTypeHaltTx              WsjtxMessageType = 8
	WsjtxMessageTypeFreeText            WsjtxMessageType = 9
	WsjtxMessageTypeWSPRDecode          WsjtxMessageType = 10
	WsjtxMessageTypeLocation            WsjtxMessageType = 11
	WsjtxMessageTypeLoggedADIF          WsjtxMessageType = 12
	WsjtxMessageTypeHighlightCallsign   WsjtxMessageType = 13
	WsjtxMessageTypeSwitchConfiguration WsjtxMessageType = 14
	WsjtxMessageTypeConfigure           WsjtxMessageType = 15
)

type SpecialOperationMode int32

const (
	SpecialOperationModeNone     SpecialOperationMode = 0
	SpecialOperationModeNAVHF    SpecialOperationMode = 1
	SpecialOperationModeEUVHF    SpecialOperationMode = 2
	SpecialOperationModeFieldDay SpecialOperationMode = 3
	SpecialOperationModeRTTYRU   SpecialOperationMode = 4
	SpecialOperationModeWWDIGI   SpecialOperationMode = 5
	SpecialOperationModeFox      SpecialOperationMode = 6
	SpecialOperationModeHound    SpecialOperationMode = 7
)

type ClearWindow int32

const (
	ClearWindowBandActivity ClearWindow = 0
	ClearWindowRxFrequency  ClearWindow = 1
	ClearWindowBoth         ClearWindow = 2
)

type UploadStatus int32

const (
	UploadStatusUnspecified UploadStatus = 0
	UploadStatusPending     UploadStatus = 1
	UploadStatusUploading   UploadStatus = 2
	UploadStatusSuccess     UploadStatus = 3
	UploadStatusFail        UploadStatus = 4
	UploadStatusIgnored     UploadStatus = 5
)

type PluginLifecycleEventType int32

const (
	PluginLifecycleEventUnspecified  PluginLifecycleEventType = 0
	PluginLifecycleEventConnected    PluginLifecycleEventType = 1
	PluginLifecycleEventDisconnected PluginLifecycleEventType = 2
	PluginLifecycleEventTimeout      PluginLifecycleEventType = 3
	PluginLifecycleEventReplaced     PluginLifecycleEventType = 4
)

type ServiceRunStatus int32

const (
	ServiceRunStatusUnspecified ServiceRunStatus = 0
	ServiceRunStatusStarting    ServiceRunStatus = 1
	ServiceRunStatusRunning     ServiceRunStatus = 2
	ServiceRunStatusStopped     ServiceRunStatus = 3
	ServiceRunStatusError       ServiceRunStatus = 4
)

type RigBackend string

const (
	RigBackendHamlib  RigBackend = "Hamlib"
	RigBackendFLRig   RigBackend = "FLRig"
	RigBackendOmniRig RigBackend = "OmniRig"
)

type ControllableWindow string

const (
	WindowSettings     ControllableWindow = "SettingsWindow"
	WindowAbout        ControllableWindow = "AboutWindow"
	WindowQSOAssistant ControllableWindow = "QSOAssistantWindow"
	WindowStationStats ControllableWindow = "StationStatisticWindow"
	WindowPolarChart   ControllableWindow = "PolarChartWindow"
)

type PluginManifest struct {
	UUID              string
	Name              string
	Version           string
	Description       string
	Metadata          map[string]string
	EventSubscription *EventSubscription
	SDKName           string
	SDKVersion        string
}

type EventSubscription struct {
	Topics []EnvelopeTopic
}

type NotificationCommand struct {
	Level   NotificationLevel
	Title   string
	Message string
}

type SettingsPatch struct {
	Values map[string]string
}

type PluginTelemetry struct {
	PluginUUID           string
	ReceivedMessageCount uint64
	SentMessageCount     uint64
	ControlRequestCount  uint64
	ControlErrorCount    uint64
	LastRoundtripMs      uint32
	UpdatedAt            time.Time
}

type RigSnapshot struct {
	Provider       string
	Endpoint       string
	ServiceRunning bool
	RigModel       string
	TXFrequencyHz  uint64
	RXFrequencyHz  uint64
	TXMode         string
	RXMode         string
	Split          bool
	Power          uint32
	SampledAt      time.Time
}

type UDPSnapshot struct {
	ServerRunning bool
	BindAddress   string
	SampledAt     time.Time
}

type QSOQueueSnapshot struct {
	Details   []QSODetail
	SampledAt time.Time
}

type SettingsSnapshot struct {
	InstanceName         string
	Language             string
	EnablePlugin         bool
	DisableAllCharts     bool
	MyMaidenheadGrid     string
	AutoQSOUploadEnabled bool
	AutoRigUploadEnabled bool
	EnableUDPServer      bool
	SampledAt            time.Time
}

type RuntimeSnapshot struct {
	ServerInfo       ServerInfo
	RigSnapshot      RigSnapshot
	UDPSnapshot      UDPSnapshot
	SettingsSnapshot SettingsSnapshot
	PluginTelemetry  []PluginTelemetry
	SampledAt        time.Time
}

type PluginInfo struct {
	UUID              string
	Name              string
	Version           string
	Description       string
	Metadata          map[string]string
	RegisteredAt      time.Time
	LastHeartbeat     time.Time
	EventSubscription EventSubscription
	Telemetry         PluginTelemetry
}

type PluginList struct {
	Plugins []PluginInfo
}

type ServerInfo struct {
	InstanceID           string
	Version              string
	KeepaliveTimeoutSec  uint32
	ConnectedPluginCount uint32
	UptimeSec            uint64
}

type RegisterResponse struct {
	Success    bool
	Message    string
	InstanceID string
	ServerInfo ServerInfo
	Timestamp  time.Time
}

type InboundKind string

const (
	InboundKindUnknown          InboundKind = "unknown"
	InboundKindRigData          InboundKind = "rig_data"
	InboundKindCLHInternal      InboundKind = "clh_internal"
	InboundKindEnvelope         InboundKind = "envelope"
	InboundKindConnectionClosed InboundKind = "connection_closed"
)

type Message struct {
	Kind             InboundKind
	Timestamp        time.Time
	RigData          *RigData
	CLHInternal      *CLHInternalMessage
	Envelope         *Envelope
	ConnectionClosed *ConnectionClosed
	Unknown          *UnknownMessage
}

type UnknownMessage struct {
	TypeURL string
	Raw     []byte
}

type ConnectionClosed struct {
	Timestamp time.Time
}

type Envelope struct {
	ID            string
	CorrelationID string
	Kind          EnvelopeKind
	Topic         EnvelopeTopic
	Success       bool
	Message       string
	ErrorCode     string
	Attributes    map[string]string
	Subscription  *EventSubscription
	Payload       any
	Timestamp     time.Time
}

type RigData struct {
	UUID        string
	Provider    string
	RigName     string
	Frequency   uint64
	Mode        string
	FrequencyRX uint64
	ModeRX      string
	Split       bool
	Power       uint32
	Timestamp   time.Time
}

type WsjtxMessage struct {
	Header              WsjtxMessageHeader
	Heartbeat           *WsjtxHeartbeat
	Status              *WsjtxStatus
	Decode              *WsjtxDecode
	Clear               *WsjtxClear
	Reply               *WsjtxReply
	QSOLogged           *WsjtxQSOLogged
	Close               *WsjtxClose
	HaltTx              *WsjtxHaltTx
	FreeText            *WsjtxFreeText
	WSPRDecode          *WsjtxWSPRDecode
	Location            *WsjtxLocation
	LoggedADIF          *WsjtxLoggedADIF
	HighlightCallsign   *WsjtxHighlightCallsign
	SwitchConfiguration *WsjtxSwitchConfiguration
	Configure           *WsjtxConfigure
	Timestamp           time.Time
}

type WsjtxMessageHeader struct {
	MagicNumber  uint32
	SchemaNumber uint32
	Type         WsjtxMessageType
	ID           string
}

type WsjtxHeartbeat struct {
	MaxSchemaNumber uint32
	Version         string
	Revision        *string
}

type WsjtxStatus struct {
	DialFrequency      uint64
	Mode               string
	DXCall             string
	Report             string
	TXMode             string
	TXEnabled          bool
	Transmitting       bool
	Decoding           bool
	RXDF               uint32
	TXDF               uint32
	DECall             string
	DEGrid             string
	DXGrid             string
	TXWatchdog         bool
	SubMode            string
	FastMode           bool
	SpecialOpMode      *SpecialOperationMode
	FrequencyTolerance *uint32
	TRPeriod           *uint32
	ConfigName         *string
	TXMessage          *string
}

type WsjtxDecode struct {
	IsNew          bool
	Time           time.Time
	SNR            int32
	DeltaTime      float64
	DeltaFrequency uint32
	Mode           string
	Message        string
	LowConfidence  bool
	OffAir         bool
}

type WsjtxClear struct {
	Window ClearWindow
}

type WsjtxReply struct {
	Time           time.Time
	SNR            int32
	DeltaTime      float64
	DeltaFrequency uint32
	Mode           string
	Message        string
	LowConfidence  bool
	Modifiers      uint32
}

type WsjtxQSOLogged struct {
	DateTimeOff         time.Time
	DXCall              string
	DXGrid              string
	TXFrequency         uint64
	Mode                string
	ReportSent          string
	ReportReceived      string
	TXPower             string
	Comments            string
	DateTimeOn          time.Time
	OperatorCall        string
	MyCall              string
	MyGrid              string
	ExchangeSent        *string
	ExchangeReceived    *string
	ADIFPropagationMode *string
}

type WsjtxClose struct{}

type WsjtxHaltTx struct {
	AutoTXOnly bool
}

type WsjtxFreeText struct {
	Text string
	Send bool
}

type WsjtxWSPRDecode struct {
	IsNew     bool
	Time      time.Time
	SNR       int32
	DeltaTime float64
	Frequency uint64
	Drift     int32
	Callsign  string
	Grid      string
	Power     int32
	OffAir    *bool
}

type WsjtxLocation struct {
	Location string
}

type WsjtxLoggedADIF struct {
	ADIFText string
}

type WsjtxHighlightCallsign struct {
	Callsign        string
	BackgroundColor uint32
	ForegroundColor uint32
	HighlightLast   bool
}

type WsjtxSwitchConfiguration struct {
	ConfigName string
}

type WsjtxConfigure struct {
	Mode               string
	FrequencyTolerance uint32
	SubMode            string
	FastMode           bool
	TRPeriod           uint32
	RXDF               uint32
	DXCall             string
	DXGrid             string
	GenerateMessages   bool
}

type PackedDecodeMessage struct {
	Messages  []WsjtxDecode
	Timestamp time.Time
}

type CLHInternalMessage struct {
	QSOUploadStatus *QSOUploadStatusChanged
	PluginLifecycle *PluginLifecycleChanged
	ServerStatus    *ServerStatusChanged
	QSOQueueStatus  *QSOQueueStatusChanged
	SettingsChanged *SettingsChanged
	PluginTelemetry *PluginTelemetryChanged
	Timestamp       time.Time
}

type QSODetail struct {
	UploadedServices             map[string]bool
	UploadedServicesErrorMessage map[string]string
	OriginalCountryName          string
	CQZone                       int32
	ITUZone                      int32
	Continent                    string
	Latitude                     float32
	Longitude                    float32
	GMTOffset                    float32
	DXCC                         string
	DateTimeOff                  time.Time
	DXCall                       string
	DXGrid                       string
	TXFrequencyHz                uint64
	TXFrequencyMeters            string
	Mode                         string
	ParentMode                   string
	ReportSent                   string
	ReportReceived               string
	TXPower                      string
	Comments                     string
	Name                         string
	DateTimeOn                   time.Time
	OperatorCall                 string
	MyCall                       string
	MyGrid                       string
	ExchangeSent                 string
	ExchangeReceived             string
	ADIFPropagationMode          string
	ClientID                     string
	RawData                      string
	FailReason                   string
	UploadStatus                 UploadStatus
	ForcedUpload                 bool
	UUID                         string
}

type QSOUploadStatusChanged struct {
	Detail *QSODetail
}

type PluginLifecycleChanged struct {
	PluginUUID    string
	PluginName    string
	PluginVersion string
	Reason        string
	EventType     PluginLifecycleEventType
	EventTime     time.Time
}

type ServerStatusChanged struct {
	InstanceID           string
	Version              string
	ConnectedPluginCount uint32
	EventTime            time.Time
}

type QSOQueueStatusChanged struct {
	PendingCount  uint32
	UploadedTotal uint64
	FailedTotal   uint64
	EventTime     time.Time
}

type SettingsChanged struct {
	ChangedPart string
	Summary     string
	EventTime   time.Time
}

type PluginTelemetryChanged struct {
	PluginUUID           string
	ReceivedMessageCount uint64
	SentMessageCount     uint64
	ControlRequestCount  uint64
	ControlErrorCount    uint64
	LastRoundtripMs      uint32
	EventTime            time.Time
}
