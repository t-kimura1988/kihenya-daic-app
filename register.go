package daicapp

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
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

func init() {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(fmt.Sprintf("failed to load timezone Asia/Tokyo: %v", err))
	}
	time.Local = loc
}

// Register migrates, connects to DB, and mounts Daic Connect handlers on mux.
// Caller must invoke the returned cleanup when shutting down.
func Register(mux *http.ServeMux) (cleanup func() error, err error) {
	if err := RunMigrations(); err != nil {
		return nil, fmt.Errorf("daic migrations: %w", err)
	}

	if err := InitDB(); err != nil {
		return nil, fmt.Errorf("daic init db: %w", err)
	}

	jwtSecret := os.Getenv("SUPABASE_JWT_SECRET")
	jwtIssuer := os.Getenv("SUPABASE_JWT_ISSUER")
	if jwtIssuer == "" {
		_ = CloseDB()
		return nil, fmt.Errorf("SUPABASE_JWT_ISSUER environment variable is not set")
	}

	tokenValidator, err := NewTokenValidator(jwtSecret, jwtIssuer)
	if err != nil {
		_ = CloseDB()
		return nil, fmt.Errorf("daic token validator: %w", err)
	}

	authInterceptor := interceptor.NewAuthInterceptor(tokenValidator, GetQueries())

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

	logrus.Info("daic handlers registered")
	return CloseDB, nil
}

// ListenAndServe starts an h2c HTTP server with all Daic handlers mounted.
func ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	cleanup, err := Register(mux)
	if err != nil {
		return err
	}
	defer func() { _ = cleanup() }()

	h2s := &http2.Server{}
	server := &http.Server{
		Addr:         addr,
		Handler:      h2c.NewHandler(mux, h2s),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logrus.WithField("addr", server.Addr).Info("Daic server listening (Connect over HTTP/2)")
	return server.ListenAndServe()
}

// InitLogging configures the shared logrus formatter used by this service.
func InitLogging() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
}
