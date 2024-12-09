package postgresql

import (
	"database/sql"
	"time"

	// Blank import required to register driver
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Service common functionality shared between the postgresql and postgresql_extensible
// packages.
type Service struct {
	DB                 *sql.DB
	SanitizedAddress   string
	ConnectionDatabase string

	dsn         string
	maxIdle     int
	maxOpen     int
	maxLifetime time.Duration
}

func (p *Service) Start() error {
	db, err := sql.Open("pgx", p.dsn)
	if err != nil {
		return err
	}
	p.DB = db

	p.DB.SetMaxOpenConns(p.maxOpen)
	p.DB.SetMaxIdleConns(p.maxIdle)
	p.DB.SetConnMaxLifetime(p.maxLifetime)

	return nil
}

func (p *Service) Stop() {
	if p.DB != nil {
		p.DB.Close()
	}
}
