package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectAll(t *testing.T) {
	tests := []struct {
		name     string
		table    string
		alias    string
		expected string
	}{
		{"Without alias", "users", "", "SELECT * FROM users"},
		{"With alias", "users", "u", "SELECT * FROM users AS u"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder().SelectAll(tt.table, tt.alias)
			assert.Equal(t, tt.expected, b.selectStatement)
			assert.Empty(t, b.whereClauses)
			assert.False(t, b.hasWhere)
			assert.Empty(t, b.orderByClauses)
			assert.Empty(t, b.limit)
			assert.Empty(t, b.offset)
		})
	}
}

func TestConditionf(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Builder)
		expected string
	}{
		{
			"Single condition",
			func(b *Builder) {
				b.Conditionf("age > %d", 30)
			},
			"WHERE age > 30",
		},
		{
			"Multiple conditions",
			func(b *Builder) {
				b.Conditionf("name = '%s'", "John")
				b.Conditionf("age > %d", 30)
			},
			"WHERE name = 'John' AND age > 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			tt.setup(b)
			assert.Equal(t, tt.expected, b.whereClauses)
			assert.Empty(t, b.selectStatement)
			assert.True(t, b.hasWhere)
			assert.Empty(t, b.orderByClauses)
			assert.Empty(t, b.limit)
			assert.Empty(t, b.offset)
		})
	}
}

func TestLimit(t *testing.T) {
	tests := []struct {
		name     string
		limit    int
		expected string
	}{
		{"Limit 10", 10, "LIMIT 10"},
		{"Limit 50", 50, "LIMIT 50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder().Limit(tt.limit)
			assert.Equal(t, tt.expected, b.limit)
			assert.Empty(t, b.selectStatement)
			assert.Empty(t, b.whereClauses)
			assert.False(t, b.hasWhere)
			assert.Empty(t, b.orderByClauses)
			assert.Empty(t, b.offset)
		})
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		expected string
	}{
		{"Offset 10", 10, "OFFSET 10"},
		{"Offset 50", 50, "OFFSET 50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder().Offset(tt.offset)
			assert.Equal(t, tt.expected, b.offset)
			assert.Empty(t, b.selectStatement)
			assert.Empty(t, b.whereClauses)
			assert.False(t, b.hasWhere)
			assert.Empty(t, b.orderByClauses)
			assert.Empty(t, b.limit)
		})
	}
}

func TestOrderBy(t *testing.T) {
	tests := []struct {
		name      string
		column    string
		direction OrderDirection
		expected  string
	}{
		{"Order by ascending", "name", OrderByAscending, "ORDER BY name ASC"},
		{"Order by descending", "age", OrderByDescending, "ORDER BY age DESC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder().OrderBy(tt.column, tt.direction)
			assert.Equal(t, tt.expected, b.orderByClauses)
			assert.Empty(t, b.selectStatement)
			assert.Empty(t, b.whereClauses)
			assert.False(t, b.hasWhere)
			assert.Empty(t, b.offset)
			assert.Empty(t, b.limit)
		})
	}
}

func TestCombined(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Builder)
		expected string
	}{
		{
			name: "SelectAll, Conditionf, Limit, Offset, OrderBy",
			setup: func(b *Builder) {
				b.SelectAll("users", "u").
					Conditionf("age > %d", 30).
					Conditionf("name = '%s'", "John").
					OrderBy("age", OrderByAscending).
					Limit(10).
					Offset(5)
			},
			expected: "SELECT * FROM users AS u WHERE age > 30 AND name = 'John' ORDER BY age ASC LIMIT 10 OFFSET 5",
		},
		{
			name: "SelectAll with alias, Conditionf, Limit",
			setup: func(b *Builder) {
				b.SelectAll("products", "p").
					Conditionf("price < %f", 100.50).
					Limit(20)
			},
			expected: "SELECT * FROM products AS p WHERE price < 100.500000 LIMIT 20",
		},
		{
			name: "SelectAll without alias, OrderBy, Offset",
			setup: func(b *Builder) {
				b.SelectAll("employees", "").
					OrderBy("salary", OrderByDescending).
					Offset(10)
			},
			expected: "SELECT * FROM employees ORDER BY salary DESC OFFSET 10",
		},
		{
			name: "SelectAll, Conditionf with special characters",
			setup: func(b *Builder) {
				b.SelectAll("logs", "l").
					Conditionf("message LIKE '%%%%error%%%%'").
					OrderBy("timestamp", OrderByDescending)
			},
			expected: "SELECT * FROM logs AS l WHERE message LIKE '%%error%%' ORDER BY timestamp DESC",
		},
		{
			name: "SelectAll, multiple Conditionf, no alias",
			setup: func(b *Builder) {
				b.SelectAll("transactions", "").
					Conditionf("amount > %d", 1000).
					Conditionf("status = '%s'", "completed")
			},
			expected: "SELECT * FROM transactions WHERE amount > 1000 AND status = 'completed'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder()
			tt.setup(b)
			assert.Equal(t, tt.expected, b.Query())
		})
	}
}

func TestConditionp(t *testing.T) {
	tests := []struct {
		name               string
		lhs                string
		op                 string
		val                any
		initialConditions  string
		initialArgs        []any
		expectedConditions string
		expectedArgs       []any
	}{
		{
			name:               "Simple equality condition",
			lhs:                "name",
			op:                 "=",
			val:                "John",
			initialConditions:  "",
			initialArgs:        []any{},
			expectedConditions: "WHERE name = $1",
			expectedArgs:       []any{"John"},
		},
		{
			name:               "Greater than condition",
			lhs:                "age",
			op:                 ">",
			val:                30,
			initialConditions:  "",
			initialArgs:        []any{},
			expectedConditions: "WHERE age > $1",
			expectedArgs:       []any{30},
		},
		{
			name:               "Multiple conditions",
			lhs:                "salary",
			op:                 "<",
			val:                5000,
			initialConditions:  "WHERE name = $1",
			initialArgs:        []any{"John"},
			expectedConditions: "WHERE name = $1 AND salary < $2",
			expectedArgs:       []any{"John", 5000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				whereClauses:     tt.initialConditions,
				args:             tt.initialArgs,
				hasWhere:         tt.initialConditions != "",
				conditionCounter: len(tt.initialArgs),
			}
			b.Conditionp(tt.lhs, tt.op, tt.val)

			assert.Equal(t, tt.expectedConditions, b.whereClauses)
			assert.Equal(t, tt.expectedArgs, b.args)
		})
	}
}
