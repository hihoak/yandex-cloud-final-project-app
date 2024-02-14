package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
	"go.uber.org/zap"

	"artemmihaylov.gitlab.yandexcloud.net/final-project/momo-store/cmd/api/app"
	"artemmihaylov.gitlab.yandexcloud.net/final-project/momo-store/cmd/api/dependencies"
	"artemmihaylov.gitlab.yandexcloud.net/final-project/momo-store/internal/logger"
)

func newRouter(app *app.Instance) (http.Handler, error) {
	r := chi.NewRouter()

	corses := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	r.Use(
		middleware.StripSlashes,
		logMiddleware,
		corses.Handler,
	)

	r.Group(func(r chi.Router) {
		r.Use(
			app.TimingsMiddleware,
			app.RequestsMiddleware,
		)

		r.Get("/products", app.ListDumplingsController)
		r.Get("/categories", app.ListCategoriesController)
		r.Post("/orders", app.CreateOrderController)

		r.Get("/auth/whoami", app.WhoAmIController)
	})

	r.Get("/health", app.HealthcheckController)
	r.Method(http.MethodGet, "/metrics", app.MetricsHandler())

	return r, nil
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("got request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
		)
		next.ServeHTTP(w, r)
	})
}

func main() {
	logger.Setup()

	if err := run(); err != nil {
		logger.Log.Fatal("unexpected error", zap.Error(err))
		os.Exit(1)
	}
}

func run() error {
	lis, err := net.Listen("tcp", ":8081")
	if err != nil {
		return err
	}

	store, err := dependencies.NewFakeDumplingsStore()
	if err != nil {
		return fmt.Errorf("cannot bootstrap dumplings store: %w", err)
	}

	logger.Log.Debug("creating app instance")
	instance, err := app.NewInstance(store)
	if err != nil {
		return fmt.Errorf("cannot create app instance: %w", err)
	}

	router, err := newRouter(instance)
	if err != nil {
		return fmt.Errorf("cannot create router instance: %w", err)
	}

	srv := &http.Server{
		Handler: router,
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Log.Info("starting HTTP server", zap.String("address", ":8081"))
		if err := srv.Serve(lis); err != nil {
			errChan <- fmt.Errorf("error serving HTTP: %w", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Log.Info("shutting down gracefully", zap.String("signal", sig.String()))

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return srv.Shutdown(ctx)
	case err := <-errChan:
		return err
	}
}
