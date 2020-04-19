package domain

import "github.com/maestre3d/alexandria/media-service/internal/shared/domain/util"

type IMediaRepository interface {
	Save(media *MediaEntity) error
	Fetch(params *util.PaginationParams, filterMap util.FilterParams) ([]*MediaEntity, error)
	FetchByID(id string) (*MediaEntity, error)
	Update(mediaUpdated *MediaEntity) error
	Remove(id string) error
}