package clhplugin

import (
	"time"

	pb "github.com/SydneyOwl/clh-proto/gen/go/v20260312"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func nowTimestamp() *timestamppb.Timestamp {
	return timestamppb.New(time.Now().UTC())
}

func toTimestamp(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t.UTC())
}

func fromTimestamp(ts *timestamppb.Timestamp) time.Time {
	if ts == nil {
		return time.Time{}
	}
	return ts.AsTime().UTC()
}

func toPBEventSubscription(in *EventSubscription) *pb.PipeEventSubscription {
	if in == nil {
		return nil
	}
	out := &pb.PipeEventSubscription{}
	for _, topic := range in.Topics {
		out.Topics = append(out.Topics, pb.PipeEnvelopeTopic(topic))
	}
	return out
}

func fromPBEventSubscription(in *pb.PipeEventSubscription) EventSubscription {
	if in == nil {
		return EventSubscription{}
	}
	out := EventSubscription{}
	for _, t := range in.Topics {
		out.Topics = append(out.Topics, EnvelopeTopic(t))
	}
	return out
}

func toPBNotification(in NotificationCommand) *pb.PipeNotificationCommand {
	return &pb.PipeNotificationCommand{
		Level:   pb.PipeNotificationLevel(in.Level),
		Title:   in.Title,
		Message: in.Message,
	}
}

func toPBSettingsPatch(in SettingsPatch) *pb.PipeSettingsPatch {
	out := &pb.PipeSettingsPatch{
		Values: map[string]string{},
	}
	for k, v := range in.Values {
		out.Values[k] = v
	}
	return out
}

func toPBManifest(in PluginManifest) *pb.PipeRegisterPluginReq {
	out := &pb.PipeRegisterPluginReq{
		Uuid:              in.UUID,
		Name:              in.Name,
		Version:           in.Version,
		Description:       in.Description,
		Metadata:          map[string]string{},
		SdkName:           in.SDKName,
		SdkVersion:        in.SDKVersion,
		EventSubscription: toPBEventSubscription(in.EventSubscription),
		Timestamp:         nowTimestamp(),
	}
	for k, v := range in.Metadata {
		out.Metadata[k] = v
	}
	return out
}

func fromPBRegisterResponse(in *pb.PipeRegisterPluginResp) RegisterResponse {
	if in == nil {
		return RegisterResponse{}
	}
	return RegisterResponse{
		Success:    in.Success,
		Message:    in.Message,
		InstanceID: in.ClhInstanceId,
		ServerInfo: fromPBServerInfo(in.ServerInfo),
		Timestamp:  fromTimestamp(in.Timestamp),
	}
}

func fromPBServerInfo(in *pb.PipeServerInfo) ServerInfo {
	if in == nil {
		return ServerInfo{}
	}
	return ServerInfo{
		InstanceID:           in.ClhInstanceId,
		Version:              in.ClhVersion,
		KeepaliveTimeoutSec:  in.KeepaliveTimeoutSec,
		ConnectedPluginCount: in.ConnectedPluginCount,
		UptimeSec:            in.UptimeSec,
	}
}

func fromPBPluginTelemetry(in *pb.PipePluginTelemetry) PluginTelemetry {
	if in == nil {
		return PluginTelemetry{}
	}
	return PluginTelemetry{
		PluginUUID:           in.PluginUuid,
		ReceivedMessageCount: in.ReceivedMessageCount,
		SentMessageCount:     in.SentMessageCount,
		ControlRequestCount:  in.ControlRequestCount,
		ControlErrorCount:    in.ControlErrorCount,
		LastRoundtripMs:      in.LastRoundtripMs,
		UpdatedAt:            fromTimestamp(in.UpdatedAt),
	}
}

func fromPBRigSnapshot(in *pb.PipeRigStatusSnapshot) RigSnapshot {
	if in == nil {
		return RigSnapshot{}
	}
	return RigSnapshot{
		Provider:       in.Provider,
		Endpoint:       in.Endpoint,
		ServiceRunning: in.ServiceRunning,
		RigModel:       in.RigModel,
		TXFrequencyHz:  in.TxFrequencyHz,
		RXFrequencyHz:  in.RxFrequencyHz,
		TXMode:         in.TxMode,
		RXMode:         in.RxMode,
		Split:          in.Split,
		Power:          in.Power,
		SampledAt:      fromTimestamp(in.SampledAt),
	}
}

func fromPBUDPSnapshot(in *pb.PipeUdpStatusSnapshot) UDPSnapshot {
	if in == nil {
		return UDPSnapshot{}
	}
	return UDPSnapshot{
		ServerRunning: in.ServerRunning,
		BindAddress:   in.BindAddress,
		SampledAt:     fromTimestamp(in.SampledAt),
	}
}

func fromPBQSOQueueSnapshot(in *pb.PipeQsoQueueSnapshot) QSOQueueSnapshot {
	if in == nil {
		return QSOQueueSnapshot{}
	}
	out := QSOQueueSnapshot{SampledAt: fromTimestamp(in.SampledAt)}
	for _, detail := range in.Details {
		out.Details = append(out.Details, fromPBQSODetail(detail))
	}
	return out
}

func fromPBSettingsSnapshot(in *pb.PipeMainSettingsSnapshot) SettingsSnapshot {
	if in == nil {
		return SettingsSnapshot{}
	}
	return SettingsSnapshot{
		InstanceName:         in.InstanceName,
		Language:             in.Language,
		EnablePlugin:         in.EnablePlugin,
		DisableAllCharts:     in.DisableAllCharts,
		MyMaidenheadGrid:     in.MyMaidenheadGrid,
		AutoQSOUploadEnabled: in.AutoQsoUploadEnabled,
		AutoRigUploadEnabled: in.AutoRigUploadEnabled,
		EnableUDPServer:      in.EnableUdpServer,
		SampledAt:            fromTimestamp(in.SampledAt),
	}
}

func fromPBRuntimeSnapshot(in *pb.PipeRuntimeSnapshot) RuntimeSnapshot {
	if in == nil {
		return RuntimeSnapshot{}
	}
	out := RuntimeSnapshot{
		ServerInfo:       fromPBServerInfo(in.ServerInfo),
		RigSnapshot:      fromPBRigSnapshot(in.RigSnapshot),
		UDPSnapshot:      fromPBUDPSnapshot(in.UdpSnapshot),
		SettingsSnapshot: fromPBSettingsSnapshot(in.SettingsSnapshot),
		SampledAt:        fromTimestamp(in.SampledAt),
	}
	for _, item := range in.PluginTelemetry {
		out.PluginTelemetry = append(out.PluginTelemetry, fromPBPluginTelemetry(item))
	}
	return out
}

func fromPBPluginInfo(in *pb.PipePluginInfo) PluginInfo {
	if in == nil {
		return PluginInfo{}
	}
	out := PluginInfo{
		UUID:              in.Uuid,
		Name:              in.Name,
		Version:           in.Version,
		Description:       in.Description,
		Metadata:          map[string]string{},
		RegisteredAt:      fromTimestamp(in.RegisteredAt),
		LastHeartbeat:     fromTimestamp(in.LastHeartbeat),
		EventSubscription: fromPBEventSubscription(in.EventSubscription),
		Telemetry:         fromPBPluginTelemetry(in.Telemetry),
	}
	for k, v := range in.Metadata {
		out.Metadata[k] = v
	}
	return out
}

func fromPBPluginList(in *pb.PipePluginList) PluginList {
	if in == nil {
		return PluginList{}
	}
	out := PluginList{}
	for _, item := range in.Plugins {
		out.Plugins = append(out.Plugins, fromPBPluginInfo(item))
	}
	return out
}

func fromPBRigData(in *pb.RigData) RigData {
	if in == nil {
		return RigData{}
	}
	return RigData{
		UUID:        in.Uuid,
		Provider:    in.Provider,
		RigName:     in.RigName,
		Frequency:   in.Frequency,
		Mode:        in.Mode,
		FrequencyRX: in.FrequencyRx,
		ModeRX:      in.ModeRx,
		Split:       in.Split,
		Power:       in.Power,
		Timestamp:   fromTimestamp(in.Timestamp),
	}
}

func fromPBInternal(in *pb.ClhInternalMessage) CLHInternalMessage {
	if in == nil {
		return CLHInternalMessage{}
	}
	out := CLHInternalMessage{
		Timestamp: fromTimestamp(in.Timestamp),
	}
	switch payload := in.Payload.(type) {
	case *pb.ClhInternalMessage_QsoUploadStatus:
		out.QSOUploadStatus = fromPBQSOUploadStatus(payload.QsoUploadStatus)
	case *pb.ClhInternalMessage_PluginLifecycle:
		out.PluginLifecycle = fromPBPluginLifecycle(payload.PluginLifecycle)
	case *pb.ClhInternalMessage_ServerStatus:
		out.ServerStatus = fromPBServerStatusChanged(payload.ServerStatus)
	case *pb.ClhInternalMessage_QsoQueueStatus:
		out.QSOQueueStatus = fromPBQSOQueueStatus(payload.QsoQueueStatus)
	case *pb.ClhInternalMessage_SettingsChanged:
		out.SettingsChanged = fromPBSettingsChanged(payload.SettingsChanged)
	case *pb.ClhInternalMessage_PluginTelemetry:
		out.PluginTelemetry = fromPBPluginTelemetryChanged(payload.PluginTelemetry)
	}
	return out
}

func fromPBQSOUploadStatus(in *pb.ClhQSOUploadStatusChanged) *QSOUploadStatusChanged {
	if in == nil {
		return nil
	}
	return &QSOUploadStatusChanged{
		Detail: ptr(fromPBQSODetail(in.Detail)),
	}
}

func fromPBQSODetail(in *pb.ClhQSODetail) QSODetail {
	if in == nil {
		return QSODetail{}
	}
	out := QSODetail{
		UploadedServices:             map[string]bool{},
		UploadedServicesErrorMessage: map[string]string{},
		OriginalCountryName:          in.OriginalCountryName,
		CQZone:                       in.CqZone,
		ITUZone:                      in.ItuZone,
		Continent:                    in.Continent,
		Latitude:                     in.Latitude,
		Longitude:                    in.Longitude,
		GMTOffset:                    in.GmtOffset,
		DXCC:                         in.Dxcc,
		DateTimeOff:                  fromTimestamp(in.DateTimeOff),
		DXCall:                       in.DxCall,
		DXGrid:                       in.DxGrid,
		TXFrequencyHz:                in.TxFrequencyInHz,
		TXFrequencyMeters:            in.TxFrequencyInMeters,
		Mode:                         in.Mode,
		ParentMode:                   in.ParentMode,
		ReportSent:                   in.ReportSent,
		ReportReceived:               in.ReportReceived,
		TXPower:                      in.TxPower,
		Comments:                     in.Comments,
		Name:                         in.Name,
		DateTimeOn:                   fromTimestamp(in.DateTimeOn),
		OperatorCall:                 in.OperatorCall,
		MyCall:                       in.MyCall,
		MyGrid:                       in.MyGrid,
		ExchangeSent:                 in.ExchangeSent,
		ExchangeReceived:             in.ExchangeReceived,
		ADIFPropagationMode:          in.AdifPropagationMode,
		ClientID:                     in.ClientId,
		RawData:                      in.RawData,
		FailReason:                   in.FailReason,
		UploadStatus:                 UploadStatus(in.UploadStatus),
		ForcedUpload:                 in.ForcedUpload,
		UUID:                         in.Uuid,
	}
	for k, v := range in.UploadedServices {
		out.UploadedServices[k] = v
	}
	for k, v := range in.UploadedServicesErrorMessage {
		out.UploadedServicesErrorMessage[k] = v
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}

func fromPBPluginLifecycle(in *pb.ClhPluginLifecycleChanged) *PluginLifecycleChanged {
	if in == nil {
		return nil
	}
	return &PluginLifecycleChanged{
		PluginUUID:    in.PluginUuid,
		PluginName:    in.PluginName,
		PluginVersion: in.PluginVersion,
		Reason:        in.Reason,
		EventType:     PluginLifecycleEventType(in.EventType),
		EventTime:     fromTimestamp(in.EventTime),
	}
}

func fromPBServerStatusChanged(in *pb.ClhServerStatusChanged) *ServerStatusChanged {
	if in == nil {
		return nil
	}
	return &ServerStatusChanged{
		InstanceID:           in.ClhInstanceId,
		Version:              in.ClhVersion,
		ConnectedPluginCount: in.ConnectedPluginCount,
		EventTime:            fromTimestamp(in.EventTime),
	}
}

func fromPBQSOQueueStatus(in *pb.ClhQsoQueueStatusChanged) *QSOQueueStatusChanged {
	if in == nil {
		return nil
	}
	return &QSOQueueStatusChanged{
		PendingCount:  in.PendingCount,
		UploadedTotal: in.UploadedTotal,
		FailedTotal:   in.FailedTotal,
		EventTime:     fromTimestamp(in.EventTime),
	}
}

func fromPBSettingsChanged(in *pb.ClhSettingsChanged) *SettingsChanged {
	if in == nil {
		return nil
	}
	return &SettingsChanged{
		ChangedPart: in.ChangedPart,
		Summary:     in.Summary,
		EventTime:   fromTimestamp(in.EventTime),
	}
}

func fromPBPluginTelemetryChanged(in *pb.ClhPluginTelemetryChanged) *PluginTelemetryChanged {
	if in == nil {
		return nil
	}
	return &PluginTelemetryChanged{
		PluginUUID:           in.PluginUuid,
		ReceivedMessageCount: in.ReceivedMessageCount,
		SentMessageCount:     in.SentMessageCount,
		ControlRequestCount:  in.ControlRequestCount,
		ControlErrorCount:    in.ControlErrorCount,
		LastRoundtripMs:      in.LastRoundtripMs,
		EventTime:            fromTimestamp(in.EventTime),
	}
}

func fromPBWsjtxMessage(in *pb.WsjtxMessage) WsjtxMessage {
	if in == nil {
		return WsjtxMessage{}
	}
	out := WsjtxMessage{
		Header: WsjtxMessageHeader{
			MagicNumber:  in.GetHeader().GetMagicNumber(),
			SchemaNumber: in.GetHeader().GetSchemaNumber(),
			Type:         WsjtxMessageType(in.GetHeader().GetType()),
			ID:           in.GetHeader().GetId(),
		},
		Timestamp: fromTimestamp(in.Timestamp),
	}

	if hb := in.GetHeartbeat(); hb != nil {
		out.Heartbeat = &WsjtxHeartbeat{
			MaxSchemaNumber: hb.MaxSchemaNumber,
			Version:         hb.Version,
			Revision:        hb.Revision,
		}
	}
	if st := in.GetStatus(); st != nil {
		out.Status = &WsjtxStatus{
			DialFrequency:      st.DialFrequency,
			Mode:               st.Mode,
			DXCall:             st.DxCall,
			Report:             st.Report,
			TXMode:             st.TxMode,
			TXEnabled:          st.TxEnabled,
			Transmitting:       st.Transmitting,
			Decoding:           st.Decoding,
			RXDF:               st.RxDf,
			TXDF:               st.TxDf,
			DECall:             st.DeCall,
			DEGrid:             st.DeGrid,
			DXGrid:             st.DxGrid,
			TXWatchdog:         st.TxWatchdog,
			SubMode:            st.SubMode,
			FastMode:           st.FastMode,
			FrequencyTolerance: st.FrequencyTolerance,
			TRPeriod:           st.TrPeriod,
			ConfigName:         st.ConfigName,
			TXMessage:          st.TxMessage,
		}
		if st.SpecialOpMode != nil {
			mode := SpecialOperationMode(*st.SpecialOpMode)
			out.Status.SpecialOpMode = &mode
		}
	}
	if d := in.GetDecode(); d != nil {
		out.Decode = &WsjtxDecode{
			IsNew:          d.IsNew,
			Time:           fromTimestamp(d.Time),
			SNR:            d.Snr,
			DeltaTime:      d.DeltaTime,
			DeltaFrequency: d.DeltaFrequency,
			Mode:           d.Mode,
			Message:        d.Message,
			LowConfidence:  d.LowConfidence,
			OffAir:         d.OffAir,
		}
	}
	if cl := in.GetClear(); cl != nil {
		out.Clear = &WsjtxClear{Window: ClearWindow(cl.Window)}
	}
	if rep := in.GetReply(); rep != nil {
		out.Reply = &WsjtxReply{
			Time:           fromTimestamp(rep.Time),
			SNR:            rep.Snr,
			DeltaTime:      rep.DeltaTime,
			DeltaFrequency: rep.DeltaFrequency,
			Mode:           rep.Mode,
			Message:        rep.Message,
			LowConfidence:  rep.LowConfidence,
			Modifiers:      rep.Modifiers,
		}
	}
	if qso := in.GetQsoLogged(); qso != nil {
		out.QSOLogged = &WsjtxQSOLogged{
			DateTimeOff:         fromTimestamp(qso.DatetimeOff),
			DXCall:              qso.DxCall,
			DXGrid:              qso.DxGrid,
			TXFrequency:         qso.TxFrequency,
			Mode:                qso.Mode,
			ReportSent:          qso.ReportSent,
			ReportReceived:      qso.ReportReceived,
			TXPower:             qso.TxPower,
			Comments:            qso.Comments,
			DateTimeOn:          fromTimestamp(qso.DatetimeOn),
			OperatorCall:        qso.OperatorCall,
			MyCall:              qso.MyCall,
			MyGrid:              qso.MyGrid,
			ExchangeSent:        qso.ExchangeSent,
			ExchangeReceived:    qso.ExchangeReceived,
			ADIFPropagationMode: qso.AdifPropagationMode,
		}
	}
	if in.GetClose() != nil {
		out.Close = &WsjtxClose{}
	}
	if halt := in.GetHaltTx(); halt != nil {
		out.HaltTx = &WsjtxHaltTx{AutoTXOnly: halt.AutoTxOnly}
	}
	if free := in.GetFreeText(); free != nil {
		out.FreeText = &WsjtxFreeText{Text: free.Text, Send: free.Send}
	}
	if w := in.GetWsprDecode(); w != nil {
		out.WSPRDecode = &WsjtxWSPRDecode{
			IsNew:     w.IsNew,
			Time:      fromTimestamp(w.Time),
			SNR:       w.Snr,
			DeltaTime: w.DeltaTime,
			Frequency: w.Frequency,
			Drift:     w.Drift,
			Callsign:  w.Callsign,
			Grid:      w.Grid,
			Power:     w.Power,
			OffAir:    w.OffAir,
		}
	}
	if loc := in.GetLocation(); loc != nil {
		out.Location = &WsjtxLocation{Location: loc.Location}
	}
	if adif := in.GetLoggedAdif(); adif != nil {
		out.LoggedADIF = &WsjtxLoggedADIF{ADIFText: adif.AdifText}
	}
	if hl := in.GetHighlightCallsign(); hl != nil {
		out.HighlightCallsign = &WsjtxHighlightCallsign{
			Callsign:        hl.Callsign,
			BackgroundColor: hl.BackgroundColor,
			ForegroundColor: hl.ForegroundColor,
			HighlightLast:   hl.HighlightLast,
		}
	}
	if sw := in.GetSwitchConfiguration(); sw != nil {
		out.SwitchConfiguration = &WsjtxSwitchConfiguration{ConfigName: sw.ConfigName}
	}
	if cfg := in.GetConfigure(); cfg != nil {
		out.Configure = &WsjtxConfigure{
			Mode:               cfg.Mode,
			FrequencyTolerance: cfg.FrequencyTolerance,
			SubMode:            cfg.SubMode,
			FastMode:           cfg.FastMode,
			TRPeriod:           cfg.TrPeriod,
			RXDF:               cfg.RxDf,
			DXCall:             cfg.DxCall,
			DXGrid:             cfg.DxGrid,
			GenerateMessages:   cfg.GenerateMessages,
		}
	}
	return out
}

func fromPBPackedDecode(in *pb.PackedDecodeMessage) PackedDecodeMessage {
	if in == nil {
		return PackedDecodeMessage{}
	}
	out := PackedDecodeMessage{
		Timestamp: fromTimestamp(in.Timestamp),
	}
	for _, item := range in.Messages {
		out.Messages = append(out.Messages, WsjtxDecode{
			IsNew:          item.IsNew,
			Time:           fromTimestamp(item.Time),
			SNR:            item.Snr,
			DeltaTime:      item.DeltaTime,
			DeltaFrequency: item.DeltaFrequency,
			Mode:           item.Mode,
			Message:        item.Message,
			LowConfidence:  item.LowConfidence,
			OffAir:         item.OffAir,
		})
	}
	return out
}

func fromPBEnvelope(in *pb.PipeEnvelope) Envelope {
	if in == nil {
		return Envelope{}
	}
	out := Envelope{
		ID:            in.Id,
		CorrelationID: in.CorrelationId,
		Kind:          EnvelopeKind(in.Kind),
		Topic:         EnvelopeTopic(in.Topic),
		Success:       in.Success,
		Message:       in.Message,
		ErrorCode:     in.ErrorCode,
		Attributes:    map[string]string{},
		Timestamp:     fromTimestamp(in.Timestamp),
	}
	for k, v := range in.Attributes {
		out.Attributes[k] = v
	}
	if in.Subscription != nil {
		sub := fromPBEventSubscription(in.Subscription)
		out.Subscription = &sub
	}
	if in.Payload != nil {
		out.Payload = decodeEnvelopePayload(in.Payload)
	}
	return out
}

func decodeEnvelopePayload(payload *anypb.Any) any {
	if payload == nil {
		return nil
	}
	msg, err := anypb.UnmarshalNew(payload, proto.UnmarshalOptions{})
	if err != nil {
		return &UnknownMessage{
			TypeURL: payload.TypeUrl,
			Raw:     append([]byte(nil), payload.Value...),
		}
	}
	return convertPayloadMessage(msg)
}

func convertPayloadMessage(msg proto.Message) any {
	switch typed := msg.(type) {
	case *pb.PipeServerInfo:
		model := fromPBServerInfo(typed)
		return model
	case *pb.PipePluginList:
		model := fromPBPluginList(typed)
		return model
	case *pb.PipeRuntimeSnapshot:
		model := fromPBRuntimeSnapshot(typed)
		return model
	case *pb.PipeRigStatusSnapshot:
		model := fromPBRigSnapshot(typed)
		return model
	case *pb.PipeUdpStatusSnapshot:
		model := fromPBUDPSnapshot(typed)
		return model
	case *pb.PipeQsoQueueSnapshot:
		model := fromPBQSOQueueSnapshot(typed)
		return model
	case *pb.PipeMainSettingsSnapshot:
		model := fromPBSettingsSnapshot(typed)
		return model
	case *pb.PipePluginTelemetry:
		model := fromPBPluginTelemetry(typed)
		return model
	case *pb.PipeEventSubscription:
		model := fromPBEventSubscription(typed)
		return model
	case *pb.ClhServerStatusChanged:
		return fromPBServerStatusChanged(typed)
	case *pb.ClhPluginLifecycleChanged:
		return fromPBPluginLifecycle(typed)
	case *pb.ClhQSOUploadStatusChanged:
		return fromPBQSOUploadStatus(typed)
	case *pb.ClhQsoQueueStatusChanged:
		return fromPBQSOQueueStatus(typed)
	case *pb.ClhSettingsChanged:
		return fromPBSettingsChanged(typed)
	case *pb.ClhPluginTelemetryChanged:
		return fromPBPluginTelemetryChanged(typed)
	case *pb.WsjtxMessage:
		model := fromPBWsjtxMessage(typed)
		return model
	case *pb.PackedDecodeMessage:
		model := fromPBPackedDecode(typed)
		return model
	case *pb.RigData:
		model := fromPBRigData(typed)
		return model
	case *pb.ClhInternalMessage:
		model := fromPBInternal(typed)
		return model
	default:
		raw, _ := proto.Marshal(msg)
		return &UnknownMessage{
			TypeURL: string(msg.ProtoReflect().Descriptor().FullName()),
			Raw:     raw,
		}
	}
}

func fromAnyMessage(anyMsg *anypb.Any) (proto.Message, Message, error) {
	msg, err := anypb.UnmarshalNew(anyMsg, proto.UnmarshalOptions{})
	if err != nil {
		return nil, Message{
			Kind: InboundKindUnknown,
			Unknown: &UnknownMessage{
				TypeURL: anyMsg.GetTypeUrl(),
				Raw:     append([]byte(nil), anyMsg.GetValue()...),
			},
		}, err
	}

	out := Message{Kind: InboundKindUnknown}
	switch typed := msg.(type) {
	case *pb.RigData:
		m := fromPBRigData(typed)
		out.Kind = InboundKindRigData
		out.Timestamp = m.Timestamp
		out.RigData = &m
	case *pb.ClhInternalMessage:
		m := fromPBInternal(typed)
		out.Kind = InboundKindCLHInternal
		out.Timestamp = m.Timestamp
		out.CLHInternal = &m
	case *pb.PipeEnvelope:
		m := fromPBEnvelope(typed)
		out.Kind = InboundKindEnvelope
		out.Timestamp = m.Timestamp
		out.Envelope = &m
	case *pb.PipeConnectionClosed:
		m := ConnectionClosed{Timestamp: fromTimestamp(typed.Timestamp)}
		out.Kind = InboundKindConnectionClosed
		out.Timestamp = m.Timestamp
		out.ConnectionClosed = &m
	default:
		out.Kind = InboundKindUnknown
		out.Unknown = &UnknownMessage{
			TypeURL: anyMsg.GetTypeUrl(),
			Raw:     append([]byte(nil), anyMsg.GetValue()...),
		}
	}
	return msg, out, nil
}
