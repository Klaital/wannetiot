package loggingresponsewriter

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type LoggingResponseWriter struct {
	w          *http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	logger     *logrus.Entry
}

func NewWithLogrus(w http.ResponseWriter, logger *logrus.Entry) *LoggingResponseWriter {
	var buf bytes.Buffer
	l := logger
	if l == nil {
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		l = logrus.NewEntry(logger)

	}
	return &LoggingResponseWriter{
		w:      &w,
		body:   &buf,
		logger: l,
	}
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := NewWithLogrus(w, nil)
		defer func() {
			lw.logger.Info(fmt.Sprintf("[%s] (%d ms) [Request: %s %s] [Response: %s]",
				time.Now().Format(time.RFC3339), time.Since(start).Milliseconds(),
				r.Method, r.RequestURI,
				lw.String()))
		}()
		next.ServeHTTP(lw, r)
	})
}

func (lw *LoggingResponseWriter) Write(buf []byte) (int, error) {
	lw.body.Write(buf)
	lw.logger.WithFields(logrus.Fields{
		"buf": string(buf),
		"len": len(buf),
	}).Debug("Adding bytes to body")
	return (*lw.w).Write(buf)
}

func (lw *LoggingResponseWriter) Header() http.Header {
	return (*lw.w).Header()
}

func (lw *LoggingResponseWriter) WriteHeader(statusCode int) {
	lw.logger.WithFields(logrus.Fields{
		"code":    statusCode,
		"oldCode": lw.statusCode,
	}).Debug("Setting status code")
	lw.statusCode = statusCode
	(*lw.w).WriteHeader(statusCode)
}

func (lw *LoggingResponseWriter) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Status=%d", lw.statusCode))
	buf.WriteString("Headers:")
	for k, v := range (*lw.w).Header() {
		buf.WriteString(fmt.Sprintf("%s=%s", k, v))
	}
	buf.WriteString("Body:\n")
	buf.WriteString(lw.body.String())
	return buf.String()
}
