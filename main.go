package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/t-kimura1988/kihenya-daic-app/external/clients"
	"github.com/t-kimura1988/kihenya-daic-app/internal/account"
	"github.com/t-kimura1988/kihenya-daic-app/internal/diary"
	"github.com/t-kimura1988/kihenya-daic-app/internal/goal"
	"github.com/t-kimura1988/kihenya-daic-app/internal/ideal"
	"github.com/t-kimura1988/kihenya-daic-app/internal/interceptor"
	"github.com/t-kimura1988/kihenya-daic-app/internal/task"
	accountconnect "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/account/accountconnect"
	diaryconnect "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/diary/diaryconnect"
	goalconnect "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/goal/goalconnect"
	idealconnect "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/ideal/idealconnect"
	taskconnect "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/task/taskconnect"
)

var port = flag.Int("port", 50054, "The server port")

func init() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(fmt.Sprintf("failed to load timezone Asia/Tokyo: %v", err))
	}
	time.Local = loc
}

func main() {
	if err := godotenv.Overload(); err != nil {
		logrus.WithError(err).Warn("Error loading .env file")
	}

	flag.Parse()
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	serverPort := *port
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			serverPort = p
		}
	}

	if err := RunMigrations(); err != nil {
		logrus.WithError(err).Fatal("failed to run database migrations")
	}

	if err := InitDB(); err != nil {
		logrus.WithError(err).Fatal("failed to initialize database")
	}
	defer CloseDB()

	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	jwtIssuer := os.Getenv("SUPABASE_JWT_ISSUER")
	if jwtIssuer == "" {
		logrus.Fatal("SUPABASE_JWT_ISSUER environment variable is not set")
	}

	tokenValidator, err := NewTokenValidator(jwtSecret, jwtIssuer)
	if err != nil {
		logrus.WithError(err).Fatal("failed to create token validator")
	}

	authInterceptor := interceptor.NewAuthInterceptor(tokenValidator, GetQueries())
	mux := http.NewServeMux()

	accountHandler := account.NewHandler(GetQueries())
	accountPath, accountHTTPHandler := accountconnect.NewAccountServiceHandler(
		accountHandler,
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(accountPath, accountHTTPHandler)

	idealService := ideal.NewService(GetQueries())
	idealPath, idealHTTPHandler := idealconnect.NewIdealServiceHandler(
		idealService,
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(idealPath, idealHTTPHandler)

	goalService := goal.NewService(GetQueries())
	goalPath, goalHTTPHandler := goalconnect.NewGoalServiceHandler(
		goalService,
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(goalPath, goalHTTPHandler)

	taskService := task.NewService(GetQueries())
	taskPath, taskHTTPHandler := taskconnect.NewTaskServiceHandler(
		taskService,
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(taskPath, taskHTTPHandler)

	kblogClient := clients.NewKblogClient()
	diaryService := diary.NewService(GetQueries(), kblogClient)
	diaryPath, diaryHTTPHandler := diaryconnect.NewDiaryServiceHandler(
		diaryService,
		connect.WithInterceptors(authInterceptor),
	)
	mux.Handle(diaryPath, diaryHTTPHandler)

	h2s := &http2.Server{}
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", serverPort),
		Handler: h2c.NewHandler(mux, h2s),
	}

	logrus.WithField("addr", server.Addr).Info("Daic server listening (Connect over HTTP/2)")
	if err := server.ListenAndServe(); err != nil {
		logrus.WithError(err).Fatal("failed to serve")
	}
}
