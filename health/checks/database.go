package checks

import (
	"context"
	"database/sql"
	"time"

	"github.com/katalabut/fast-app/health"
)

// DatabaseOptions contains options for database health check
type DatabaseOptions struct {
	PingTimeout time.Duration
	Query       string // optional custom query instead of ping
}

// DatabaseCheck checks database connectivity
type DatabaseCheck struct {
	name string
	db   *sql.DB
	opts DatabaseOptions
}

// NewDatabaseCheck creates a new database health check
func NewDatabaseCheck(name string, db *sql.DB) *DatabaseCheck {
	return &DatabaseCheck{
		name: name,
		db:   db,
		opts: DatabaseOptions{
			PingTimeout: 5 * time.Second,
		},
	}
}

// NewDatabaseCheckWithOptions creates a new database health check with options
func NewDatabaseCheckWithOptions(name string, db *sql.DB, opts DatabaseOptions) *DatabaseCheck {
	if opts.PingTimeout == 0 {
		opts.PingTimeout = 5 * time.Second
	}
	
	return &DatabaseCheck{
		name: name,
		db:   db,
		opts: opts,
	}
}

// Name returns the name of the health check
func (d *DatabaseCheck) Name() string {
	return d.name
}

// Check executes the database health check
func (d *DatabaseCheck) Check(ctx context.Context) health.HealthResult {
	start := time.Now()
	
	// Create a context with timeout
	checkCtx, cancel := context.WithTimeout(ctx, d.opts.PingTimeout)
	defer cancel()

	var err error
	
	if d.opts.Query != "" {
		// Execute custom query
		err = d.executeQuery(checkCtx)
	} else {
		// Use ping
		err = d.db.PingContext(checkCtx)
	}

	duration := time.Since(start)

	if err != nil {
		if checkCtx.Err() == context.DeadlineExceeded {
			return health.NewUnhealthyResult("database ping timeout").
				WithDetails("timeout", d.opts.PingTimeout.String()).
				WithDetails("duration", duration.String()).
				WithDuration(duration)
		}
		
		return health.NewUnhealthyResult("database ping failed").
			WithDetails("error", err.Error()).
			WithDetails("duration", duration.String()).
			WithDuration(duration)
	}

	result := health.NewHealthyResult("database connection successful").
		WithDetails("duration", duration.String()).
		WithDuration(duration)

	// Check if response time is concerning
	if duration > d.opts.PingTimeout/2 {
		result = health.NewDegradedResult("database connection slow").
			WithDetails("duration", duration.String()).
			WithDetails("threshold", (d.opts.PingTimeout / 2).String()).
			WithDuration(duration)
	}

	return result
}

// executeQuery executes a custom query for health check
func (d *DatabaseCheck) executeQuery(ctx context.Context) error {
	rows, err := d.db.QueryContext(ctx, d.opts.Query)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Just check if we can execute the query, don't need to process results
	return nil
}
