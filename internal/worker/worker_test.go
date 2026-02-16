package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockWorkerQuerier struct {
	mock.Mock
}

func (m *mockWorkerQuerier) RebuildAllUserProfiles(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestProfileUpdater_updateAllProfiles(t *testing.T) {
	mockQueries := new(mockWorkerQuerier)
	updater := NewProfileUpdater(mockQueries)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mockQueries.On("RebuildAllUserProfiles", ctx).Return(nil).Once()
		err := updater.updateAllProfiles(ctx)
		assert.NoError(t, err)
		mockQueries.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockQueries.On("RebuildAllUserProfiles", ctx).Return(errors.New("db error")).Once()
		err := updater.updateAllProfiles(ctx)
		assert.Error(t, err)
		mockQueries.AssertExpectations(t)
	})
}

func TestInitializeRedis(t *testing.T) {
	rd := InitializeRedis()
	assert.NotNil(t, rd)
}

func TestProfileUpdater_Start(t *testing.T) {
	mockQueries := new(mockWorkerQuerier)
	updater := NewProfileUpdater(mockQueries)
	c := updater.Start(context.Background())
	assert.NotNil(t, c)
	c.Stop()
}
