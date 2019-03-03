package usecases_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tadoku/api/domain"
	"github.com/tadoku/api/usecases"

	gomock "github.com/golang/mock/gomock"
)

func setupContestTest(t *testing.T) (
	*gomock.Controller,
	*usecases.MockContestRepository,
	*usecases.MockValidator,
	usecases.ContestInteractor,
) {
	ctrl := gomock.NewController(t)

	repo := usecases.NewMockContestRepository(ctrl)
	validator := usecases.NewMockValidator(ctrl)
	interactor := usecases.NewContestInteractor(repo, validator)

	return ctrl, repo, validator, interactor
}

func TestContestInteractor_CreateContest(t *testing.T) {
	ctrl, repo, validator, interactor := setupContestTest(t)
	defer ctrl.Finish()

	{
		contest := domain.Contest{
			Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
			Open:  true,
		}

		repo.EXPECT().Store(contest)
		repo.EXPECT().GetOpenContests().Return(nil, nil)
		validator.EXPECT().Validate(contest).Return(true, nil)

		err := interactor.CreateContest(contest)

		assert.NoError(t, err)
	}

	{
		contest := domain.Contest{
			Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
			Open:  true,
		}

		repo.EXPECT().GetOpenContests().Return([]uint64{1}, usecases.ErrOpenContestAlreadyExists)
		validator.EXPECT().Validate(contest).Return(true, nil)

		err := interactor.CreateContest(contest)

		assert.Error(t, err, usecases.ErrOpenContestAlreadyExists)
	}

	{
		contest := domain.Contest{
			Start: time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:  true,
		}

		validator.EXPECT().Validate(contest).Return(false, usecases.ErrInvalidContest)

		err := interactor.CreateContest(contest)

		assert.Error(t, err)
	}
}

func TestContestInteractor_UpdateContest(t *testing.T) {
	ctrl, repo, validator, interactor := setupContestTest(t)
	defer ctrl.Finish()

	{
		contest := domain.Contest{
			ID:    1,
			Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
			Open:  false,
		}

		repo.EXPECT().Store(contest)
		validator.EXPECT().Validate(contest).Return(true, nil)

		err := interactor.UpdateContest(contest)

		assert.NoError(t, err)
	}

	{
		contest := domain.Contest{
			Start: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2019, 1, 31, 0, 0, 0, 0, time.UTC),
			Open:  false,
		}

		err := interactor.UpdateContest(contest)

		assert.Error(t, err, usecases.ErrContestIDMissing)
	}
}