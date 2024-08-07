package migrations

// Migrator is an interface for database migrations.
type Migrator interface {
	// Up applies the migration and returns whether it was applied and any error that occurred.
	Up(applied bool, err error)
	// Down rolls back the migration and returns whether it was rolled back and any error that occurred.
	Down(applied bool, err error)
}
