package diary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/t-kimura1988/kihenya-daic-app/external/clients"
	"github.com/t-kimura1988/kihenya-daic-app/internal/account"
	"github.com/t-kimura1988/kihenya-daic-app/internal/db/sqlc"
	diaryv1 "github.com/t-kimura1988/kihenya-daic-proto/gen/go/daic/v1/diary"
)

// Service implements the DiaryService (ConnectRPC).
// 理想(t_ideal)ごとにkihenya-blog(kblog)の冊子を連携し、日々の感想を日記として同期する。
type Service struct {
	queries     DiaryQuerier
	kblogClient KblogCaller
}

// NewService creates a new diary service.
func NewService(queries DiaryQuerier, kblogClient KblogCaller) *Service {
	return &Service{
		queries:     queries,
		kblogClient: kblogClient,
	}
}

// GetDiaryStatus は理想の日記連携状態（冊子リンク / 日記一覧）を取得する。
func (s *Service) GetDiaryStatus(ctx context.Context, req *connect.Request[diaryv1.GetDiaryStatusRequest]) (*connect.Response[diaryv1.GetDiaryStatusResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, authHeader, err := s.authContext(ctx)
	if err != nil {
		return nil, err
	}

	idealID := req.Msg.IdealId

	// 理想の所有権確認
	if err := s.ensureIdealOwnership(ctx, idealID, accountID); err != nil {
		return nil, err
	}

	// kblog冊子連携を確認
	link, err := s.queries.GetKblogBookletLink(ctx, sqlc.GetKblogBookletLinkParams{
		IdealID:   idealID,
		AccountID: accountID,
	})
	if err != nil {
		// 冊子未連携 → kblogアカウント存在確認
		if _, kblogErr := s.kblogClient.GetAccount(ctx, authHeader); kblogErr != nil {
			if clients.IsKblogAccountNotFound(kblogErr) {
				return connect.NewResponse(&diaryv1.GetDiaryStatusResponse{
					ErrorCd: "ERR_KBLOG_ACCOUNT_NOT_FOUND",
					Error:   "kblog account not found",
				}), nil
			}
			logrus.WithError(kblogErr).Error("diary service: failed to get kblog account")
			return connect.NewResponse(&diaryv1.GetDiaryStatusResponse{
				ErrorCd: "ERR_KBLOG_UNKNOWN",
				Error:   "failed to check kblog account",
			}), nil
		}

		// kblogアカウントは存在するが冊子未連携
		return connect.NewResponse(&diaryv1.GetDiaryStatusResponse{
			ErrorCd: "ERR_KBLOG_BOOKLET_NOT_LINKED",
			Error:   "kblog booklet not linked",
		}), nil
	}

	// 冊子が連携済み → 日記一覧を取得
	detail, err := s.kblogClient.GetBookletDetail(ctx, authHeader, link.KblogBookletID)
	if err != nil {
		if clients.IsKblogAccountNotFound(err) {
			return connect.NewResponse(&diaryv1.GetDiaryStatusResponse{
				ErrorCd: "ERR_KBLOG_ACCOUNT_NOT_FOUND",
				Error:   "kblog account not found",
			}), nil
		}
		logrus.WithError(err).Error("diary service: failed to get booklet detail")
		return connect.NewResponse(&diaryv1.GetDiaryStatusResponse{
			ErrorCd: "ERR_KBLOG_UNKNOWN",
			Error:   "failed to get diary list",
		}), nil
	}

	diaries := make([]*diaryv1.DiaryItem, 0, len(detail.Diaries))
	for _, d := range detail.Diaries {
		diaries = append(diaries, &diaryv1.DiaryItem{
			DiaryId:         d.DiaryId,
			DiaryCd:         d.DiaryCd,
			DiaryCreateDate: d.DiaryCreateDate,
			DiaryTitle:      d.DiaryTitle,
			DiaryBodyText:   d.DiaryBodyText,
			PublishLevel:    d.PublishLevel,
		})
	}

	return connect.NewResponse(&diaryv1.GetDiaryStatusResponse{
		BookletId: link.KblogBookletID,
		Diaries:   diaries,
	}), nil
}

// CreateDiary は新しい日記（感想）を作成し、連携中の冊子に追加する。
func (s *Service) CreateDiary(ctx context.Context, req *connect.Request[diaryv1.CreateDiaryRequest]) (*connect.Response[diaryv1.CreateDiaryResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, authHeader, err := s.authContext(ctx)
	if err != nil {
		return nil, err
	}

	idealID := req.Msg.IdealId

	// 理想の所有権確認
	if err := s.ensureIdealOwnership(ctx, idealID, accountID); err != nil {
		return nil, err
	}

	// 冊子リンク確認
	link, err := s.queries.GetKblogBookletLink(ctx, sqlc.GetKblogBookletLinkParams{
		IdealID:   idealID,
		AccountID: accountID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("kblog booklet not linked"))
	}

	// 日記作成
	created, err := s.kblogClient.CreateDiary(ctx, authHeader, req.Msg.Title, req.Msg.BodyText, req.Msg.BodyJson)
	if err != nil {
		if clients.IsKblogAccountNotFound(err) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("kblog account not found"))
		}
		logrus.WithError(err).Error("diary service: failed to create diary")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create diary"))
	}

	// 日記を冊子に追加（日記作成は成功しているため失敗しても警告のみ）
	if err := s.kblogClient.AddDiaryToBooklet(ctx, authHeader, link.KblogBookletID, created.DiaryCd); err != nil {
		logrus.WithError(err).Warn("diary service: failed to add diary to booklet")
	}

	return connect.NewResponse(&diaryv1.CreateDiaryResponse{
		DiaryId:         created.DiaryId,
		DiaryCd:         created.DiaryCd,
		DiaryCreateDate: created.DiaryCreateDate,
	}), nil
}

// LinkBooklet はkblog冊子を作成して理想に連携する。
func (s *Service) LinkBooklet(ctx context.Context, req *connect.Request[diaryv1.LinkBookletRequest]) (*connect.Response[diaryv1.LinkBookletResponse], error) {
	if err := validate(req.Msg); err != nil {
		return nil, err
	}

	accountID, authHeader, err := s.authContext(ctx)
	if err != nil {
		return nil, err
	}

	idealID := req.Msg.IdealId

	// 理想の所有権確認
	ideal, err := s.queries.GetIdealByID(ctx, sqlc.GetIdealByIDParams{
		ID:        idealID,
		AccountID: accountID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("ideal not found"))
		}
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}

	// kblogに冊子を作成
	bookletTitle := fmt.Sprintf("%sの記録", ideal.Title)
	created, err := s.kblogClient.CreateBooklet(ctx, authHeader, bookletTitle, "Daic（大工）の理想と連携した記録冊子")
	if err != nil {
		if clients.IsKblogAccountNotFound(err) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("kblog account not found"))
		}
		logrus.WithError(err).Error("diary service: failed to create booklet")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to create booklet"))
	}

	// t_kblog_booklet_link に保存
	if _, err := s.queries.CreateKblogBookletLink(ctx, sqlc.CreateKblogBookletLinkParams{
		IdealID:        idealID,
		AccountID:      accountID,
		KblogBookletID: created.BookletId,
	}); err != nil {
		logrus.WithError(err).Error("diary service: failed to save booklet link")
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to save booklet link"))
	}

	return connect.NewResponse(&diaryv1.LinkBookletResponse{
		BookletId:    created.BookletId,
		BookletTitle: created.BookletTitle,
	}), nil
}

// authContext はコンテキストからaccountIDとauthヘッダーを取り出す。
func (s *Service) authContext(ctx context.Context) (int64, string, error) {
	accountID, err := account.GetAccountIDFromContext(ctx)
	if err != nil {
		return 0, "", connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("account not authenticated"))
	}
	authHeader, err := account.GetAuthHeaderFromContext(ctx)
	if err != nil {
		return 0, "", connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("auth header not found"))
	}
	return accountID, authHeader, nil
}

// ensureIdealOwnership はログインアカウントが理想の所有者であることを確認する。
func (s *Service) ensureIdealOwnership(ctx context.Context, idealID, accountID int64) error {
	if _, err := s.queries.GetIdealByID(ctx, sqlc.GetIdealByIDParams{
		ID:        idealID,
		AccountID: accountID,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("ideal not found"))
		}
		return connect.NewError(connect.CodePermissionDenied, fmt.Errorf("forbidden"))
	}
	return nil
}

func validate(msg proto.Message) error {
	validator, err := protovalidate.New()
	if err != nil {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("validation error: %w", err))
	}
	if err := validator.Validate(msg); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("validation failed: %w", err))
	}
	return nil
}
