package events

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/ma111e/melody/internal/logdata"

	"github.com/rs/xid"

	"github.com/google/gopacket/layers"

	"github.com/ma111e/melody/internal/sessions"

	"github.com/ma111e/melody/internal/config"

	"github.com/google/gopacket"
	"github.com/ma111e/melody/internal/httpparser"
)

// HTTPEvent describes the structure of an event generated by a reassembled HTTP packet
type HTTPEvent struct {
	Verb          string            `json:"verb"`
	Proto         string            `json:"proto"`
	RequestURI    string            `json:"URI"`
	SourcePort    uint16            `json:"src_port"`
	DestHost      string            `json:"dst_host"`
	DestPort      uint16            `json:"dst_port"`
	Headers       map[string]string `json:"headers"`
	HeadersKeys   []string          `json:"headers_keys"`
	HeadersValues []string          `json:"headers_values"`
	InlineHeaders []string
	Errors        []string        `json:"errors"`
	Body          logdata.Payload `json:"body"`
	IsTLS         bool            `json:"is_tls"`
	Req           *http.Request
	LogData       logdata.HTTPEventLog
	BaseEvent
}

// GetIPHeader satisfies the Event interface by returning nil. As they're application-level data, HTTP events
// does not support IP header data
func (ev HTTPEvent) GetIPHeader() *layers.IPv4 {
	return nil
}

// GetHTTPData returns the event's data
func (ev HTTPEvent) GetHTTPData() HTTPEvent {
	return ev
}

// ToLog parses the event structure and generate an EventLog almost ready to be sent to the logging file
func (ev HTTPEvent) ToLog() EventLog {
	ev.LogData = logdata.HTTPEventLog{}
	ev.LogData.Timestamp = time.Now().Format(time.RFC3339Nano)
	//ev.LogData.NsTimestamp = strconv.FormatInt(time.Now().UnixNano(), 10)
	//ev.LogData.Type = ev.Kind
	//ev.LogData.SourceIP = ev.SourceIP
	//ev.LogData.DestPort = ev.DestPort
	//ev.LogData.Session = ev.Session
	//
	//if len(ev.Tags) == 0 {
	//	ev.LogData.Tags = make(map[string][]string)
	//} else {
	//	ev.LogData.Tags = ev.Tags
	//}

	ev.LogData.Init(ev.BaseEvent)

	ev.LogData.Session = ev.Session
	ev.LogData.HTTP.Verb = ev.Verb
	ev.LogData.HTTP.Proto = ev.Proto
	ev.LogData.HTTP.RequestURI = ev.RequestURI
	ev.LogData.HTTP.SourcePort = ev.SourcePort
	ev.LogData.HTTP.DestHost = ev.DestHost
	ev.LogData.DestPort = ev.DestPort
	ev.LogData.SourceIP = ev.SourceIP
	ev.LogData.HTTP.Headers = ev.Headers
	ev.LogData.HTTP.Body = ev.Body
	ev.LogData.HTTP.IsTLS = ev.IsTLS
	ev.LogData.Additional = ev.Additional

	if val, ok := ev.Headers["User-Agent"]; ok {
		ev.LogData.HTTP.UserAgent = val
	}

	var headersKeys []string
	var headersValues []string

	for key, val := range ev.Headers {
		headersKeys = append(headersKeys, key)
		headersValues = append(headersValues, val)
	}

	ev.LogData.HTTP.HeadersKeys = headersKeys
	ev.LogData.HTTP.HeadersValues = headersValues

	return ev.LogData
}

// NewHTTPEvent creates an HTTPEvent from a reassembled http.Request. It uses flow information if available to allow
// quality source and destination information. Only available to HTTP events, as HTTPS events are generated from a
// webserver and thus not reassembled
func NewHTTPEvent(r *http.Request, network gopacket.Flow, transport gopacket.Flow) (*HTTPEvent, error) {
	headers := make(map[string]string)
	var inlineHeaders []string
	var errs []string
	var params []byte
	var err error

	for header := range r.Header {
		headers[header] = r.Header.Get(header)
		inlineHeaders = append(inlineHeaders, header+": "+r.Header.Get(header))
	}

	dstPort, _ := strconv.ParseUint(transport.Dst().String(), 10, 16)
	srcPort, _ := strconv.ParseUint(transport.Src().String(), 10, 16)

	params, err = httpparser.GetBodyPayload(r)
	if err != nil {
		errs = append(errs, err.Error())
	}

	ev := &HTTPEvent{
		Verb:          r.Method,
		Proto:         r.Proto,
		RequestURI:    r.URL.RequestURI(),
		SourcePort:    uint16(srcPort),
		DestPort:      uint16(dstPort),
		DestHost:      network.Dst().String(),
		Body:          logdata.NewPayloadLogData(params, config.Cfg.MaxPOSTDataSize),
		IsTLS:         r.TLS != nil,
		Headers:       headers,
		InlineHeaders: inlineHeaders,
		Errors:        errs,
	}

	// Cannot use promoted (inherited) fields in struct literal
	ev.Session = sessions.SessionMap.GetUID(transport.String())
	ev.SourceIP = network.Src().String()
	ev.Tags = make(Tags)
	ev.Additional = make(map[string]string)

	if ev.IsTLS {
		ev.Kind = config.HTTPSKind
	} else {
		ev.Kind = config.HTTPKind
	}

	return ev, nil
}

// NewHTTPEventFromRequest creates an HTTPEvent from an http.Request if flow information is not available. It is used
// for HTTPS events, as they're generated from the dummy webserver and not reassembled by Melody
func NewHTTPEventFromRequest(r *http.Request) (*HTTPEvent, error) {
	headers := make(map[string]string)
	var inlineHeaders []string
	var errs []string
	var params []byte
	var srcIP string
	var dstHost string
	var rawDstPort string
	var rawSrcPort string
	var err error

	for header := range r.Header {
		headers[header] = r.Header.Get(header)
		inlineHeaders = append(inlineHeaders, header+": "+r.Header.Get(header))
	}

	dstHost, rawDstPort, err = net.SplitHostPort(r.Host)
	if err != nil {
		errs = append(errs, err.Error())
	}

	srcIP, rawSrcPort, err = net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		errs = append(errs, err.Error())
	}

	params, err = httpparser.GetBodyPayload(r)
	if err != nil {
		errs = append(errs, err.Error())
	}

	srcPort, _ := strconv.ParseUint(rawSrcPort, 10, 16)
	dstPort, _ := strconv.ParseUint(rawDstPort, 10, 16)

	ev := &HTTPEvent{
		Verb:          r.Method,
		Proto:         r.Proto,
		RequestURI:    r.URL.RequestURI(),
		SourcePort:    uint16(srcPort),
		DestPort:      uint16(dstPort),
		DestHost:      dstHost,
		Body:          logdata.NewPayloadLogData(params, config.Cfg.MaxPOSTDataSize),
		IsTLS:         r.TLS != nil,
		Headers:       headers,
		InlineHeaders: inlineHeaders,
		Errors:        errs,
	}

	// Cannot use promoted (inherited) fields in struct literal
	ev.Session = xid.New().String()
	ev.SourceIP = srcIP
	ev.Tags = make(Tags)
	ev.Additional = make(map[string]string)

	if ev.IsTLS {
		ev.Kind = config.HTTPSKind
	} else {
		ev.Kind = config.HTTPKind
	}

	return ev, nil
}
