package pg

import "strings"

// Query is a helper to build queries
type Query []string

// SQL returns the query
func (q Query) SQL() string {
	return strings.Join(q, " ")
}

// Select will insert a select statement and columns
func (q Query) Select(columns ...string) Query {
	if len(q) != 0 {
		logger.Panic().Strs("query", q).
			Msg("cannot use select on a query with elements")
	}
	if len(columns) == 0 {
		columns = []string{"*"}
	}
	return append([]string{"SELECT"}, strings.Join(columns, ", "))
}

// From will add a from block
func (q Query) From(table string) Query {
	if len(q) == 0 {
		logger.Panic().Strs("query", q).
			Msg("cannot use From on a query without elements")
	}
	return append(q, "FROM", identifierNameSQL(table))
}

// Where will add a WHERE directive
func (q Query) Where() Query {
	// FROM should be in there
	if len(q) == 0 {
		logger.Panic().Strs("query", q).
			Msg("cannot use Where on a query without elements")
	}
}

// And will add an AND directive
func (q Query) And(table string) Query {
	// WHERE should be in there
	if len(q) == 0 {
		logger.Panic().Strs("query", q).
			Msg("cannot use And on a query without elements")
	}
}

// In will add an IN directive
func (q Query) In(table string) Query {
	// WHERE should be in there
	if len(q) == 0 {
		logger.Panic().Strs("query", q).
			Msg("cannot use In on a query without elements")
	}
}
