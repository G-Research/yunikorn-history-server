package health

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/G-Research/yunikorn-history-server/internal/yunikorn"
)

type ComponentStatus struct {
	Identifier string `json:"identifier"`
	Healthy    bool   `json:"healthy"`
	Error      string `json:"error,omitempty"`
}

type Component interface {
	Identifier() string
	Check(context.Context) *ComponentStatus
}

type YunikornComponent struct {
	c yunikorn.Client
}

func NewYunikornComponent(client yunikorn.Client) *YunikornComponent {
	return &YunikornComponent{c: client}
}

func (c *YunikornComponent) Identifier() string {
	return "yunikorn"
}

func (c *YunikornComponent) Check(ctx context.Context) *ComponentStatus {
	s := &ComponentStatus{Identifier: c.Identifier()}
	_, err := c.c.Healthcheck(ctx)
	if err != nil {
		s.Error = err.Error()
		return s
	}
	s.Healthy = true
	return s
}

type PostgresComponent struct {
	pool *pgxpool.Pool
}

func NewPostgresComponent(pool *pgxpool.Pool) *PostgresComponent {
	return &PostgresComponent{pool: pool}
}

func (c *PostgresComponent) Identifier() string {
	return "postgres"
}

func (c *PostgresComponent) Check(ctx context.Context) *ComponentStatus {
	s := &ComponentStatus{Identifier: c.Identifier()}
	_, err := c.pool.Exec(ctx, "SELECT 1")
	if err != nil {
		s.Error = err.Error()
		return s
	}
	s.Healthy = true
	return s
}
