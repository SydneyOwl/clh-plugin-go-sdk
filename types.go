package pluginsdk

import (
	"time"

	plugin "github.com/SydneyOwl/clh-proto/gen/go/v20260219"
)

type PluginCapability int32

const (
	CapabilityWsjtxMessage       PluginCapability = PluginCapability(plugin.Capability_WSJTX_MESSAGE)
	CapabilityRigData            PluginCapability = PluginCapability(plugin.Capability_RIG_DATA)
	CapabilityClhInternalMessage PluginCapability = PluginCapability(plugin.Capability_CLH_INTERNAL_DATA)
)

// RigData contains radio rig information
type RigData struct {
	UUID        string
	Provider    string
	RigName     string
	Frequency   uint64
	Mode        string
	FrequencyRx uint64
	ModeRx      string
	Split       bool
	Power       uint32
	Timestamp   time.Time
}

// MessageType defines WSJT-X message types
type MessageType int32

const (
	MessageType_HEARTBEAT            MessageType = 0
	MessageType_STATUS               MessageType = 1
	MessageType_DECODE               MessageType = 2
	MessageType_CLEAR                MessageType = 3
	MessageType_REPLY                MessageType = 4
	MessageType_QSO_LOGGED           MessageType = 5
	MessageType_CLOSE                MessageType = 6
	MessageType_REPLAY               MessageType = 7
	MessageType_HALT_TX              MessageType = 8
	MessageType_FREE_TEXT            MessageType = 9
	MessageType_WSPR_DECODE          MessageType = 10
	MessageType_LOCATION             MessageType = 11
	MessageType_LOGGED_ADIF          MessageType = 12
	MessageType_HIGHLIGHT_CALLSIGN   MessageType = 13
	MessageType_SWITCH_CONFIGURATION MessageType = 14
	MessageType_CONFIGURE            MessageType = 15
)

// SpecialOperationMode defines special operation modes
type SpecialOperationMode int32

const (
	SpecialOperationMode_NONE      SpecialOperationMode = 0
	SpecialOperationMode_NA_VHF    SpecialOperationMode = 1
	SpecialOperationMode_EU_VHF    SpecialOperationMode = 2
	SpecialOperationMode_FIELD_DAY SpecialOperationMode = 3
	SpecialOperationMode_RTTY_RU   SpecialOperationMode = 4
	SpecialOperationMode_WW_DIGI   SpecialOperationMode = 5
	SpecialOperationMode_FOX       SpecialOperationMode = 6
	SpecialOperationMode_HOUND     SpecialOperationMode = 7
)

// ClearWindow defines clear window options
type ClearWindow int32

const (
	ClearWindow_CLEAR_BAND_ACTIVITY ClearWindow = 0
	ClearWindow_CLEAR_RX_FREQUENCY  ClearWindow = 1
	ClearWindow_CLEAR_BOTH          ClearWindow = 2
)

// KeyModifiers defines keyboard modifiers
type KeyModifiers uint32

const (
	KeyModifiers_NO_MODIFIER  KeyModifiers = 0x00
	KeyModifiers_SHIFT        KeyModifiers = 0x02
	KeyModifiers_CTRL         KeyModifiers = 0x04
	KeyModifiers_ALT          KeyModifiers = 0x08
	KeyModifiers_META         KeyModifiers = 0x10
	KeyModifiers_KEYPAD       KeyModifiers = 0x20
	KeyModifiers_GROUP_SWITCH KeyModifiers = 0x40
)

// MessageHeader contains common WSJT-X message header
type MessageHeader struct {
	MagicNumber uint32
	SchemaNumer uint32
	Type        MessageType
	ID          string
}

// Heartbeat message (type 0)
type Heartbeat struct {
	MaxSchemaNumer uint32
	Version        string
	Revision       *string
}

// Status message (type 1)
type Status struct {
	DialFrequency      uint64
	Mode               string
	DXCall             string
	Report             string
	TXMode             string
	TXEnabled          bool
	Transmitting       bool
	Decoding           bool
	RxDf               uint32
	TxDf               uint32
	DECall             string
	DEGrid             string
	DXGrid             string
	TxWatchdog         bool
	SubMode            string
	FastMode           bool
	SpecialOpMode      *SpecialOperationMode
	FrequencyTolerance *uint32
	TRPeriod           *uint32
	ConfigName         *string
	TxMessage          *string
}

// Decode message (type 2)
type Decode struct {
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

// Clear message (type 3)
type Clear struct {
	Window ClearWindow
}

// Reply message (type 4)
type Reply struct {
	Time           time.Time
	SNR            int32
	DeltaTime      float64
	DeltaFrequency uint32
	Mode           string
	Message        string
	LowConfidence  bool
	Modifiers      uint32
}

// QSOLogged message (type 5)
type QSOLogged struct {
	DatetimeOff         time.Time
	DXCall              string
	DXGrid              string
	TxFrequency         uint64
	Mode                string
	ReportSent          string
	ReportReceived      string
	TxPower             string
	Comments            string
	DatetimeOn          time.Time
	OperatorCall        string
	MyCall              string
	MyGrid              string
	ExchangeSent        *string
	ExchangeReceived    *string
	ADIFPropagationMode *string
}

// Close message (type 6)
type Close struct{}

// HaltTx message (type 8)
type HaltTx struct {
	AutoTxOnly bool
}

// FreeText message (type 9)
type FreeText struct {
	Text string
	Send bool
}

// WSPRDecode message (type 10)
type WSPRDecode struct {
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

// Location message (type 11)
type Location struct {
	Location string
}

// LoggedADIF message (type 12)
type LoggedADIF struct {
	ADIFText string
}

// HighlightCallsign message (type 13)
type HighlightCallsign struct {
	Callsign        string
	BackgroundColor uint32
	ForegroundColor uint32
	HighlightLast   bool
}

// SwitchConfiguration message (type 14)
type SwitchConfiguration struct {
	ConfigName string
}

// Configure message (type 15)
type Configure struct {
	Mode               string
	FrequencyTolerance uint32
	SubMode            string
	FastMode           bool
	TRPeriod           uint32
	RxDf               uint32
	DXCall             string
	DXGrid             string
	GenerateMessages   bool
}

// WsjtxMessage contains a complete WSJT-X message
type WsjtxMessage struct {
	Header    MessageHeader
	Payload   WsjtxMessagePayload
	Timestamp time.Time
}

// WsjtxMessagePayload is a discriminated union (interface) for message payloads
type WsjtxMessagePayload interface {
	isWsjtxMessagePayload()
}

// Implement isWsjtxMessagePayload for all message types
func (Heartbeat) isWsjtxMessagePayload()           {}
func (Status) isWsjtxMessagePayload()              {}
func (Decode) isWsjtxMessagePayload()              {}
func (Clear) isWsjtxMessagePayload()               {}
func (Reply) isWsjtxMessagePayload()               {}
func (QSOLogged) isWsjtxMessagePayload()           {}
func (Close) isWsjtxMessagePayload()               {}
func (HaltTx) isWsjtxMessagePayload()              {}
func (FreeText) isWsjtxMessagePayload()            {}
func (WSPRDecode) isWsjtxMessagePayload()          {}
func (Location) isWsjtxMessagePayload()            {}
func (LoggedADIF) isWsjtxMessagePayload()          {}
func (HighlightCallsign) isWsjtxMessagePayload()   {}
func (SwitchConfiguration) isWsjtxMessagePayload() {}
func (Configure) isWsjtxMessagePayload()           {}

// PackedDecodeMessage contains multiple WSJT-X messages
type PackedDecodeMessage struct {
	Messages  []*Decode
	Timestamp time.Time
}

// Message is the interface for all messages returned by WaitMessage
type Message interface {
	isMessage()
}

// Implement isMessage for message types
func (RigData) isMessage()             {}
func (WsjtxMessage) isMessage()        {}
func (PackedDecodeMessage) isMessage() {}

// ClhQSOUploadStatusChanged carries internal CLH QSO upload status
type ClhQSOUploadStatusChanged struct {
	UploadedServices             map[string]bool
	UploadedServicesErrorMessage map[string]string
	OriginalCountryName          string
	CqZone                       int32
	ItuZone                      int32
	Continent                    string
	Latitude                     float32
	Longitude                    float32
	GmtOffset                    float32
	Dxcc                         string
	DateTimeOff                  time.Time
	DxCall                       string
	DxGrid                       string
	TxFrequencyInHz              uint64
	TxFrequencyInMeters          string
	Mode                         string
	ParentMode                   string
	ReportSent                   string
	ReportReceived               string
	TxPower                      string
	Comments                     string
	Name                         string
	DateTimeOn                   time.Time
	OperatorCall                 string
	MyCall                       string
	MyGrid                       string
	ExchangeSent                 string
	ExchangeReceived             string
	AdifPropagationMode          string
	ClientId                     string
	RawData                      string
	FailReason                   string
	UploadStatus                 int32
	ForcedUpload                 bool
}

// ClhInternalMessage is an internal representation of CLH internal messages
type ClhInternalMessage struct {
	QsoUploadStatus *ClhQSOUploadStatusChanged
	Timestamp       time.Time
}

func (ClhInternalMessage) isMessage() {}

type PipeConnectionClosed struct {
	Timestamp time.Time
}

func (PipeConnectionClosed) isMessage() {}
