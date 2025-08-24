package db

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

// Opt определяет тип функции, которая применяет конфигурацию к sqlx.DB.
type Opt func(*sqlx.DB)

// New создаёт подключение к базе данных и применяет переданные опции.
func New(driver string, dsn string, opts ...Opt) (*sqlx.DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}

	// Применяем все функциональные опции
	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}

// WithMaxOpenConns устанавливает максимальное количество открытых соединений.
func WithMaxOpenConns(opts ...int) Opt {
	return func(db *sqlx.DB) {
		for _, opt := range opts {
			if opt > 0 {
				db.SetMaxOpenConns(opt)
				break
			}
		}
	}
}

// WithMaxIdleConns устанавливает максимальное количество простаивающих соединений.
func WithMaxIdleConns(opts ...int) Opt {
	return func(db *sqlx.DB) {
		for _, opt := range opts {
			if opt > 0 {
				db.SetMaxIdleConns(opt)
				break
			}
		}
	}
}
