package handler

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/dityuiri/UserServiceTest/repository"
)

func TestNewServer(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		mockRepo = repository.NewMockRepositoryInterface(mockCtrl)
	)

	t.Run("positive", func(t *testing.T) {
		sv := NewServer(NewServerOptions{Repository: mockRepo})
		assert.NotEmpty(t, sv.Repository)
	})
}
