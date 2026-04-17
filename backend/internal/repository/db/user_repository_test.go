package db

import (
	"context"
	"regexp"
	"testing"

	"code-repo/internal/model/entity"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	dialector := postgres.New(postgres.Config{Conn: sqlDB})
	db, err := gorm.Open(dialector, &gorm.Config{})
	assert.NoError(t, err)
	return db, mock
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewUserRepository(db)
	user := &entity.User{Email: "test@example.com", Nickname: "testuser"}

	// 预期执行 INSERT 语句
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
