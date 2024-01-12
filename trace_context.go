package tracing

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"net/http"
)

type key string

const traceKey key = "traceData"

type TracingStruct struct {
	TraceId      string
	SpanId       string
	ParentSpanId string
}

type Values struct {
	m map[string]string
}

func (v *Values) Get(key string) string {
	val, ok := v.m[key]
	if !ok {
		return ""
	}

	return val
}

func GetTracingHeadersFromContext(ctx context.Context) *TracingStruct {
	traceValues, ok := ctx.Value(traceKey).(Values)
	trace := TracingStruct{}
	if ok {
		trace.TraceId = traceValues.Get("TraceId")
		trace.SpanId = traceValues.Get("SpanId")
		trace.ParentSpanId = traceValues.Get("ParentSpanId")
	}

	return &trace
}

func GetLoggerTracing(ctx context.Context, log *logrus.Entry) *logrus.Entry {
	t := GetTracingHeadersFromContext(ctx)
	return log.WithFields(logrus.Fields{
		"X-B3-TraceId":      t.TraceId,
		"X-B3-SpanId":       t.SpanId,
		"X-B3-ParentSpanId": t.ParentSpanId,
	})
}

func contextGen(ctx context.Context, tracingData map[string]string) context.Context {
	traceId, ok := tracingData["X-B3-TraceId"]
	if traceId == "" || !ok {
		traceId = NewRandom64().TraceID().String()
	}

	parentSpanId, ok := tracingData["X-B3-SpanId"]
	if !ok {
		parentSpanId = ""
	}

	spanId := NewRandom64().SpanID(NewRandom64().TraceID()).String()

	v := Values{m: map[string]string{
		"TraceId":      traceId,
		"ParentSpanId": parentSpanId,
		"SpanId":       spanId,
	}}

	return context.WithValue(ctx, traceKey, v)
}

func getHeader(key string, m map[string][]string) (string, bool) {
	val, ok := m[key]
	if !ok {
		return "", false
	}

	return val[0], true
}

func getTracingFromHeaders(headers map[string][]string) map[string]string {
	tracing := make(map[string]string, 3)
	traceId, ok := getHeader("X-B3-Traceid", headers)
	if !ok {
		return tracing
	}
	tracing["X-B3-TraceId"] = traceId
	spanId, ok := getHeader("X-B3-Spanid", headers)
	if !ok {
		return tracing
	}
	tracing["X-B3-SpanId"] = spanId
	parentSpanId, ok := getHeader("X-B3-Parentspanid", headers)
	if !ok {
		return tracing
	}
	tracing["X-B3-ParentSpanId"] = parentSpanId

	return tracing
}

func GetLoggerTracingFromRequest(log *logrus.Entry, req *http.Request, w http.ResponseWriter) (context.Context, http.ResponseWriter, *logrus.Entry) {
	headers := getTracingFromHeaders(req.Header)
	ctx := contextGen(req.Context(), headers)
	t := GetTracingHeadersFromContext(ctx)

	w.Header().Set("X-B3-TraceId", t.TraceId)
	w.Header().Set("X-B3-SpanId", t.SpanId)
	w.Header().Set("X-B3-ParentSpanId", t.ParentSpanId)

	return ctx, w, log.WithFields(logrus.Fields{
		"X-B3-TraceId":      t.TraceId,
		"X-B3-SpanId":       t.SpanId,
		"X-B3-ParentSpanId": t.ParentSpanId,
	})
}

func GetLoggerTracingFromAmqp(ctx context.Context, log *logrus.Entry, headers amqp.Table) (context.Context, *logrus.Entry) {
	h := make(map[string]string, 3)

	t, ok := headers["X-B3-TraceId"]
	if ok {
		h["X-B3-TraceId"] = fmt.Sprintf("%v", t)
	}

	s, ok := headers["X-B3-SpanId"]
	if ok {
		h["X-B3-SpanId"] = fmt.Sprintf("%v", s)
	}

	p, ok := headers["X-B3-ParentSpanId"]
	if ok {
		h["X-B3-ParentSpanId"] = fmt.Sprintf("%v", p)
	}

	c := contextGen(ctx, h)
	tr := GetTracingHeadersFromContext(c)

	return c, log.WithFields(logrus.Fields{
		"X-B3-TraceId":      tr.TraceId,
		"X-B3-SpanId":       tr.SpanId,
		"X-B3-ParentSpanId": tr.ParentSpanId,
	})
}

func GetTracingAmqpTableFromContext(ctx context.Context) amqp.Table {

	t := GetTracingHeadersFromContext(ctx)

	return amqp.Table{
		"X-B3-TraceId":      t.TraceId,
		"X-B3-SpanId":       t.SpanId,
		"X-B3-ParentSpanId": t.ParentSpanId,
	}
}

func SetTracingForRequest(ctx context.Context, req *http.Request) {
	trace := GetTracingHeadersFromContext(ctx)
	req.Header.Set("X-B3-TraceId", trace.TraceId)
	req.Header.Set("X-B3-SpanId", trace.SpanId)
	req.Header.Set("X-B3-ParentSpanId", trace.ParentSpanId)
}
