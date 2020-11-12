package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebaseAuth "firebase.google.com/go/auth"
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	muxhandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/oklog/run"
	repositories2 "github.com/thethan/fdr-users/internal/oauth/repositories"
	"github.com/thethan/fdr-users/internal/oauth/yahoo"
	handlers2 "github.com/thethan/fdr-users/internal/oauth/yahoo/handlers"
	yahoo3 "github.com/thethan/fdr-users/internal/yahoo"
	otelmux "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/metric/controller/pull"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/thethan/fdr-users/pkg/kubemq"
	"github.com/thethan/fdr-users/pkg/mongo"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/oauth2"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	firebase2 "github.com/thethan/fdr-users/pkg/firebase"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go"
	gokitLogrus "github.com/go-kit/kit/log/logrus"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/auth"

	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"google.golang.org/grpc"
	"net"
	"os"

	"github.com/thethan/fdr-users/handlers"
)

var DefaultConfig Config

// Config contains the required fields for running a server
type Config struct {
	HTTPAddr                   string
	DebugAddr                  string
	GRPCAddr                   string
	ServiceAccountFileLocation string
}

type KubemqConfig struct {
	Address  string `env:"KUBEMQ_SERVICE" envDefault:"kubemq-cluster-grpc.kubemq"`
	ClientID string `env:"POD_NAME"`
	Port     int    `env:"KUBEMQ_PORT" envDefault:"50000"`
}

var oauthConfig *oauth2.Config

var logrusLogger = &logrus.Logger{
	Out:       os.Stdout,
	Hooks:     make(logrus.LevelHooks),
	Level:     logrus.DebugLevel,
	Formatter: &logrus.JSONFormatter{},
}

func init() {
	flag.StringVar(&DefaultConfig.DebugAddr, "debug.addr", ":8080", "Debug and metrics listen address")
	flag.StringVar(&DefaultConfig.HTTPAddr, "http.addr", ":8081", "HTTP listen address")
	flag.StringVar(&DefaultConfig.GRPCAddr, "grpc.addr", ":8082", "gRPC (HTTP) listen address")
	flag.StringVar(&DefaultConfig.ServiceAccountFileLocation, "service account file location", "/Users/ethan/Code/fdr-users/serviceAccountKey.json", "used for your firebase location")
	// Use environment variables, if set. Flags have priority over Env vars.
	if addr := os.Getenv("DEBUG_ADDR"); addr != "" {
		DefaultConfig.DebugAddr = addr
	}
	if port := os.Getenv("PORT"); port != "" {

		DefaultConfig.HTTPAddr = fmt.Sprintf(":%s", port)
	}
	if addr := os.Getenv("HTTP_ADDR"); addr != "" {
		DefaultConfig.HTTPAddr = addr
	}
	if addr := os.Getenv("GRPC_ADDR"); addr != "" {
		DefaultConfig.GRPCAddr = addr
	}

	if addr := os.Getenv("SERVICE_ACCOUNT_FILE_LOCATION"); addr != "" {
		DefaultConfig.ServiceAccountFileLocation = addr
	} else {
		fmt.Println(fmt.Sprintf("could not get service account location %s"))
		os.Exit(1)
	}

	oauthConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("YAHOO_CLIENT_REDIRECT"),
		ClientID:     os.Getenv("YAHOO_CLIENT_ID"),
		ClientSecret: os.Getenv("YAHOO_CLIENT_SECRET"),
		Scopes:       []string{"fspt-w"},
		Endpoint:     yahoo3.Endpoint,
	}

}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(os.Getenv("JAEGER_ENDPOINT")),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "fdr-users",
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
				label.Float64("float", 312.23),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		panic("tracer not made")
	}

	return func() {
		flush()
	}
}

func initMeter() *prometheus.Exporter {
	exporter, err := prometheus.NewExportPipeline(
		prometheus.Config{
			DefaultHistogramBoundaries: []float64{-0.5, 1},
		},
		pull.WithCachePeriod(0),
		pull.WithResource(resource.New(label.String("R", "V"))),
	)
	if err != nil {
		panic("Could not init Meter")
	}
	return exporter
}

// Run starts a new http server, gRPC server, and a debug server with the
// passed config and logger
func main() {
	ctx := context.Background()
	exporter := initMeter()
	initTracer()

	tracer := global.Tracer("fantasydraftroom.com/users")
	meter := global.Meter("fantasydraftroom.com/users")

	logger := gokitLogrus.NewLogrusLogger(logrusLogger)
	logger = log.WithPrefix(logger, "caller_a", log.DefaultCaller, "caller_b", log.Caller(2), "caller_c", log.Caller(1))

	_, firebaseauthclient := initializeAppDefault(ctx, DefaultConfig, logger)

	//repo := firebase2.NewFirebaseRepository(logger, firestoreclient, firebaseauthclient)
	authRepo := firebase2.NewFirestoreAuthRepo(logger, firebaseauthclient)
	authSvc := auth.NewAuthService(logger, &authRepo)
	authMiddleware := authSvc.NewAuthMiddleware(tracer, meter)

	mongoClient, err := mongo.NewMongoDBClient(ctx, os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		os.Exit(1)
	}

	oauthRepo := repositories2.NewMongoOauthRepository(logger, mongoClient, tracer)
	oauthYahooService := yahoo.NewOauthYahooService(logger, tracer, oauthConfig, &oauthRepo)

	oauthYahooEndpoints := handlers2.NewYahooHandlersEndpoints(logger, oauthConfig, tracer, &oauthYahooService, authMiddleware)

	var kubemqConfig KubemqConfig
	err = env.Parse(&kubemqConfig)
	if err != err {
		level.Error(logger).Log("message", "could not parse kubemq config", "error", err)
		os.Exit(1)
	}

	kubemqClient, err := kubemq.NewKubeMQClient(ctx, kubemqConfig.Address, kubemqConfig.Port, kubemqConfig.ClientID)
	if err != nil {
		level.Error(logger).Log("message", "could initiate kubemq client", "error", err)
		os.Exit(1)
	}

	defer kubemqClient.Close()
	ogGrouter := mux.NewRouter()
	ogGrouter.Use(otelmux.Middleware("fdr-users"))
	// prometheus metrics
	ogGrouter.Path("/metrics").HandlerFunc(exporter.ServeHTTP)

	hamboneCount, err := meter.NewInt64Counter("hambone_counter")
	ogGrouter.Methods(http.MethodGet).PathPrefix("/hambone").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hamboneCount.Add(request.Context(), 1)
		_, _ = writer.Write([]byte("hambone gonna get you"))
		writer.WriteHeader(http.StatusAccepted)
	})

	ogGrouter = handlers2.MakeHTTPHandler(logger, oauthYahooEndpoints, ogGrouter, authSvc.ServerBefore, tracer)

	// Mechanical domain.
	errc := make(chan error)

	// Interrupt handler.
	go handlers.InterruptHandler(errc)

	_ = ogGrouter.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("ROUTE:", pathTemplate)
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			fmt.Println("Path regexp:", pathRegexp)
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			fmt.Println("Queries templates:", strings.Join(queriesTemplates, ","))
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			fmt.Println("Queries regexps:", strings.Join(queriesRegexps, ","))
		}
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Println("Methods:", strings.Join(methods, ","))
		}
		fmt.Println()
		return nil
	})

	// gRPC transport.
	var g run.Group
	{
		g.Add(func() error {
			ln, _ := net.Listen("tcp", DefaultConfig.HTTPAddr)
			return http.Serve(ln,
				muxhandlers.CORS(muxhandlers.AllowedHeaders([]string{
					"X-Requested-With",
					"Content-Type",
					"Authorization"}),
					muxhandlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), muxhandlers.AllowedOrigins([]string{"*"}))(ogGrouter))
		}, func(err error) {
			return
		})
	}
	{
		g.Add(func() error {

			logger.Log("transport", "gRPC", "addr", DefaultConfig.GRPCAddr)
			ln, err := net.Listen("tcp", DefaultConfig.GRPCAddr)
			if err != nil {
				errc <- err
			}

			authInterceptor := grpc_auth.UnaryServerInterceptor(authSvc.ServerAuthentication(ctx, logger))
			apmInterceptor := otelgrpc.UnaryServerInterceptor(tracer)
			middlewareInterceptor := grpc_middleware.ChainUnaryServer(apmInterceptor, authInterceptor)

			//grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authSvc.ServerAuthentication(ctx, logger))
			s := grpc.NewServer(grpc.UnaryInterceptor(middlewareInterceptor), grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor(tracer)))

			return s.Serve(ln)
		}, func(err error) {
		})
	}
	{
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case err := <-errc:
				return err
			}
		}, func(err error) {
			level.Error(logger).Log("interrupt triggered", err.Error())
			close(errc)
		})
	}

	// Run!
	_ = level.Error(logger).Log("exit", g.Run())
}

func initializeAppDefault(ctx context.Context, config Config, logger log.Logger) (*firestore.Client, *firebaseAuth.Client) {
	sa := option.WithCredentialsFile(config.ServiceAccountFileLocation)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		level.Error(logger).Log("message", "error in setting up firebase")
		os.Exit(1)
	}
	firestoreClient := initializeFirestore(ctx, logger, app)

	firebaseAuthClient := initializeFirebaseAuth(ctx, logger, app)

	return firestoreClient, firebaseAuthClient
}

func initializeFirestore(ctx context.Context, logger log.Logger, app *firebase.App) *firestore.Client {
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		level.Error(logger).Log("message", "error in setting up firestore")
		os.Exit(1)
	}
	return firestoreClient
}

func initializeFirebaseAuth(ctx context.Context, logger log.Logger, app *firebase.App) *firebaseAuth.Client {

	firebaseAuthClient, err := app.Auth(ctx)
	if err != nil {
		level.Error(logger).Log("message", "error in setting up firestore")

		os.Exit(1)
	}

	return firebaseAuthClient
}
