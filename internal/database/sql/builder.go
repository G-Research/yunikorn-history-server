package sql

import (
	"fmt"
	"strings"
)

type OrderDirection string

const (
	OrderByAscending  OrderDirection = "ASC"
	OrderByDescending OrderDirection = "DESC"
)

type Builder struct {
	selectStatement  string
	whereClauses     string
	hasWhere         bool
	conditionCounter int
	args             []any
	orderByClauses   string
	limit            string
	offset           string
}

func NewBuilder() *Builder {
	return &Builder{args: make([]any, 0)}
}

// SelectAll creates a new query with a SELECT statement which selects all ('*') entities.
// If an alias is provided, it will be used as the table alias.
func (b *Builder) SelectAll(table string, alias string) *Builder {
	b.selectStatement = fmt.Sprintf("SELECT * FROM %s", table)
	if alias != "" {
		b.selectStatement += " AS " + alias
	}
	return b
}

// Conditionf adds a condition to the query.
// This function accepts a format string and arguments, which will be used to create the condition.
// If the condition uses special characters like '%', they need to be escaped like '%%'.
// Example: Conditionf("name = '%s'", "John")
func (b *Builder) Conditionf(format string, args ...any) *Builder {
	expression := fmt.Sprintf(format, args...)
	return b.condition(expression)
}

// Condition adds a condition to the query.
// This function accepts a raw expression string.
// Special characters like '%' do not need to be escaped.
// Example: Condition("name = '@name'")
func (b *Builder) Condition(expression string) *Builder {
	return b.condition(expression)
}

// Conditionp adds a condition to the query as a positional argument ('$1', '$2'...).
// lhs is the left-hand side of the condition, op is the operator and val is the value.
//
// Example: Conditionp("name", "=", "John") will be added as "name = '$1'".
func (b *Builder) Conditionp(lhs, op string, val any) *Builder {
	b.conditionCounter++
	expression := fmt.Sprintf("%s %s $%d", lhs, op, b.conditionCounter)
	b.args = append(b.args, val)
	return b.condition(expression)
}

func (b *Builder) condition(expression string) *Builder {
	if !b.hasWhere {
		b.whereClauses += "WHERE " + expression
		b.hasWhere = true
		return b
	}
	b.whereClauses += " AND " + expression
	return b
}

// Limit adds a LIMIT clause to the query.
func (b *Builder) Limit(limit int) *Builder {
	b.limit = fmt.Sprintf("LIMIT %d", limit)
	return b
}

// Offset adds an OFFSET clause to the query.
func (b *Builder) Offset(offset int) *Builder {
	b.offset = fmt.Sprintf("OFFSET %d", offset)
	return b
}

func (b *Builder) OrderBy(column string, direction OrderDirection) *Builder {
	b.orderByClauses = fmt.Sprintf("ORDER BY %s %s", column, direction)
	return b
}

// Query returns the final query string.
func (b *Builder) Query() string {
	query := strings.Builder{}
	query.WriteString(b.selectStatement)
	if b.whereClauses != "" {
		query.WriteString(" ")
		query.WriteString(b.whereClauses)
	}
	if b.orderByClauses != "" {
		query.WriteString(" ")
		query.WriteString(b.orderByClauses)
	}
	if b.limit != "" {
		query.WriteString(" ")
		query.WriteString(b.limit)
	}
	if b.offset != "" {
		query.WriteString(" ")
		query.WriteString(b.offset)
	}
	return query.String()
}

// Args returns the positional arguments used in the query.
func (b *Builder) Args() []any {
	return b.args
}
