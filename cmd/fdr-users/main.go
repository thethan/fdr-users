package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	firebase2 "github.com/thethan/fdr-users/pkg/firebase"
	"go.elastic.co/apm/module/apmgrpc"
	"go.elastic.co/apm/module/apmhttp"
	"google.golang.org/api/option"

	firebase "firebase.google.com/go"
	gokitLogrus "github.com/go-kit/kit/log/logrus"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/thethan/fdr-users/pkg/auth"
	"github.com/thethan/fdr-users/pkg/auth/transports"
	"go.elastic.co/apm"
	"go.elastic.co/apm/module/apmlogrus"
	"io/ioutil"

	"net"
	"net/http"
	"net/http/pprof"
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

///{
//  "type": "service_account",
//  "project_id": "fantasy-draft-room-248123",
//  "private_key_id": "ffa92c991c4bd5d59284296012c1ff6285dbaa1f",
//  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDZwpnWlGiDPDtQ\ngrGcAV3d/UImIfzMgI0TouG2CIi5qd9BRtjKt/umKy/9jpcTJLdUrCWdOFInOWui\ndcGxGxzk6Zs6J1V/HT2xbI6JL5B0/Yxt8EDbHwC5caWjsjj/e8VW+JLv5dSB95ql\nmtcEJzl5j98yt9CEXI37LCnMUVYo8e5QwKTpUHgLkTJJ0NEfTzIcAcIJ6NXQ1g0F\n/KOp+iFPusbmUK3aCDqQyGRSJLgwbaDWeYEfSoMyRoykyYn24XQOUR7X04UKVJtw\nmyoXZQ3Ne+IuBnkqaM8WymjwMusL5mB4Z+8y/e9i0Cf3USDi1kclcy7ghLBcIOJV\nfcikrW2pAgMBAAECggEAIfAoVb8PgtSSUfvsfmngHUbpVlQZuC2YzySllN9Dn9wP\nxXarNvzxpXY5poTgmsUwJWwm+Jfchex3D/zWUSnumOanoKqcspD2Gn7WwB6/ntwd\nVM0K7puoWz6RGDAgngDGQsW+8NCbDB5w5bp6JFWQqZd4q8jmIJrkLe82HHfYu8yf\nL07r6t6Z3NyfeeVB1qbMjpuerJc2+1V0NUCxRBZXrWbjKaKAy7AyUcWeQSXGqxLK\n80hJGXh+yavPltXGGwrn1kF/W99L5oFSnltUZx0500jPWN2OJAe69Cd7YLojamys\npwORA13OCe3VxmRFJJH0ZbRjBOGArcFtm/pZ/DsAGQKBgQD3wqnas9DNCOZf4nKN\nhSHgMvWJ4jmkECdbgCDeIJ608aX6MTPliYtYBWH1wwRlAaNEN6LSMSgcKV77riDL\nztMJWp7pokznkeObu9MhyPnOJmcuOhJVQEr64w/W27v7SrnUtH6lDm/m1rSn5/38\niKINeJVS6FBHqu5gSybE/934swKBgQDhAIbk1PDJc33ZaJpfk1f/hBZDRyrEz/XV\n5ceOrYI6S/idLHlOgBV44ISOxZBQWrcplsmKV+nAjtRi1InXP/ER/2etoW718zj6\ngD/RTjpNBhs0IvNQIx1NkCcOypbQXF2kyV6jC5Iooqr4z4/Ow5gazi5+86FL4zi9\nWxiOe1SWMwKBgQDaZ7Oro0+xLuNGKyyoLHAMX1+ryMzfH45STsSqiz7caxjRUIZb\nFcDMOxJ7vwoksCjofdL+T274RFACtSEcCJpoaIYllnkMucJXCl+4LJ5pZ9kVGwQG\nOsLeH0NbOCCiCOF/7AyoG+3xI9vlF9EByMBx95ZKm5gJVVkFcbofdx6JmQKBgQDN\nGYndNi53tAtYDv4JeWqRxHn2wfy+g0L4xAhwisFXGsF5pHy/jgoEscSj0HuIg+jK\nxGTa8uBlYs0/ebZcvDCn00VTBQD8ucWKszV5OfHzHEnX8LQSrK+dcHXqCcoIDOpf\nuB/ISFfnKsDnJW1VcP5KEQBZQQQbBPlHwq5T0yB7+QKBgQCXjvGEED4LysW3OaJu\nvM/0LY3SRgZG1OsmCWTIF4RfYQA3M+A9YF4NF2Rm/UVeTJoHwDA8owjVN5D1th/2\nD2ExpF90aCuapFLMQ5th9vLr7LqH1SFt8sq6GBdqxBnxbqFRQjCSLHhr/wnBrzu0\nWUx+x0Vm+rIwUzfXYRvPAfmG2A==\n-----END PRIVATE KEY-----\n",
//  "client_email": "firebase-adminsdk-dw8dr@fantasy-draft-room-248123.iam.gserviceaccount.com",
//  "client_id": "112531787805144036682",
//  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
//  "token_uri": "https://oauth2.googleapis.com/token",
//  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
//  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-dw8dr%40fantasy-draft-room-248123.iam.gserviceaccount.com"
//}
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
	authEndpoint := auth.NewEndpoints(logger)

	firestoreclient := initializeAppDefault(ctx, DefaultConfig, logger)
	repo := firebase2.NewFirebaseRepository(logger, firestoreclient)


	endpoints := users.NewEndpoints(logger, &repo)


	// Mechanical domain.
	errc := make(chan error)

	// Interrupt handler.
	go handlers.InterruptHandler(errc)

	// Debug listener.
	go func() {
		logger.Log("transport", "debug", "addr", DefaultConfig.DebugAddr)

		m := http.NewServeMux()
		m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
		m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
		m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
		m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
		m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

		errc <- http.ListenAndServe(DefaultConfig.DebugAddr, m)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", DefaultConfig.HTTPAddr)
		m := mux.NewRouter()

		m = users.MakeHTTPHandler(endpoints, m)
		m = transports.MakeHTTPHandler(authEndpoint, m)

		m.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {

			fmt.Println(route.GetName())
			fmt.Println(route.GetPathRegexp())
			fmt.Println(route.GetMethods())
			return nil
		})

		errc <- http.ListenAndServe(DefaultConfig.HTTPAddr, apmhttp.Wrap(m))
	}()

	// gRPC transport.
	go func() {
		logger.Log("transport", "gRPC", "addr", DefaultConfig.GRPCAddr)
		ln, err := net.Listen("tcp", DefaultConfig.GRPCAddr)
		if err != nil {
			errc <- err
			return
		}

		srv := users.MakeGRPCServer(endpoints)
		s := grpc.NewServer(grpc.UnaryInterceptor(apmgrpc.NewUnaryServerInterceptor()))
		pb.RegisterUsersServer(s, srv)

		errc <- s.Serve(ln)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func initializeAppDefault(ctx context.Context, config Config, logger log.Logger) *firestore.Client {

	sa := option.WithCredentialsFile(config.ServiceAccountFileLocation)
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		level.Error(logger).Log("message", "error in setting up firebase")
		os.Exit(1)
	}
	firestoreClient, err := app.Firestore(ctx)
	if err != nil {
		level.Error(logger).Log("message", "error in setting up firestore")

		os.Exit(1)
	}
	return firestoreClient
}
