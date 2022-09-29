// Copyright 2022 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queries

import (
	"github.com/dolthub/vitess/go/sqltypes"

	"github.com/dolthub/go-mysql-server/sql"
)

var ColumnAliasQueries = []ScriptTest{
	{
		Name: "column aliases in a single scope",
		SetUpScript: []string{
			"create table xy (x int primary key, y int);",
			"create table uv (u int primary key, v int);",
			"insert into xy values (0,0),(1,1),(2,2),(3,3);",
			"insert into uv values (0,3),(3,0),(2,1),(1,2);",
		},
		Assertions: []ScriptTestAssertion{
			{
				// Projections can create expression aliases
				Query: `SELECT i AS cOl FROM mytable`,
				ExpectedColumns: sql.Schema{
					{
						Name: "cOl",
						Type: sql.Int64,
					},
				},
				Expected: []sql.Row{{int64(1)}, {int64(2)}, {int64(3)}},
			},
			{
				Query: `SELECT i AS cOl, s as COL FROM mytable`,
				ExpectedColumns: sql.Schema{
					{
						Name: "cOl",
						Type: sql.Int64,
					},
					{
						Name: "COL",
						Type: sql.MustCreateStringWithDefaults(sqltypes.VarChar, 20),
					},
				},
				Expected: []sql.Row{{int64(1), "first row"}, {int64(2), "second row"}, {int64(3), "third row"}},
			},
			{
				// Projection expressions may NOT reference aliases defined in projection expressions
				// in the same scope
				Query:       `SELECT i AS new1, new1 as new2 FROM mytable`,
				ExpectedErr: sql.ErrMisusedAlias,
			},
			{
				// The SQL standard disallows aliases in the same scope from being used in filter conditions
				Query:       `SELECT i AS cOl, s as COL FROM mytable where cOl = 1`,
				ExpectedErr: sql.ErrColumnNotFound,
			},
			{
				// Alias expressions may NOT be used in from clauses
				Query:       "select t1.i as a, t1.s as b from mytable as t1 left join mytable as t2 on a = t2.i;",
				ExpectedErr: sql.ErrColumnNotFound,
			},
			{
				// OrderBy clause may reference expression aliases at current scope
				Query:    "select 1 as a order by a desc;",
				Expected: []sql.Row{{1}},
			},
			{
				// If there is ambiguity between one table column and one alias, the alias gets precedence in the order
				// by clause. (This is different from subqueries in projection expressions.)
				Query:    "select v as u from uv order by u;",
				Expected: []sql.Row{{0}, {1}, {2}, {3}},
			},
			{
				// If there is ambiguity between multiple aliases in an order by clause, it is an error
				Query:       "select u as u, v as u from uv order by u;",
				ExpectedErr: sql.ErrAmbiguousColumnOrAliasName,
			},
			{
				// GroupBy may use expression aliases in grouping expressions
				Query: `SELECT s as COL1, SUM(i) COL2 FROM mytable group by col1 order by col2`,
				ExpectedColumns: sql.Schema{
					{
						Name: "COL1",
						Type: sql.MustCreateStringWithDefaults(sqltypes.VarChar, 20),
					},
					{
						Name: "COL2",
						Type: sql.Int64,
					},
				},
				Expected: []sql.Row{
					{"first row", float64(1)},
					{"second row", float64(2)},
					{"third row", float64(3)},
				},
			},
			{
				// Having clause may reference expression aliases current scope
				Query:    "select t1.u as a from uv as t1 having a > 0 order by a;",
				Expected: []sql.Row{{1}, {2}, {3}},
			},
			{
				// This test currently fails with error "found HAVING clause with no GROUP BY"
				Skip: true,

				// Having clause may reference expression aliases from current scope
				Query:    "select t1.u as a from uv as t1 having a = t1.u order by a;",
				Expected: []sql.Row{{0}, {1}, {2}, {3}},
			},
			{
				// Expression aliases work when implicitly referenced by ordinal position
				Query: `SELECT s as coL1, SUM(i) coL2 FROM mytable group by 1 order by 2`,
				ExpectedColumns: sql.Schema{
					{
						Name: "coL1",
						Type: sql.MustCreateStringWithDefaults(sqltypes.VarChar, 20),
					},
					{
						Name: "coL2",
						Type: sql.Int64,
					},
				},
				Expected: []sql.Row{
					{"first row", float64(1)},
					{"second row", float64(2)},
					{"third row", float64(3)},
				},
			},
			{
				// Expression aliases work when implicitly referenced by ordinal position
				Query: `SELECT s as Date, SUM(i) TimeStamp FROM mytable group by 1 order by 2`,
				ExpectedColumns: sql.Schema{
					{
						Name: "Date",
						Type: sql.MustCreateStringWithDefaults(sqltypes.VarChar, 20),
					},
					{
						Name: "TimeStamp",
						Type: sql.Int64,
					},
				},
				Expected: []sql.Row{
					{"first row", float64(1)},
					{"second row", float64(2)},
					{"third row", float64(3)},
				},
			},
		},
	},
	{
		Name: "column aliases in two scopes",
		SetUpScript: []string{
			"create table xy (x int primary key, y int);",
			"create table uv (u int primary key, v int);",
			"insert into xy values (0,0),(1,1),(2,2),(3,3);",
			"insert into uv values (0,3),(3,0),(2,1),(1,2);",
		},
		Assertions: []ScriptTestAssertion{
			{
				// https://github.com/dolthub/dolt/issues/4344
				Query:    "select x as v, (select u from uv where v = y) as u from xy;",
				Expected: []sql.Row{{0, 3}, {1, 2}, {2, 1}, {3, 0}},
			},
			{
				// GMS currently returns {0, 0, 0} The second alias seems to get overwritten.
				// https://github.com/dolthub/go-mysql-server/issues/1286
				Skip: true,

				// When multiple aliases are defined with the same name, a subquery prefers the first definition
				Query:    "select 0 as a, 1 as a, (SELECT x from xy where x = a);",
				Expected: []sql.Row{{0, 1, 0}},
			},
			{
				Query:    "SELECT 1 as a, (select a) as a;",
				Expected: []sql.Row{{1, 1}},
			},
			{
				Query:    "SELECT 1 as a, (select a) as b;",
				Expected: []sql.Row{{1, 1}},
			},
			{
				Query:    "SELECT 1 as a, (select a) as b from dual;",
				Expected: []sql.Row{{1, 1}},
			},
			{
				Query:    "SELECT 1 as a, (select a) as b from xy;",
				Expected: []sql.Row{{1, 1}, {1, 1}, {1, 1}, {1, 1}},
			},
			{
				Query:    "select x, (select 1) as y from xy;",
				Expected: []sql.Row{{0, 1}, {1, 1}, {2, 1}, {3, 1}},
			},
			{
				Query:    "SELECT 1 as a, (select a) from xy;",
				Expected: []sql.Row{{1, 1}, {1, 1}, {1, 1}, {1, 1}},
			},
		},
	},
	{
		Name: "column aliases in three scopes",
		SetUpScript: []string{
			"create table xy (x int primary key, y int);",
			"create table uv (u int primary key, v int);",
			"insert into xy values (0,0),(1,1),(2,2),(3,3);",
			"insert into uv values (0,3),(3,0),(2,1),(1,2);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select x, (select 1) as y, (select (select y as q)) as z from (select * from xy) as xy;",
				Expected: []sql.Row{{0, 1, 0}, {1, 1, 1}, {2, 1, 2}, {3, 1, 3}},
			},
		},
	},
	{
		Name: "various broken alias queries",
		Assertions: []ScriptTestAssertion{
			{
				// The dual table's schema collides with this alias name
				// https://github.com/dolthub/dolt/issues/4256
				Skip:     true,
				Query:    `select "foo" as dummy, (select dummy)`,
				Expected: []sql.Row{{"foo", "foo"}},
			},
			{
				// The second query in the union subquery returns "x" instead of mytable.i
				// https://github.com/dolthub/dolt/issues/4256
				Skip:     true,
				Query:    `SELECT *, (select i union select i) as a from mytable;`,
				Expected: []sql.Row{{1, "first row", 1}, {2, "second row", 2}, {3, "third row", 3}},
			},
			{
				// Fails with an unresolved *plan.Project node error
				// The second Project in the union subquery doens't seem to get its alias reference resolved
				// TODO: Something with opaque nodes not being processed? the qualifyExpresions code need to run here? Or resolveColumns?
				Skip:     true,
				Query:    `SELECT 1 as a, (select a union select a) as b;`,
				Expected: []sql.Row{{1, 1}},
			},
			{
				// GMS executes this query, but it is not valid because of the forward ref of alias b.
				// GMS should return an error about an invalid forward-ref.
				Skip:        true,
				Query:       `select 1 as a, (select b), 0 as b;`,
				ExpectedErr: sql.ErrColumnNotFound,
			},
			{
				// GMS returns the error "found HAVING clause with no GROUP BY", but MySQL executes
				// this query without any problems.
				// https://github.com/dolthub/go-mysql-server/issues/1289
				Skip:     true,
				Query:    "select t1.i as a from mytable as t1 having a = t1.i;",
				Expected: []sql.Row{{1}, {2}, {3}},
			},
			{
				// GMS returns "expression 'dt.two' doesn't appear in the group by expressions", but MySQL will execute
				// this query.
				Skip:     true,
				Query:    "select 1 as a, one + 1 as mod1, dt.* from mytable as t1, (select 1, 2 from mytable) as dt (one, two) where dt.one > 0 group by one;",
				Expected: []sql.Row{{1}},
			},
		},
	},
	{
		// TODO: This isn't specific to column aliases, so this might not be the best place for these tests, but getting them started.
		// TODO: We should also include the NTC example query that errors out, too
		// Include that as a separate/third ScriptTest, to keep the SetupScripts separate
		// Use a separate/second ScriptTest for error cases
		Name: "outer scope visibility for derived tables",
		SetUpScript: []string{
			"create table t1 (a int primary key, b int, c int, d int, e int);",
			"create table t2 (a int primary key, b int, c int, d int, e int);",
			"insert into t1 values (1, 1, 1, 100, 100), (2, 2, 2, 200, 200);",
			"insert into t2 values (2, 2, 2, 2, 2);",
			"create table numbers (val int);",
			"insert into numbers values (1), (1), (2), (3), (3), (3), (4), (5), (6), (6), (6);",
		},
		Assertions: []ScriptTestAssertion{
			{
				// A subquery containing a derived table, used in the WHERE clause of a top-level query, has visibility
				// to tables and columns in the top-level query.
				Query:    "SELECT * FROM t1 WHERE t1.d > (SELECT dt.a FROM (SELECT t2.a AS a FROM t2 WHERE t2.b = t1.b) dt);",
				Expected: []sql.Row{{2, 2, 2, 200, 200}},
			},
			{
				// A subquery containing a derived table, used in the HAVING clause of a top-level query, has visibility
				// to tables and columns in the top-level query.
				Query:    "SELECT * FROM t1 HAVING t1.d > (SELECT dt.a FROM (SELECT t2.a AS a FROM t2 WHERE t2.b = t1.b) dt);",
				Expected: []sql.Row{{2, 2, 2, 200, 200}},
			},
			{
				// TODO: Remove this assertion after finished debugging/testing
				// Interesting! This query works and correctly returns null... but... when it's executed as a subquery,
				// it somehow returns 1 for the same data?
				Query:    "SELECT max(dt.a) FROM (SELECT t2.a AS a FROM t2 WHERE t2.b = 1) dt;",
				Expected: []sql.Row{{nil}},
			},
			{
				// TODO: Testing a simpler query with the missing NULL repro
				//       flattenAggregationExprs is setting this GetField index incorrectly!!
				//       Running this without an aggregation function would probably make it pass? YUP!!!
				//Query:    "SELECT (SELECT dt.z FROM (SELECT t2.a AS z FROM t2 WHERE t2.b = t1.b) dt) FROM t1;",
				Query:    "SELECT (SELECT max(dt.z) FROM (SELECT t2.a AS z FROM t2 WHERE t2.b = t1.b) dt) FROM t1;",
				Expected: []sql.Row{{nil}, {2}},
			},
			{
				// A subquery containing a derived table, projected in a SELECT query, has visibility to tables and columns
				// in the top-level query.
				// TODO: Does it have visibility to alias expressions, too? or just tables/columns?
				// TODO: Was failing with: unable to find field with index 6 in row of 2 columns
				//       Updated to prepend the row with the additional columns from the Child results, but now failing
				//       with unexpected results: (1 instead of nil in the first record)
				//       []sql.Row{{1, 1, 1, 100, 100, 1}, {2, 2, 2, 200, 200, 2}},
				Query:    "SELECT t1.*, (SELECT max(dt.a) FROM (SELECT t2.a AS a FROM t2 WHERE t2.b = t1.b) dt) FROM t1;",
				Expected: []sql.Row{{1, 1, 1, 100, 100, nil}, {2, 2, 2, 200, 200, 2}},
			},
			{
				// A subquery containing a derived table, projected in a GROUPBY query, has visibility to tables and columns
				// in the top-level query.
				// TODO: Does it have visibility to alias expressions, too? or just tables/columns?
				// TODO: Currently failing with: expression 't1.b' doesn't appear in the group by expressions
				//       Seems like the root of this error is really a separate GroupBy issue where Dolt/GMS is more
				//       restrictive than MySQL.
				// TODO: Can we repro this in another scenario, cut an issue, and get Jennifer to tackle it?
				// TODO: Now we're getting this query to run, but we're still seeing the same problem as above where
				//       we don't get the expected NULL value and instead get "1".
				// https://github.com/dolthub/dolt/issues/1448
				Query: "SELECT t1.a, t1.b, (SELECT max(dt.a) FROM (SELECT t2.a AS a FROM t2 WHERE t2.b = t1.b) dt) FROM t1 GROUP BY 1, 2, 3;",
				//Expected: []sql.Row{{1, 1, 1, 100, 100, nil}, {2, 2, 2, 200, 200, 2}},
				Expected: []sql.Row{{1, 1, nil}, {2, 2, 2}},
			},
			{
				// A subquery containing a derived table, projected in a WINDOW query, has visibility to tables and columns
				// in the top-level query.
				Query:    "SELECT val, row_number() over (partition by val) as 'row_number', (SELECT two from (SELECT val*2, val*3) as dt(one, two)) as a1 from numbers having a1 > 10;",
				Expected: []sql.Row{{4, 1, 12}, {5, 1, 15}, {6, 1, 18}, {6, 2, 18}, {6, 3, 18}},
			},
			{
				// A subquery containing a derived table, used in the WINDOW clause of a top-level query, has visibility
				// to tables and columns in the top-level query.
				Query:    "SELECT val, row_number() over (partition by val) as 'row_number', (SELECT two from (SELECT val*2, val*3) as dt(one, two)) as a1 from numbers having a1 > 10;",
				Expected: []sql.Row{{4, 1, 12}, {5, 1, 15}, {6, 1, 18}, {6, 2, 18}, {6, 3, 18}},
			},
			{
				// A subquery containing a derived table, used in the GROUP BY clause of a top-level query as a grouping
				// expression, has visibility to tables and columns in the top-level query.
				Query:    "SELECT max(val), (select max(dt.a) from (SELECT val as a) as dt(a)) as a1 from numbers group by a1;",
				Expected: []sql.Row{{1, 1}, {2, 2}, {3, 3}, {4, 4}, {5, 5}, {6, 6}},
			},
			{
				// Error Tests...
				// Currently returns: found HAVING clause with no GROUP BY
				// https://github.com/dolthub/go-mysql-server/issues/1289
				Skip: true,

				// A derived table inside a derived table does not have visibility to outer scopes.
				Query:       "SELECT 1 as a1, dt.* from (select * from (select * from numbers having val = a1) as dt2(val)) as dt(val);",
				ExpectedErr: sql.ErrUnknownColumn,
			},
		},
	},
}
