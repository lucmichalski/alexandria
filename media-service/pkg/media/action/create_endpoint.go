package action

import (
	"context"
	"github.com/alexandria-oss/core/middleware"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/maestre3d/alexandria/media-service/internal/domain"
	"github.com/maestre3d/alexandria/media-service/pkg/media/usecase"
	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
)

type CreateRequest struct {
	Title        string `json:"title"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	LanguageCode string `json:"language_code"`
	PublisherID  string `json:"publisher_id"`
	AuthorID     string `json:"author_id"`
	PublishDate  string `json:"publish_date"`
	MediaType    string `json:"media_type"`
}

type CreateResponse struct {
	Media *domain.Media `json:"media"`
	Err   error         `json:"-"`
}

func MakeCreateMediaEndpoint(svc usecase.MediaInteractor, logger log.Logger, duration metrics.Histogram,
	tracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer) endpoint.Endpoint {
	ep := func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(CreateRequest)

		createdMedia, err := svc.Create(ctx, &domain.MediaAggregate{
			Title:        req.Title,
			DisplayName:  req.DisplayName,
			Description:  req.Description,
			LanguageCode: req.LanguageCode,
			PublisherID:  req.PublisherID,
			AuthorID:     req.AuthorID,
			PublishDate:  req.PublishDate,
			MediaType:    req.MediaType,
		})
		if err != nil {
			return CreateResponse{
				Media: nil,
				Err:   err,
			}, err
		}

		return CreateResponse{
			Media: createdMedia,
			Err:   nil,
		}, nil
	}

	// Required resiliency and instrumentation
	action := "create"
	ep = middleware.WrapResiliency(ep, "media", action)
	return middleware.WrapInstrumentation(ep, "media", action, &middleware.WrapInstrumentParams{
		Logger:       logger,
		Duration:     duration,
		Tracer:       tracer,
		ZipkinTracer: zipkinTracer,
	})
}

// compile time assertions for our response types implementing endpoint.Failer.
var (
	_ endpoint.Failer = CreateResponse{}
)

func (r CreateResponse) Failed() error { return r.Err }
