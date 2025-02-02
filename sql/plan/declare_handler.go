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

package plan

import (
	"fmt"
	"io"

	"github.com/gabereiser/go-mysql-server/sql"
	"github.com/gabereiser/go-mysql-server/sql/expression"
)

type DeclareHandlerAction byte

const (
	DeclareHandlerAction_Continue DeclareHandlerAction = iota
	DeclareHandlerAction_Exit
	DeclareHandlerAction_Undo
)

// DeclareHandler represents the DECLARE ... HANDLER statement.
type DeclareHandler struct {
	Action    DeclareHandlerAction
	Statement sql.Node
	pRef      *expression.ProcedureReference
	//TODO: implement other conditions besides NOT FOUND
}

var _ sql.Node = (*DeclareHandler)(nil)
var _ sql.DebugStringer = (*DeclareHandler)(nil)
var _ expression.ProcedureReferencable = (*DeclareHandler)(nil)

// NewDeclareHandler returns a new *DeclareHandler node.
func NewDeclareHandler(action DeclareHandlerAction, statement sql.Node) (*DeclareHandler, error) {
	if action == DeclareHandlerAction_Undo {
		return nil, sql.ErrDeclareHandlerUndo.New()
	}
	return &DeclareHandler{
		Action:    action,
		Statement: statement,
	}, nil
}

// Resolved implements the interface sql.Node.
func (d *DeclareHandler) Resolved() bool {
	return true
}

// String implements the interface sql.Node.
func (d *DeclareHandler) String() string {
	var action string
	switch d.Action {
	case DeclareHandlerAction_Continue:
		action = "CONTINUE"
	case DeclareHandlerAction_Exit:
		action = "EXIT"
	case DeclareHandlerAction_Undo:
		action = "UNDO"
	}
	return fmt.Sprintf("DECLARE %s HANDLER FOR NOT FOUND %s", action, d.Statement.String())
}

// DebugString implements the interface sql.DebugStringer.
func (d *DeclareHandler) DebugString() string {
	var action string
	switch d.Action {
	case DeclareHandlerAction_Continue:
		action = "CONTINUE"
	case DeclareHandlerAction_Exit:
		action = "EXIT"
	case DeclareHandlerAction_Undo:
		action = "UNDO"
	}
	return fmt.Sprintf("DECLARE %s HANDLER FOR NOT FOUND %s", action, sql.DebugString(d.Statement))
}

// Schema implements the interface sql.Node.
func (d *DeclareHandler) Schema() sql.Schema {
	return nil
}

// Children implements the interface sql.Node.
func (d *DeclareHandler) Children() []sql.Node {
	return []sql.Node{d.Statement}
}

// WithChildren implements the interface sql.Node.
func (d *DeclareHandler) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(d, len(children), 1)
	}

	nd := *d
	nd.Statement = children[0]
	return &nd, nil
}

// CheckPrivileges implements the interface sql.Node.
func (d *DeclareHandler) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

// RowIter implements the interface sql.Node.
func (d *DeclareHandler) RowIter(ctx *sql.Context, row sql.Row) (sql.RowIter, error) {
	return &declareHandlerIter{d}, nil
}

// WithParamReference implements the interface expression.ProcedureReferencable.
func (d *DeclareHandler) WithParamReference(pRef *expression.ProcedureReference) sql.Node {
	nd := *d
	nd.pRef = pRef
	return &nd
}

// declareHandlerIter is the sql.RowIter of *DeclareHandler.
type declareHandlerIter struct {
	*DeclareHandler
}

var _ sql.RowIter = (*declareHandlerIter)(nil)

// Next implements the interface sql.RowIter.
func (d *declareHandlerIter) Next(ctx *sql.Context) (sql.Row, error) {
	d.pRef.InitializeHandler(d.Statement, d.Action == DeclareHandlerAction_Exit)
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (d *declareHandlerIter) Close(ctx *sql.Context) error {
	return nil
}
