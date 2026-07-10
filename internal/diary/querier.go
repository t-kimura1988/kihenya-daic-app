package diary

import (
	"context"

	kblogaccountv1 "github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/account"
	kblogbookletv1 "github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/booklet"
	kblogdiaryv1 "github.com/t-kimura1988/kihenya-blog-proto/gen/go/kblog/v1/diary"

	"github.com/t-kimura1988/kihenya-daic-app/internal/db/sqlc"
)

// DiaryQuerier defines the database operations used by the diary service.
type DiaryQuerier interface {
	GetIdealByID(ctx context.Context, arg sqlc.GetIdealByIDParams) (sqlc.TIdeal, error)
	GetKblogBookletLink(ctx context.Context, arg sqlc.GetKblogBookletLinkParams) (sqlc.TKblogBookletLink, error)
	CreateKblogBookletLink(ctx context.Context, arg sqlc.CreateKblogBookletLinkParams) (sqlc.TKblogBookletLink, error)
}

// KblogCaller defines the external kblog service operations used by the diary service.
type KblogCaller interface {
	GetAccount(ctx context.Context, authToken string) (*kblogaccountv1.GetAccountResponse, error)
	GetBookletDetail(ctx context.Context, authToken string, bookletID int64) (*kblogbookletv1.GetBookletDetailResponse, error)
	CreateBooklet(ctx context.Context, authToken, title, description string) (*kblogbookletv1.CreateBookletResponse, error)
	CreateDiary(ctx context.Context, authToken, title, bodyText, bodyJSON string) (*kblogdiaryv1.CreateDiaryResponse, error)
	AddDiaryToBooklet(ctx context.Context, authToken string, bookletID int64, diaryCd string) error
}
