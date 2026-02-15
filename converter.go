package pluginsdk

import (
	"time"

	pbplugin "github.com/sydneyowl/clh-plugin-go-sdk/clh-proto/gen/go"
)

// convertRigData converts protobuf RigData to internal RigData
func convertRigData(pb *pbplugin.RigData) RigData {
	timestamp := time.Time{}
	if pb.Timestamp != nil {
		timestamp = pb.Timestamp.AsTime()
	}

	return RigData{
		UUID:        pb.Uuid,
		Provider:    pb.Provider,
		RigName:     pb.RigName,
		Frequency:   pb.Frequency,
		Mode:        pb.Mode,
		FrequencyRx: pb.FrequencyRx,
		ModeRx:      pb.ModeRx,
		Split:       pb.Split,
		Power:       pb.Power,
		Timestamp:   timestamp,
	}
}

// convertWsjtxMessage converts protobuf WsjtxMessage to internal WsjtxMessage
func convertWsjtxMessage(pb *pbplugin.WsjtxMessage) WsjtxMessage {
	msg := WsjtxMessage{
		Header: MessageHeader{
			MagicNumber: pb.Header.MagicNumber,
			SchemaNumer: pb.Header.SchemaNumber,
			Type:        MessageType(pb.Header.Type),
			ID:          pb.Header.Id,
		},
	}

	if pb.Timestamp != nil {
		msg.Timestamp = pb.Timestamp.AsTime()
	}

	// Convert payload based on type
	switch payload := pb.Payload.(type) {
	case *pbplugin.WsjtxMessage_Heartbeat:
		msg.Payload = convertHeartbeat(payload.Heartbeat)
	case *pbplugin.WsjtxMessage_Status:
		msg.Payload = convertStatus(payload.Status)
	case *pbplugin.WsjtxMessage_Decode:
		msg.Payload = convertDecode(payload.Decode)
	case *pbplugin.WsjtxMessage_Clear:
		msg.Payload = convertClear(payload.Clear)
	case *pbplugin.WsjtxMessage_Reply:
		msg.Payload = convertReply(payload.Reply)
	case *pbplugin.WsjtxMessage_QsoLogged:
		msg.Payload = convertQSOLogged(payload.QsoLogged)
	case *pbplugin.WsjtxMessage_Close:
		msg.Payload = convertClose(payload.Close)
	case *pbplugin.WsjtxMessage_HaltTx:
		msg.Payload = convertHaltTx(payload.HaltTx)
	case *pbplugin.WsjtxMessage_FreeText:
		msg.Payload = convertFreeText(payload.FreeText)
	case *pbplugin.WsjtxMessage_WsprDecode:
		msg.Payload = convertWSPRDecode(payload.WsprDecode)
	case *pbplugin.WsjtxMessage_Location:
		msg.Payload = convertLocation(payload.Location)
	case *pbplugin.WsjtxMessage_LoggedAdif:
		msg.Payload = convertLoggedADIF(payload.LoggedAdif)
	case *pbplugin.WsjtxMessage_HighlightCallsign:
		msg.Payload = convertHighlightCallsign(payload.HighlightCallsign)
	case *pbplugin.WsjtxMessage_SwitchConfiguration:
		msg.Payload = convertSwitchConfiguration(payload.SwitchConfiguration)
	case *pbplugin.WsjtxMessage_Configure:
		msg.Payload = convertConfigure(payload.Configure)
	}

	return msg
}

// convertHeartbeat converts protobuf Heartbeat to internal Heartbeat
func convertHeartbeat(pb *pbplugin.Heartbeat) Heartbeat {
	var revision *string
	if pb.Revision != nil {
		revision = pb.Revision
	}

	return Heartbeat{
		MaxSchemaNumer: pb.MaxSchemaNumber,
		Version:        pb.Version,
		Revision:       revision,
	}
}

// convertStatus converts protobuf Status to internal Status
func convertStatus(pb *pbplugin.Status) Status {
	status := Status{
		DialFrequency: pb.DialFrequency,
		Mode:          pb.Mode,
		DXCall:        pb.DxCall,
		Report:        pb.Report,
		TXMode:        pb.TxMode,
		TXEnabled:     pb.TxEnabled,
		Transmitting:  pb.Transmitting,
		Decoding:      pb.Decoding,
		RxDf:          pb.RxDf,
		TxDf:          pb.TxDf,
		DECall:        pb.DeCall,
		DEGrid:        pb.DeGrid,
		DXGrid:        pb.DxGrid,
		TxWatchdog:    pb.TxWatchdog,
		SubMode:       pb.SubMode,
		FastMode:      pb.FastMode,
	}

	if pb.SpecialOpMode != nil {
		mode := SpecialOperationMode(*pb.SpecialOpMode)
		status.SpecialOpMode = &mode
	}

	if pb.FrequencyTolerance != nil {
		status.FrequencyTolerance = pb.FrequencyTolerance
	}

	if pb.TrPeriod != nil {
		status.TRPeriod = pb.TrPeriod
	}

	if pb.ConfigName != nil {
		status.ConfigName = pb.ConfigName
	}

	if pb.TxMessage != nil {
		status.TxMessage = pb.TxMessage
	}

	return status
}

// convertDecode converts protobuf Decode to internal Decode
func convertDecode(pb *pbplugin.Decode) Decode {
	decodeTime := time.Time{}
	if pb.Time != nil {
		decodeTime = pb.Time.AsTime()
	}

	return Decode{
		IsNew:          pb.IsNew,
		Time:           decodeTime,
		SNR:            pb.Snr,
		DeltaTime:      pb.DeltaTime,
		DeltaFrequency: pb.DeltaFrequency,
		Mode:           pb.Mode,
		Message:        pb.Message,
		LowConfidence:  pb.LowConfidence,
		OffAir:         pb.OffAir,
	}
}

// convertClear converts protobuf Clear to internal Clear
func convertClear(pb *pbplugin.Clear) Clear {
	return Clear{
		Window: ClearWindow(pb.Window),
	}
}

// convertReply converts protobuf Reply to internal Reply
func convertReply(pb *pbplugin.Reply) Reply {
	replyTime := time.Time{}
	if pb.Time != nil {
		replyTime = pb.Time.AsTime()
	}

	return Reply{
		Time:           replyTime,
		SNR:            pb.Snr,
		DeltaTime:      pb.DeltaTime,
		DeltaFrequency: pb.DeltaFrequency,
		Mode:           pb.Mode,
		Message:        pb.Message,
		LowConfidence:  pb.LowConfidence,
		Modifiers:      pb.Modifiers,
	}
}

// convertQSOLogged converts protobuf QSOLogged to internal QSOLogged
func convertQSOLogged(pb *pbplugin.QSOLogged) QSOLogged {
	datetimeOff := time.Time{}
	if pb.DatetimeOff != nil {
		datetimeOff = pb.DatetimeOff.AsTime()
	}

	datetimeOn := time.Time{}
	if pb.DatetimeOn != nil {
		datetimeOn = pb.DatetimeOn.AsTime()
	}

	qso := QSOLogged{
		DatetimeOff:    datetimeOff,
		DXCall:         pb.DxCall,
		DXGrid:         pb.DxGrid,
		TxFrequency:    pb.TxFrequency,
		Mode:           pb.Mode,
		ReportSent:     pb.ReportSent,
		ReportReceived: pb.ReportReceived,
		TxPower:        pb.TxPower,
		Comments:       pb.Comments,
		DatetimeOn:     datetimeOn,
		OperatorCall:   pb.OperatorCall,
		MyCall:         pb.MyCall,
		MyGrid:         pb.MyGrid,
	}

	if pb.ExchangeSent != nil {
		qso.ExchangeSent = pb.ExchangeSent
	}

	if pb.ExchangeReceived != nil {
		qso.ExchangeReceived = pb.ExchangeReceived
	}

	if pb.AdifPropagationMode != nil {
		qso.ADIFPropagationMode = pb.AdifPropagationMode
	}

	return qso
}

// convertClose converts protobuf Close to internal Close
func convertClose(pb *pbplugin.Close) Close {
	return Close{}
}

// convertHaltTx converts protobuf HaltTx to internal HaltTx
func convertHaltTx(pb *pbplugin.HaltTx) HaltTx {
	return HaltTx{
		AutoTxOnly: pb.AutoTxOnly,
	}
}

// convertFreeText converts protobuf FreeText to internal FreeText
func convertFreeText(pb *pbplugin.FreeText) FreeText {
	return FreeText{
		Text: pb.Text,
		Send: pb.Send,
	}
}

// convertWSPRDecode converts protobuf WSPRDecode to internal WSPRDecode
func convertWSPRDecode(pb *pbplugin.WSPRDecode) WSPRDecode {
	decodeTime := time.Time{}
	if pb.Time != nil {
		decodeTime = pb.Time.AsTime()
	}

	wspr := WSPRDecode{
		IsNew:     pb.IsNew,
		Time:      decodeTime,
		SNR:       pb.Snr,
		DeltaTime: pb.DeltaTime,
		Frequency: pb.Frequency,
		Drift:     pb.Drift,
		Callsign:  pb.Callsign,
		Grid:      pb.Grid,
		Power:     pb.Power,
	}

	if pb.OffAir != nil {
		wspr.OffAir = pb.OffAir
	}

	return wspr
}

// convertLocation converts protobuf Location to internal Location
func convertLocation(pb *pbplugin.Location) Location {
	return Location{
		Location: pb.Location,
	}
}

// convertLoggedADIF converts protobuf LoggedADIF to internal LoggedADIF
func convertLoggedADIF(pb *pbplugin.LoggedADIF) LoggedADIF {
	return LoggedADIF{
		ADIFText: pb.AdifText,
	}
}

// convertHighlightCallsign converts protobuf HighlightCallsign to internal HighlightCallsign
func convertHighlightCallsign(pb *pbplugin.HighlightCallsign) HighlightCallsign {
	return HighlightCallsign{
		Callsign:        pb.Callsign,
		BackgroundColor: pb.BackgroundColor,
		ForegroundColor: pb.ForegroundColor,
		HighlightLast:   pb.HighlightLast,
	}
}

// convertSwitchConfiguration converts protobuf SwitchConfiguration to internal SwitchConfiguration
func convertSwitchConfiguration(pb *pbplugin.SwitchConfiguration) SwitchConfiguration {
	return SwitchConfiguration{
		ConfigName: pb.ConfigName,
	}
}

// convertConfigure converts protobuf Configure to internal Configure
func convertConfigure(pb *pbplugin.Configure) Configure {
	return Configure{
		Mode:               pb.Mode,
		FrequencyTolerance: pb.FrequencyTolerance,
		SubMode:            pb.SubMode,
		FastMode:           pb.FastMode,
		TRPeriod:           pb.TrPeriod,
		RxDf:               pb.RxDf,
		DXCall:             pb.DxCall,
		DXGrid:             pb.DxGrid,
		GenerateMessages:   pb.GenerateMessages,
	}
}

// convertPackedWsjtxMessage converts protobuf PackedWsjtxMessage to internal PackedWsjtxMessage
func convertPackedWsjtxMessage(pb *pbplugin.PackedWsjtxMessage) PackedWsjtxMessage {
	timestamp := time.Time{}
	if pb.Timestamp != nil {
		timestamp = pb.Timestamp.AsTime()
	}

	messages := make([]*WsjtxMessage, len(pb.Messages))
	for i, msg := range pb.Messages {
		converted := convertWsjtxMessage(msg)
		messages[i] = &converted
	}

	return PackedWsjtxMessage{
		Messages:  messages,
		Timestamp: timestamp,
	}
}
