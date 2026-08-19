package postgresql

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

// Service common functionality shared between the postgresql and postgresql_extensible
// packages.
type Service struct {
	DB                 *sql.DB
	SanitizedAddress   string
	ConnectionDatabase string

	connector   driver.Connector
	maxIdle     int
	maxOpen     int
	maxLifetime time.Duration
}

func (p *Service) Start() error {
	p.DB = sql.OpenDB(p.connector)

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
