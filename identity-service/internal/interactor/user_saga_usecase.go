package interactor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexandria-oss/core/exception"
	"github.com/go-kit/kit/log"
	"github.com/maestre3d/alexandria/identity-service/internal/domain"
)

type UserSAGA struct {
	logger     log.Logger
	repository domain.UserRepository
	event      domain.UserEventSAGA
}

func NewUserSAGA(logger log.Logger, repo domain.UserRepository, event domain.UserEventSAGA) *UserSAGA {
	return &UserSAGA{
		logger:     logger,
		repository: repo,
		event:      event,
	}
}

// Verifier implementation

func (u *UserSAGA) Verify(ctx context.Context, usersJSON []byte) error {
	ctxR, cancel := context.WithCancel(ctx)
	defer cancel()

	// owners contains just an array with users id's string
	var owners []string
	err := json.Unmarshal(usersJSON, &owners)
	if err != nil {
		_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())

		// Rollback if format error
		err = u.event.Failed(ctxR)
		if err != nil {
			// Error during publishing
			_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())
			return err
		}

		_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
		return exception.NewErrorDescription(exception.InvalidFieldFormat, fmt.Sprintf(exception.InvalidFieldFormatString,
			"owner_pool", "[]string"))
	}

	for _, id := range owners {
		_, err := u.repository.FetchByID(ctxR, id)
		if err != nil {
			_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())

			err = u.event.Failed(ctxR)
			if err != nil {
				_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())
				return err
			}

			_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerFailed+" integration event published")
			return nil
		}
	}

	// All users have been verified
	err = u.event.Verified(ctxR)
	if err != nil {
		_ = u.logger.Log("method", "identity.interactor.saga.verify", "err", err.Error())
		return err
	}

	_ = u.logger.Log("method", "identity.interactor.saga.verify", "msg", domain.OwnerVerified+" integration event published")
	return nil
}
