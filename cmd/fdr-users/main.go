package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebaseAuth "firebase.google.com/go/auth"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	muxhandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/oklog/run"
	"github.com/thethan/fdr-users/pkg/coordinator/transports"
	"github.com/thethan/fdr-users/pkg/goff"
	"github.com/thethan/fdr-users/pkg/yahoo"
	"go.elastic.co/apm/module/apmgorilla"
	"net/http"
	"os/signal"
	"syscall"

	firebase2 "github.com/thethan/fdr-users/pkg/firebase"
	"go.elastic.co/apm/module/apmgrpc"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go"
	gokitLogrus "github.com/go-kit/kit/log/logrus"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/coordinator"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
	"io/ioutil"

	"net"
	"os"

	// 3d Party
	"google.golang.org/grpc"

	"github.com/thethan/fdr-users/handlers"
	"github.com/thethan/fdr-users/pkg/users"
	// This Service
	pb "github.com/thethan/fdr_proto"
)

var DefaultConfig Config

// Config contains the required fields for running a server
type Config struct {
	HTTPAddr                   string
	DebugAddr                  string
	GRPCAddr                   string
	ServiceAccountFileLocation string
}

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
		fmt.Println("could not get service account location")
		os.Exit(1)
	}

	apm.DefaultTracer.SetLogger(logrusLogger)
	logrusLogger.AddHook(&apmlogrus.Hook{})
}


type serviceAccount struct {
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertUrl string `json:"auth_provider_x509_cert_url"`
	ClientX509CertUrl       string `json:"client_x509_cert_url"`
}

func readJsonFile(cfg Config) serviceAccount {
	// Open our jsonFile
	jsonFile, err := os.Open(cfg.ServiceAccountFileLocation)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// we initialize our Users array
	var serviceAccount serviceAccount
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &serviceAccount)
	if err != nil {
		fmt.Errorf("error in unmarshalling service account")
		os.Exit(1)
	}

	return serviceAccount
}

// Run starts a new http server, gRPC server, and a debug server with the
// passed config and logger
func main() {
	ctx := context.Background()

	logger := gokitLogrus.NewLogrusLogger(logrusLogger)
	logger = log.WithPrefix(logger, "caller_a", log.DefaultCaller, "caller_b", log.Caller(2), "caller_c", log.Caller(1))

	firestoreclient, firebaseauthclient := initializeAppDefault(ctx, DefaultConfig, logger)

	repo := firebase2.NewFirebaseRepository(logger, firestoreclient)
	authRepo := firebase2.NewFirestoreAuthRepo(logger, firebaseauthclient)

	authSvc := auth.NewAuthService(logger, &authRepo)

	yahooProvider := yahoo.NewService(logger, &repo,)
	goffYahooClient := goff.NewClient(yahooProvider)
	coordinatorService := coordinator.NewService(logger, logrusLogger, goffYahooClient, &repo)
	//mongoClient, err := mongo.NewMongoDBClient(os.Getenv("MONGO_USERNAME"), os.Getenv("MONGO_PASSWORD"), os.Getenv("MONGO_HOST"))
	//if err != nil {
	//	logger.Log("message", "error in initializing mongo client", "error", err)
	//	os.Exit(1)
	//}
	coordinatorEndpoints := coordinator.NewEndpoints(logrusLogger, coordinatorService, authSvc.NewAuthMiddleware(&repo))
	_ = transports.NewServer(logger, logrusLogger, coordinatorEndpoints)
	coordinatorHTTPServer := transports.NewHTTPServer(logrusLogger, coordinatorEndpoints, authSvc.ServerBefore)
	endpoints := users.NewEndpoints(logger, &repo, &repo, authSvc.NewAuthMiddleware(&repo), authSvc.ServerBefore)

	router := mux.NewRouter()
	apmgorilla.Instrument(router)
	router.Methods(http.MethodGet).PathPrefix("/import").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAlreadyReported)
	})

	router.Methods(http.MethodPost).PathPrefix("/import").Handler(coordinatorHTTPServer)
	users.MakeHTTPHandler(endpoints, router, authSvc.ServerBefore)

	// Mechanical domain.
	errc := make(chan error)

	apm.DefaultTracer.SetLogger(logrusLogger)
	logrusLogger.AddHook(&apmlogrus.Hook{})

	// Interrupt handler.
	go handlers.InterruptHandler(errc)

	// gRPC transport.
	var g run.Group
	{
		g.Add(func() error {
			ln, _ := net.Listen("tcp", DefaultConfig.HTTPAddr)

			return http.Serve(ln, muxhandlers.CORS(muxhandlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}), muxhandlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), muxhandlers.AllowedOrigins([]string{"*"}))(router))
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

			srv := users.MakeGRPCServer(endpoints)

			authInterceptor := grpc_auth.UnaryServerInterceptor(authSvc.ServerAuthentication(ctx, logger))
			apmInterceptor := apmgrpc.NewUnaryServerInterceptor()
			middlewareInterceptor := grpc_middleware.ChainUnaryServer(apmInterceptor, authInterceptor)

			//grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(authSvc.ServerAuthentication(ctx, logger))
			s := grpc.NewServer(grpc.UnaryInterceptor(middlewareInterceptor))
			pb.RegisterUsersServer(s, srv)

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
