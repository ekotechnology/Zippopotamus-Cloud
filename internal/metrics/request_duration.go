package metrics

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/zippopotamus/zippopotamus/internal"
	"net/http"
	"time"
)

type responseObserver struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (o *responseObserver) Write(p []byte) (n int, err error) {
	if !o.wroteHeader {
		o.WriteHeader(http.StatusOK)
	}
	n, err = o.ResponseWriter.Write(p)
	o.written += int64(n)
	return
}

func (o *responseObserver) WriteHeader(code int) {
	o.ResponseWriter.WriteHeader(code)
	if o.wroteHeader {
		return
	}
	o.wroteHeader = true
	o.status = code
}

func CollectRequestDuration(log *logrus.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			o := &responseObserver{ResponseWriter: w}

			next.ServeHTTP(o, r)

			ctx := chi.RouteContext(r.Context())

			if ctx == nil {
				log.Warn("Failed to get route context")
				return
			}

			path := ctx.RoutePattern()

			if path == "" || path == "/metrics" {
				return
			}

			s := r.Context().Value("requestStart")

			if s == nil {
				log.WithField("path", path).Warn("Unable to calculate request duration, requestStart is nil")
				return
			}

			RequestDuration.With(prometheus.Labels{"method": r.Method, "path": ctx.RoutePattern(), "apiversion": internal.GetVersion(r), "code": fmt.Sprintf("%d", o.status)}).Observe(time.Since(s.(time.Time)).Seconds() * 1000)

			PayloadBytes.With(prometheus.Labels{"method": r.Method, "path": ctx.RoutePattern(), "apiversion": internal.GetVersion(r), "code": fmt.Sprintf("%d", o.status)}).Observe(float64(o.written))
		}

		return http.HandlerFunc(fn)
	}
}
