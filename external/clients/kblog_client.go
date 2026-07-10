package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	kblogaccountv1 "github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/account"
	"github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/account/accountconnect"
	kblogbookletv1 "github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/booklet"
	"github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/booklet/bookletconnect"
	kblogdiaryv1 "github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/diary"
	"github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/diary/diaryconnect"
)

// KblogError はkblogサービスから返されるカスタムエラー
type KblogError struct {
	Code    string
	Message string
}

func (e *KblogError) Error() string {
	return fmt.Sprintf("kblog error [%s]: %s", e.Code, e.Message)
}

// 定義済みエラーコード
var (
	ErrKblogAccountNotFound = &KblogError{Code: "ERR_KBLOG_ACCOUNT_NOT_FOUND", Message: "kblog account not found"}
	ErrKblogUnknown         = &KblogError{Code: "ERR_KBLOG_UNKNOWN", Message: "unknown kblog error"}
)

func IsKblogAccountNotFound(err error) bool {
	var ke *KblogError
	if errors.As(err, &ke) {
		return ke.Code == "ERR_KBLOG_ACCOUNT_NOT_FOUND"
	}
	return false
}

// KblogClient はkihenya-blog gRPCサービスのクライアント
type KblogClient struct {
	accountClient accountconnect.AccountServiceClient
	bookletClient bookletconnect.BookletServiceClient
	diaryClient   diaryconnect.DiaryServiceClient
}

// NewKblogClient はKblogClientを生成する
func NewKblogClient() *KblogClient {
	kblogURL := os.Getenv("KBLOG_SERVICE_URL")
	if kblogURL == "" {
		kblogURL = "http://localhost:50053"
	}

	httpClient := &http.Client{}
	logrus.Printf("KblogClient initialized with URL: %s", kblogURL)

	return &KblogClient{
		accountClient: accountconnect.NewAccountServiceClient(httpClient, kblogURL),
		bookletClient: bookletconnect.NewBookletServiceClient(httpClient, kblogURL),
		diaryClient:   diaryconnect.NewDiaryServiceClient(httpClient, kblogURL),
	}
}

// GetAccount はkblogアカウント情報を取得する。
// アカウント未存在の場合は ErrKblogAccountNotFound を返す。
func (kc *KblogClient) GetAccount(ctx context.Context, authToken string) (*kblogaccountv1.GetAccountResponse, error) {
	req := connect.NewRequest(&kblogaccountv1.GetAccountRequest{})
	req.Header().Set("Authorization", authToken)

	resp, err := kc.accountClient.GetAccount(ctx, req)
	if err != nil {
		logrus.WithError(err).Warn("KblogClient.GetAccount error")
		return nil, convertKblogError(err)
	}
	return resp.Msg, nil
}

// CreateBooklet はkblog上に新しい冊子を作成する。
func (kc *KblogClient) CreateBooklet(ctx context.Context, authToken, title, description string) (*kblogbookletv1.CreateBookletResponse, error) {
	req := connect.NewRequest(&kblogbookletv1.CreateBookletRequest{
		BookletTitle:       title,
		BookletDescription: description,
	})
	req.Header().Set("Authorization", authToken)

	resp, err := kc.bookletClient.CreateBooklet(ctx, req)
	if err != nil {
		logrus.WithError(err).Warn("KblogClient.CreateBooklet error")
		return nil, convertKblogError(err)
	}
	return resp.Msg, nil
}

// GetBookletDetail は冊子の詳細（日記一覧含む）を取得する。
func (kc *KblogClient) GetBookletDetail(ctx context.Context, authToken string, bookletID int64) (*kblogbookletv1.GetBookletDetailResponse, error) {
	req := connect.NewRequest(&kblogbookletv1.GetBookletDetailRequest{
		BookletId: bookletID,
	})
	req.Header().Set("Authorization", authToken)

	resp, err := kc.bookletClient.GetBookletDetail(ctx, req)
	if err != nil {
		logrus.WithError(err).Warn("KblogClient.GetBookletDetail error")
		return nil, convertKblogError(err)
	}
	return resp.Msg, nil
}

// CreateDiary はkblogに日記を作成する。
func (kc *KblogClient) CreateDiary(ctx context.Context, authToken, title, bodyText, bodyJSON string) (*kblogdiaryv1.CreateDiaryResponse, error) {
	req := connect.NewRequest(&kblogdiaryv1.CreateDiaryRequest{
		DiaryTitle:    title,
		DiaryBodyText: bodyText,
		DiaryBodyJson: bodyJSON,
		PublishLevel:  kblogdiaryv1.PublishLevel_PUBLISH_LEVEL_ONLY_SELF,
		DiarySaveType: kblogdiaryv1.DiarySaveType_DIARY_SAVE_TYPE_PUBLISH,
	})
	req.Header().Set("Authorization", authToken)

	resp, err := kc.diaryClient.CreateDiary(ctx, req)
	if err != nil {
		logrus.WithError(err).Warn("KblogClient.CreateDiary error")
		return nil, convertKblogError(err)
	}
	return resp.Msg, nil
}

// AddDiaryToBooklet は日記を冊子に追加する。
func (kc *KblogClient) AddDiaryToBooklet(ctx context.Context, authToken string, bookletID int64, diaryCd string) error {
	req := connect.NewRequest(&kblogbookletv1.AddDiaryToBookletRequest{
		BookletId: bookletID,
		DiaryCd:   diaryCd,
	})
	req.Header().Set("Authorization", authToken)

	_, err := kc.bookletClient.AddDiaryToBooklet(ctx, req)
	if err != nil {
		logrus.WithError(err).Warn("KblogClient.AddDiaryToBooklet error")
		return convertKblogError(err)
	}
	return nil
}

// convertKblogError は connect.Error を KblogError に変換する。
func convertKblogError(err error) error {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return &KblogError{Code: "ERR_KBLOG_UNKNOWN", Message: err.Error()}
	}

	// CodeUnauthenticated は kblog アカウント未存在 (E6) の可能性が高い
	if connectErr.Code() == connect.CodeUnauthenticated {
		return ErrKblogAccountNotFound
	}

	return &KblogError{Code: "ERR_KBLOG_UNKNOWN", Message: connectErr.Message()}
}
