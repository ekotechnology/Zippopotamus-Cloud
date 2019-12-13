package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-redis/redis"
	"github.com/zippopotamus/zippopotamus/internal/metrics"
	"net/http"
	"os"
	"time"

	"github.com/evalphobia/logrus_sentry"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/zippopotamus/zippopotamus/internal"
)

func setupSentry(l *logrus.Logger, sentryDsn *string) {
	hook, err := logrus_sentry.NewSentryHook(*sentryDsn, []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	})

	if err == nil {
		hook.Timeout = 2 * time.Second
		hook.StacktraceConfiguration.Enable = true

		if env := os.Getenv("APP_ENV"); env != "" {
			hook.SetEnvironment(env)
		}

		hook.SetRelease(fmt.Sprintf("%s@%s", internal.Version, internal.GitCommit))

		l.Hooks.Add(hook)
	}
}

func main() {
	log := logrus.StandardLogger()

	log.Printf("Zippopotam.us API Server\nVersion: %s\tSHA: %s\tBuilt: %s", internal.Version, internal.GitCommit, internal.BuildDate)

	redisAddr := flag.String("redis-addr", "127.0.0.1:6379", "the location at which Redis server can be found")
	sentryDsn := flag.String("sentry-dsn", "", "Sentry/Raven DSN to send log/errors to")

	flag.Parse()

	if sentryDsn != nil && *sentryDsn != "" {
		setupSentry(log, sentryDsn)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     *redisAddr,
		Password: "",
		DB:       0,
	})

	codeNames := &internal.AdminCodeNames{
		Admin1:    &internal.Admin1,
		Admin2:    &internal.Admin2,
		Countries: &internal.Countries,
	}

	srv := internal.NewHttpServerHandlers(redisClient, log, codeNames)

	r := chi.NewRouter()
	r.Use(func(handler http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			r = r.WithContext(context.WithValue(ctx, "requestStart", time.Now()))
			handler.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	})

	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(internal.SetVersion)
	r.Use(metrics.CollectRequestDuration(log))

	r.Handle("/metrics", promhttp.Handler())

	r.Group(func(r chi.Router) {
		r.Get("/", srv.HandleIndex)
		r.Get("/static/sample_us.html", nil)
		r.Get("/static/sample_de.html", nil)
		r.Get("/static/sample_fr.html", nil)
		r.Get("/static/sample_es.html", nil)
	})

	// actual service routes
	r.Group(func(r chi.Router) {
		r.Route("/{countryCode:[a-z]{2}}", func(r chi.Router) {
			r.HandleFunc("/", srv.HandleCheckCountryAvailable)
			r.Get("/{postalCode}", srv.HandleGetPlacesByCountryAndPostCode)
			r.Get("/{area:[a-z]{2}}/{place}", srv.HandleGetPlacesByCountryAreaAndPlaceName)
		})
		r.Get("/nearby/{countryCode:[a-z]{2}}/{postalCode}", srv.HandleGetNearbyPlaces)
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
