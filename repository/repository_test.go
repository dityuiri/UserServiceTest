package repository

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		// Define valid options for the repository
		validOptions := NewRepositoryOptions{
			Dsn: "valid_dsn",
		}

		repository := NewRepository(validOptions)

		assert.NotNil(t, repository)
		assert.NotNil(t, repository.Db)
		defer func(db *sql.DB) {
			_ = db.Close()
		}(repository.Db)
	})
}
