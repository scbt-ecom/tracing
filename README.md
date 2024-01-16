# Tracing package for Совкомбанк Технологии

## Getting started
```bash
go get github.com/skbt-ecom/tracing
```

## Development

### Get logger instance from trace context
```
log := GetLoggerTracingFromContext(сtx, log)
```

### Get logger instance from trace HTTP request
```
ctx, w, log := GetLoggerTracingFromRequest(log, req, w)
```

### Get logger instance from trace AMQP headers
```
ctx, log := GetLoggerTracingFromAmqp(ctx, log, headers)
```
