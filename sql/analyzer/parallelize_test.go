// Copyright 2020-2021 Dolthub, Inc.
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

package analyzer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gabereiser/go-mysql-server/memory"
	"github.com/gabereiser/go-mysql-server/sql"
	"github.com/gabereiser/go-mysql-server/sql/expression"
	"github.com/gabereiser/go-mysql-server/sql/expression/function/aggregation/window"
	"github.com/gabereiser/go-mysql-server/sql/plan"
	"github.com/gabereiser/go-mysql-server/sql/transform"
	"github.com/gabereiser/go-mysql-server/sql/types"
)

func TestParallelize(t *testing.T) {
	require := require.New(t)
	table := memory.NewTable("t", sql.PrimaryKeySchema{}, nil)
	rule := getRuleFrom(OnceAfterAll, parallelizeId)
	node := plan.NewProject(
		nil,
		plan.NewInnerJoin(
			plan.NewFilter(
				expression.NewLiteral(1, types.Int64),
				plan.NewResolvedTable(table, nil, nil),
			),
			plan.NewFilter(
				expression.NewLiteral(1, types.Int64),
				plan.NewResolvedTable(table, nil, nil),
			),
			expression.NewLiteral(1, types.Int64),
		),
	)

	expected := plan.NewProject(
		nil,
		plan.NewInnerJoin(
			plan.NewExchange(
				2,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			plan.NewExchange(
				2,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			expression.NewLiteral(1, types.Int64),
		),
	)

	result, _, err := rule.Apply(sql.NewEmptyContext(), &Analyzer{Parallelism: 2}, node, nil, DefaultRuleSelector)
	require.NoError(err)
	require.Equal(expected, result)
}

func TestParallelizeCreateIndex(t *testing.T) {
	require := require.New(t)
	table := memory.NewTable("t", sql.PrimaryKeySchema{}, nil)
	rule := getRuleFrom(OnceAfterAll, parallelizeId)
	node := plan.NewCreateIndex(
		"",
		plan.NewResolvedTable(table, nil, nil),
		nil,
		"",
		nil,
	)

	result, _, err := rule.Apply(sql.NewEmptyContext(), &Analyzer{Parallelism: 1}, node, nil, DefaultRuleSelector)
	require.NoError(err)
	require.Equal(node, result)
}

func TestIsParallelizable(t *testing.T) {
	table := memory.NewTable("t", sql.PrimaryKeySchema{}, nil)

	testCases := []struct {
		name           string
		node           sql.Node
		parallelizable bool
	}{
		{
			"just table",
			plan.NewResolvedTable(table, nil, nil),
			true,
		},
		{
			"filter",
			plan.NewFilter(
				expression.NewLiteral(1, types.Int64),
				plan.NewResolvedTable(table, nil, nil),
			),
			true,
		},
		{
			"filter with a subquery",
			plan.NewFilter(
				eq(
					lit(1),
					plan.NewSubquery(
						plan.NewProject([]sql.Expression{lit(1)}, plan.NewResolvedTable(table, nil, nil)), "select 1 from table")),
				plan.NewResolvedTable(table, nil, nil),
			),
			true,
		},
		{
			"filter with an incompatible subquery",
			plan.NewFilter(
				eq(
					lit(1),
					plan.NewSubquery(
						plan.NewProject([]sql.Expression{gf(0, "", "row_number()")},
							plan.NewWindow([]sql.Expression{window.NewRowNumber()}, plan.NewResolvedTable(table, nil, nil)),
						),
						"select row_number over () from table",
					),
				),
				plan.NewResolvedTable(table, nil, nil),
			),
			false,
		},
		{
			"project",
			plan.NewProject(
				nil,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			true,
		},
		{
			"project with a subquery",
			plan.NewProject([]sql.Expression{
				plan.NewSubquery(
					plan.NewProject([]sql.Expression{lit(1)}, plan.NewResolvedTable(table, nil, nil)), "select 1 from table"),
			},
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			true,
		},
		{
			"project with an incompatible subquery",
			plan.NewProject([]sql.Expression{
				plan.NewSubquery(
					plan.NewProject([]sql.Expression{gf(0, "", "row_number()")},
						plan.NewWindow([]sql.Expression{window.NewRowNumber()}, plan.NewResolvedTable(table, nil, nil)),
					),
					"select row_number over () from table",
				),
			},
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			false,
		},
		{
			"join",
			plan.NewInnerJoin(
				plan.NewResolvedTable(table, nil, nil),
				plan.NewResolvedTable(table, nil, nil),
				expression.NewLiteral(1, types.Int64),
			),
			false,
		},
		{
			"group by",
			plan.NewGroupBy(
				nil,
				nil,
				plan.NewResolvedTable(nil, nil, nil),
			),
			false,
		},
		{
			"limit",
			plan.NewLimit(
				expression.NewLiteral(5, types.Int8),
				plan.NewResolvedTable(nil, nil, nil),
			),
			false,
		},
		{
			"offset",
			plan.NewOffset(
				expression.NewLiteral(5, types.Int8),
				plan.NewResolvedTable(nil, nil, nil),
			),
			false,
		},
		{
			"sort",
			plan.NewSort(
				nil,
				plan.NewResolvedTable(nil, nil, nil),
			),
			false,
		},
		{
			"distinct",
			plan.NewDistinct(
				plan.NewResolvedTable(nil, nil, nil),
			),
			false,
		},
		{
			"ordered distinct",
			plan.NewOrderedDistinct(
				plan.NewResolvedTable(nil, nil, nil),
			),
			false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.parallelizable, isParallelizable(tt.node))
		})
	}
}

func TestRemoveRedundantExchanges(t *testing.T) {
	require := require.New(t)

	table := memory.NewTable("t", sql.PrimaryKeySchema{}, nil)

	node := plan.NewProject(
		nil,
		plan.NewInnerJoin(
			plan.NewExchange(
				1,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewExchange(
						1,
						plan.NewResolvedTable(table, nil, nil),
					),
				),
			),
			plan.NewExchange(
				1,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewExchange(
						1,
						plan.NewResolvedTable(table, nil, nil),
					),
				),
			),
			expression.NewLiteral(1, types.Int64),
		),
	)

	expected := plan.NewProject(
		nil,
		plan.NewInnerJoin(
			plan.NewExchange(
				1,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			plan.NewExchange(
				1,
				plan.NewFilter(
					expression.NewLiteral(1, types.Int64),
					plan.NewResolvedTable(table, nil, nil),
				),
			),
			expression.NewLiteral(1, types.Int64),
		),
	)

	result, _, err := transform.Node(node, removeRedundantExchanges)
	require.NoError(err)
	require.Equal(expected, result)
}
