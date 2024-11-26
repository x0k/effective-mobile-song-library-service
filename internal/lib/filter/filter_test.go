package filter

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestFilter_Parse(t *testing.T) {
	filter := New("test", map[string]ValueType{
		"string_column": StringType,
		"array_column":  ArrayOf(StringType),
		"date_column":   DateType,
	}, func(s string) (any, error) {
		return s, nil
	})
	tests := []struct {
		name     string
		input    string
		wantSql  string
		wantArgs []any
		err      error
	}{
		{
			name: "empty",
			err:  ErrInvalidExpression,
		},
		{
			name:     "eq",
			input:    "EQ(1, 1)",
			wantSql:  `$1 = $2`,
			wantArgs: []any{int64(1), int64(1)},
		},
		{
			name:     "in",
			input:    "IN(1, (2, 3, 4))",
			wantSql:  `$1 IN ($2, $3, $4)`,
			wantArgs: []any{int64(1), int64(2), int64(3), int64(4)},
		},
		{
			name:     "in with array column",
			input:    `IN("value", array_column)`,
			wantSql:  `$1 = ANY("test"."array_column")`,
			wantArgs: []any{"value"},
		},
		{
			name:     "gt",
			input:    "GT(1, 2)",
			wantSql:  `$1 > $2`,
			wantArgs: []any{int64(1), int64(2)},
		},
		{
			name:     "or",
			input:    `OR(EQ(1, 1), EQ(2, 2), EQ(3, 3))`,
			wantSql:  `($1 = $2 OR $3 = $4 OR $5 = $6)`,
			wantArgs: []any{int64(1), int64(1), int64(2), int64(2), int64(3), int64(3)},
		},
		{
			name:     "like",
			input:    `LIKE(string_column, "%pattern%")`,
			wantSql:  `"test"."string_column" ILIKE $1`,
			wantArgs: []any{`%pattern%`},
		},
		{
			name:    "alike",
			input:   `ALIKE(array_column, "%pattern%")`,
			wantSql: `EXISTS (SELECT 1 FROM unnest("test"."array_column") AS element WHERE element ILIKE $1)`,
			wantArgs: []any{
				`%pattern%`,
			},
		},
		{
			name:    "date",
			input:   `EQ(date_column, DATE("2022-01-01"))`,
			wantSql: `"test"."date_column" = $1`,
			wantArgs: []any{
				"2022-01-01",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := filter.Parse(tt.input)
			if err != nil {
				if !errors.Is(err, tt.err) {
					t.Errorf("Filter.Parse() error = %v, wantErr %v", err, tt.err)
				}
				return
			}
			b := strings.Builder{}
			args := got.ToSQL(&b, nil)
			sql := b.String()
			if sql != tt.wantSql {
				t.Errorf("Filter.Parse() = %v, want sql %v", sql, tt.wantSql)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("Filter.Parse() = %v, want args %v", args, tt.wantArgs)
			}
		})
	}
}
