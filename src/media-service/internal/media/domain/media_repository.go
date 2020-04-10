package domain

import "github.com/maestre3d/alexandria/src/media-service/internal/shared/domain/util"

type IMediaRepository interface {
	Save(book *MediaAggregate) error
	Fetch(params *util.PaginationParams) ([]*MediaAggregate, error)
	FetchByID(id int64) (*MediaAggregate, error)
	FetchByTitle(title string) (*MediaAggregate, error)
	UpdateOne(id int64, bookUpdated *MediaAggregate) error
	RemoveOne(id int64) error
}
