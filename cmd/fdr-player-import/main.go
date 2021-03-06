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
	"github.com/oklog/run"
	"github.com/thethan/fdr-users/handlers"
	players2 "github.com/thethan/fdr-users/internal/importer/players"
	yahoo2 "github.com/thethan/fdr-users/internal/importer/repositories/yahoo"
	repositories2 "github.com/thethan/fdr-users/internal/oauth/repositories"
	yahoo4 "github.com/thethan/fdr-users/internal/oauth/yahoo"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/coordinator"
	"github.com/thethan/fdr-users/pkg/coordinator/transports"
	"github.com/thethan/fdr-users/pkg/draft/repositories"
	firebase2 "github.com/thethan/fdr-users/pkg/firebase"
	kubemq3 "github.com/thethan/fdr-users/pkg/kubemq"
	"github.com/thethan/fdr-users/pkg/league"
	"github.com/thethan/fdr-users/pkg/mongo"
	"github.com/thethan/fdr-users/pkg/players"
	"github.com/thethan/fdr-users/pkg/players/entities"
	"github.com/thethan/fdr-users/pkg/players/importer/queue"
	repositories3 "github.com/thethan/fdr-users/pkg/players/repositories"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"

	firebase "firebase.google.com/go"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
	"go.opentelemetry.io/otel/exporters/metric/prometheus"
	"os"
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

	oauthConfig = yahoo4.Oauth2Config()

	apm.DefaultTracer.SetLogger(logrusLogger)
	logrusLogger.AddHook(&apmlogrus.Hook{})
}

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint(os.Getenv("JAEGER_ENDPOINT")),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "fdr-players-import",
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
				label.Float64("float", 312.23),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		panic("tracerr not made")
	}

	return func() {
		flush()
	}
}

func initMeter() *prometheus.Exporter {
	exporter, err := prometheus.InstallNewPipeline(prometheus.Config{})
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

	tracer := global.Tracer("fantasydraftroom.com/player-import")
	metrics := global.Meter("fantasydraftroom.com/player-import")

	logger := log.NewJSONLogger(os.Stdout)
	logger = log.WithPrefix(logger, "caller_a", log.DefaultCaller, "caller_b", log.Caller(2), "caller_c", log.Caller(1))

	firestoreClient, firebaseauthclient := initializeAppDefault(ctx, DefaultConfig, logger)

	authRepo := firebase2.NewFirestoreAuthRepo(logger, firebaseauthclient)

	authSvc := auth.NewAuthService(logger, &authRepo)
	authMiddleware := authSvc.NewAuthMiddleware(tracer, metrics)
	firebaseRepo := firebase2.NewFirebaseRepository(logger, firestoreClient, firebaseauthclient)
	getUserInfoMiddleware := authSvc.UserInformationToContext(&firebaseRepo)

	mongoClient, err := mongo.NewMongoDBClient(ctx, os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"), os.Getenv("MONGO_PORT"))
	if err != nil {
		logger.Log("message", "error in initializing mongo client", "error", err)
		os.Exit(1)
	}
	mongoRepo := repositories.NewMongoRepository(logger, mongoClient, "fdr", "draft", "fdr_user", "roster")

	oauthRepo := repositories2.NewMongoOauthRepository(logger, mongoClient, tracer)

	var kubemqConfig KubemqConfig
	err = env.Parse(&kubemqConfig)
	if err != err {
		level.Error(logger).Log("message", "could not parse kubemq config", "error", err)
		os.Exit(1)
	}

	kubemqClient, err := kubemq3.NewKubeMQClient(ctx, kubemqConfig.Address, kubemqConfig.Port, kubemqConfig.ClientID)
	if err != nil {
		level.Error(logger).Log("message", "could initiate kubemq client", "error", err)
		os.Exit(1)
	}

	ogGrouter := mux.NewRouter()
	ogGrouter.Use(otelmux.Middleware(os.Getenv("SERVICE_NAME")))

	ogGrouter.Methods(http.MethodGet).PathPrefix("/hambone").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte("hambone"))
		writer.WriteHeader(http.StatusAccepted)

	})

	// fdr-players-import
	yahooRepository := yahoo2.NewYahooRepository(logger, oauthConfig, tracer, &metrics)
	broadcastMq := queue.NewQueueImportStats(logger, kubemqClient)
	statsRepo := repositories3.NewMongoStatsRepo(logger, mongoClient)
	// new importer
	importerService := players2.NewImporterClient(logger, &mongoRepo, &oauthRepo, &broadcastMq, &yahooRepository, statsRepo, tracer, metrics)

	// Mechanical domain.
	errc := make(chan error)

	apm.DefaultTracer.SetLogger(logrusLogger)
	logrusLogger.AddHook(&apmlogrus.Hook{})

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

	importPlayerChannel := make(chan entities.ImportPlayer, 25)
	importPlayerStat := make(chan entities.ImportPlayerStat, 25)
	defer close(importPlayerChannel)
	defer close(importPlayerStat)
	yahooService := yahoo.NewService(logger, &firebaseRepo)
	leagueImportService := league.NewImportService(logger, yahooService, &mongoRepo, tracer)
	coordinatorEndpoints := coordinator.NewEndpoints(logrusLogger, leagueImportService, authMiddleware)
	ogGrouter = transports.NewHTTPServer(logrusLogger, ogGrouter, coordinatorEndpoints, authSvc.ServerBefore)
	playersService := players.NewService(logger, mongoRepo)
	players.NewEndpoint(logger, &playersService, authMiddleware, getUserInfoMiddleware)

	// gRPC transport.
	var g run.Group
	{
		g.Add(func() error {
			ln, _ := net.Listen("tcp", DefaultConfig.HTTPAddr)
			http.HandleFunc("/", exporter.ServeHTTP)
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
		// sender
		g.Add(func() error {
			return broadcastMq.StartPlayerWorker(ctx, importPlayerChannel)
		}, func(err error) {
			errc <- err
		})
	}
	{
		// worker
		g.Add(func() error {
			return broadcastMq.StartWorker(ctx, importPlayerStat)
		}, func(err error) {
			errc <- err
		})
	}
	{
		// workerProcessor
		g.Add(func() error {
			return importerService.Start(ctx, importPlayerStat)
		}, func(err error) {
			errc <- err
		})
	}
	{
		// workerProcessor
		g.Add(func() error {
			return importerService.StartPlayersWorker(ctx, importPlayerChannel)
		}, func(err error) {
			errc <- err
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
