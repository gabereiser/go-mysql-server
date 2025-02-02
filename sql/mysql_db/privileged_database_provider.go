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

package mysql_db

import (
	"strings"

	"github.com/gabereiser/go-mysql-server/sql"
)

// PrivilegedDatabaseProvider is a wrapper around a normal sql.DatabaseProvider that takes a context's client's
// privileges into consideration when returning a sql.Database. In addition, any returned databases are wrapped with
// PrivilegedDatabase.
type PrivilegedDatabaseProvider struct {
	grantTables *MySQLDb
	provider    sql.DatabaseProvider
}

var _ sql.DatabaseProvider = PrivilegedDatabaseProvider{}

// NewPrivilegedDatabaseProvider returns a new PrivilegedDatabaseProvider. As a sql.DatabaseProvider may be added to an
// analyzer when Grant Tables are disabled (and Grant Tables may be enabled or disabled at any time), a new
// PrivilegedDatabaseProvider is returned whenever the sql.DatabaseProvider is needed (as long as Grant Tables are
// enabled) rather than wrapping a sql.DatabaseProvider when it is provided to the analyzer.
func NewPrivilegedDatabaseProvider(grantTables *MySQLDb, p sql.DatabaseProvider) sql.DatabaseProvider {
	return PrivilegedDatabaseProvider{
		grantTables: grantTables,
		provider:    p,
	}
}

// Database implements the interface sql.DatabaseProvider.
func (pdp PrivilegedDatabaseProvider) Database(ctx *sql.Context, name string) (sql.Database, error) {

	if lowName := strings.ToLower(name); lowName != sql.InformationSchemaDatabaseName {
		privSet := pdp.grantTables.UserActivePrivilegeSet(ctx)
		// If the user has no global static privileges or database-relevant privileges then the database is not accessible.
		if privSet.Count() == 0 && !privSet.Database(name).HasPrivileges() {
			return nil, sql.ErrDatabaseAccessDeniedForUser.New(pdp.usernameFromCtx(ctx), name)
		}
		if lowName == "mysql" {
			return pdp.grantTables, nil
		}
	}
	db, err := pdp.provider.Database(ctx, name)
	if err != nil {
		return nil, err
	}
	return NewPrivilegedDatabase(pdp.grantTables, db), nil
}

// HasDatabase implements the interface sql.DatabaseProvider.
func (pdp PrivilegedDatabaseProvider) HasDatabase(ctx *sql.Context, name string) bool {
	if name != sql.InformationSchemaDatabaseName {
		privSet := pdp.grantTables.UserActivePrivilegeSet(ctx)
		// If the user has no global static privileges or database-relevant privileges then the database is not accessible.
		if privSet.Count() == 0 && !privSet.Database(name).HasPrivileges() {
			return false
		}
	}
	return pdp.provider.HasDatabase(ctx, name)
}

// AllDatabases implements the interface sql.DatabaseProvider.
func (pdp PrivilegedDatabaseProvider) AllDatabases(ctx *sql.Context) []sql.Database {
	privilegeSet := pdp.grantTables.UserActivePrivilegeSet(ctx)
	privilegeSetCount := privilegeSet.Count()

	var databasesWithAccess []sql.Database
	allDatabases := pdp.provider.AllDatabases(ctx)
	for _, db := range allDatabases {
		// If the user has any global static privileges or database-relevant privileges then the database is accessible.
		// 'information_schema' database is always accessible.
		if db.Name() == sql.InformationSchemaDatabaseName || privilegeSetCount > 0 || privilegeSet.Database(db.Name()).HasPrivileges() {
			databasesWithAccess = append(databasesWithAccess, NewPrivilegedDatabase(pdp.grantTables, db))
		}
	}
	return databasesWithAccess
}

// usernameFromCtx returns the username from the context, properly formatted for returned errors.
func (pdp PrivilegedDatabaseProvider) usernameFromCtx(ctx *sql.Context) string {
	client := ctx.Session.Client()
	return User{User: client.User, Host: client.Address}.UserHostToString("'")
}

// PrivilegedDatabase is a wrapper around a normal sql.Database that takes a context's client's privileges into
// consideration when returning a sql.Table.
type PrivilegedDatabase struct {
	grantTables *MySQLDb
	db          sql.Database
	//TODO: this should also handle views as the relevant privilege exists
}

var _ sql.Database = PrivilegedDatabase{}
var _ sql.VersionedDatabase = PrivilegedDatabase{}
var _ sql.TableCreator = PrivilegedDatabase{}
var _ sql.TableDropper = PrivilegedDatabase{}
var _ sql.TableRenamer = PrivilegedDatabase{}
var _ sql.TriggerDatabase = PrivilegedDatabase{}
var _ sql.StoredProcedureDatabase = PrivilegedDatabase{}
var _ sql.TableCopierDatabase = PrivilegedDatabase{}
var _ sql.ReadOnlyDatabase = PrivilegedDatabase{}
var _ sql.TemporaryTableDatabase = PrivilegedDatabase{}
var _ sql.CollatedDatabase = PrivilegedDatabase{}

// NewPrivilegedDatabase returns a new PrivilegedDatabase.
func NewPrivilegedDatabase(grantTables *MySQLDb, db sql.Database) sql.Database {
	return PrivilegedDatabase{
		grantTables: grantTables,
		db:          db,
	}
}

// Name implements the interface sql.Database.
func (pdb PrivilegedDatabase) Name() string {
	return pdb.db.Name()
}

// GetTableInsensitive implements the interface sql.Database.
func (pdb PrivilegedDatabase) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	if pdb.db.Name() != sql.InformationSchemaDatabaseName {
		privSet := pdb.grantTables.UserActivePrivilegeSet(ctx)
		dbSet := privSet.Database(pdb.db.Name())
		// If there are no usable privileges for this database then the table is inaccessible.
		if privSet.Count() == 0 && !dbSet.HasPrivileges() {
			return nil, false, sql.ErrDatabaseAccessDeniedForUser.New(pdb.usernameFromCtx(ctx), pdb.db.Name())
		}

		tblSet := dbSet.Table(tblName)
		// If the user has no global static privileges, database-level privileges, or table-relevant privileges then the
		// table is not accessible.
		if privSet.Count() == 0 && dbSet.Count() == 0 && !tblSet.HasPrivileges() {
			return nil, false, sql.ErrTableAccessDeniedForUser.New(pdb.usernameFromCtx(ctx), tblName)
		}
	}
	return pdb.db.GetTableInsensitive(ctx, tblName)
}

// GetTableNames implements the interface sql.Database.
func (pdb PrivilegedDatabase) GetTableNames(ctx *sql.Context) ([]string, error) {
	var tablesWithAccess []string
	var err error
	if pdb.db.Name() != sql.InformationSchemaDatabaseName {
		privSet := pdb.grantTables.UserActivePrivilegeSet(ctx)
		dbSet := privSet.Database(pdb.db.Name())
		// If there are no usable privileges for this database then no table is accessible.
		privSetCount := privSet.Count()
		if privSetCount == 0 && !dbSet.HasPrivileges() {
			return nil, nil
		}

		tblNames, err := pdb.db.GetTableNames(ctx)
		if err != nil {
			return nil, err
		}
		dbSetCount := dbSet.Count()
		for _, tblName := range tblNames {
			// If the user has any global static privileges, database-level privileges, or table-relevant privileges then a
			// table is accessible.
			if privSetCount > 0 || dbSetCount > 0 || dbSet.Table(tblName).HasPrivileges() {
				tablesWithAccess = append(tablesWithAccess, tblName)
			}
		}
	} else {
		tablesWithAccess, err = pdb.db.GetTableNames(ctx)
		if err != nil {
			return nil, err
		}
	}
	return tablesWithAccess, nil
}

// GetTableInsensitiveAsOf returns a new sql.VersionedDatabase.
func (pdb PrivilegedDatabase) GetTableInsensitiveAsOf(ctx *sql.Context, tblName string, asOf interface{}) (sql.Table, bool, error) {
	db, ok := pdb.db.(sql.VersionedDatabase)
	if !ok {
		return nil, false, sql.ErrAsOfNotSupported.New(pdb.db.Name())
	}

	if pdb.db.Name() != sql.InformationSchemaDatabaseName {
		privSet := pdb.grantTables.UserActivePrivilegeSet(ctx)
		dbSet := privSet.Database(pdb.db.Name())
		// If there are no usable privileges for this database then the table is inaccessible.
		if privSet.Count() == 0 && !dbSet.HasPrivileges() {
			return nil, false, sql.ErrDatabaseAccessDeniedForUser.New(pdb.usernameFromCtx(ctx), pdb.db.Name())
		}

		tblSet := dbSet.Table(tblName)
		// If the user has no global static privileges, database-level privileges, or table-relevant privileges then the
		// table is not accessible.
		if privSet.Count() == 0 && dbSet.Count() == 0 && !tblSet.HasPrivileges() {
			return nil, false, sql.ErrTableAccessDeniedForUser.New(pdb.usernameFromCtx(ctx), tblName)
		}
	}
	return db.GetTableInsensitiveAsOf(ctx, tblName, asOf)
}

// GetTableNamesAsOf returns a new sql.VersionedDatabase.
func (pdb PrivilegedDatabase) GetTableNamesAsOf(ctx *sql.Context, asOf interface{}) ([]string, error) {
	db, ok := pdb.db.(sql.VersionedDatabase)
	if !ok {
		return nil, nil
	}

	var tablesWithAccess []string
	var err error
	if pdb.db.Name() != sql.InformationSchemaDatabaseName {
		privSet := pdb.grantTables.UserActivePrivilegeSet(ctx)
		dbSet := privSet.Database(pdb.db.Name())
		// If there are no usable privileges for this database then no table is accessible.
		if privSet.Count() == 0 && !dbSet.HasPrivileges() {
			return nil, nil
		}

		tblNames, err := db.GetTableNamesAsOf(ctx, asOf)
		if err != nil {
			return nil, err
		}
		privSetCount := privSet.Count()
		dbSetCount := dbSet.Count()
		for _, tblName := range tblNames {
			// If the user has any global static privileges, database-level privileges, or table-relevant privileges then a
			// table is accessible.
			if privSetCount > 0 || dbSetCount > 0 && dbSet.Table(tblName).HasPrivileges() {
				tablesWithAccess = append(tablesWithAccess, tblName)
			}
		}
	} else {
		tablesWithAccess, err = db.GetTableNamesAsOf(ctx, asOf)
		if err != nil {
			return nil, err
		}
	}

	return tablesWithAccess, nil
}

// CreateTable implements the interface sql.TableCreator.
func (pdb PrivilegedDatabase) CreateTable(ctx *sql.Context, name string, schema sql.PrimaryKeySchema, collation sql.CollationID) error {
	if db, ok := pdb.db.(sql.TableCreator); ok {
		return db.CreateTable(ctx, name, schema, collation)
	}
	return sql.ErrCreateTableNotSupported.New(pdb.db.Name())
}

// DropTable implements the interface sql.TableDropper.
func (pdb PrivilegedDatabase) DropTable(ctx *sql.Context, name string) error {
	if db, ok := pdb.db.(sql.TableDropper); ok {
		return db.DropTable(ctx, name)
	}
	return sql.ErrDropTableNotSupported.New(pdb.db.Name())
}

// RenameTable implements the interface sql.TableRenamer.
func (pdb PrivilegedDatabase) RenameTable(ctx *sql.Context, oldName, newName string) error {
	if db, ok := pdb.db.(sql.TableRenamer); ok {
		return db.RenameTable(ctx, oldName, newName)
	}
	return sql.ErrRenameTableNotSupported.New(pdb.db.Name())
}

// GetTriggers implements the interface sql.TriggerDatabase.
func (pdb PrivilegedDatabase) GetTriggers(ctx *sql.Context) ([]sql.TriggerDefinition, error) {
	if pdb.db.Name() == "information_schema" {
		return nil, nil
	}
	if db, ok := pdb.db.(sql.TriggerDatabase); ok {
		return db.GetTriggers(ctx)
	}
	return nil, sql.ErrTriggersNotSupported.New(pdb.db.Name())
}

// CreateTrigger implements the interface sql.TriggerDatabase.
func (pdb PrivilegedDatabase) CreateTrigger(ctx *sql.Context, definition sql.TriggerDefinition) error {
	if db, ok := pdb.db.(sql.TriggerDatabase); ok {
		return db.CreateTrigger(ctx, definition)
	}
	return sql.ErrTriggersNotSupported.New(pdb.db.Name())
}

// DropTrigger implements the interface sql.TriggerDatabase.
func (pdb PrivilegedDatabase) DropTrigger(ctx *sql.Context, name string) error {
	if db, ok := pdb.db.(sql.TriggerDatabase); ok {
		return db.DropTrigger(ctx, name)
	}
	return sql.ErrTriggersNotSupported.New(pdb.db.Name())
}

// GetStoredProcedure implements the interface sql.StoredProcedureDatabase.
func (pdb PrivilegedDatabase) GetStoredProcedure(ctx *sql.Context, name string) (sql.StoredProcedureDetails, bool, error) {
	if pdb.db.Name() == "information_schema" {
		return sql.StoredProcedureDetails{}, false, nil
	}
	if db, ok := pdb.db.(sql.StoredProcedureDatabase); ok {
		return db.GetStoredProcedure(ctx, name)
	}
	return sql.StoredProcedureDetails{}, false, sql.ErrStoredProceduresNotSupported.New(pdb.db.Name())
}

// GetStoredProcedures implements the interface sql.StoredProcedureDatabase.
func (pdb PrivilegedDatabase) GetStoredProcedures(ctx *sql.Context) ([]sql.StoredProcedureDetails, error) {
	if pdb.db.Name() == "information_schema" {
		return nil, nil
	}
	if db, ok := pdb.db.(sql.StoredProcedureDatabase); ok {
		return db.GetStoredProcedures(ctx)
	}
	return nil, sql.ErrStoredProceduresNotSupported.New(pdb.db.Name())
}

// SaveStoredProcedure implements the interface sql.StoredProcedureDatabase.
func (pdb PrivilegedDatabase) SaveStoredProcedure(ctx *sql.Context, spd sql.StoredProcedureDetails) error {
	if db, ok := pdb.db.(sql.StoredProcedureDatabase); ok {
		return db.SaveStoredProcedure(ctx, spd)
	}
	return sql.ErrStoredProceduresNotSupported.New(pdb.db.Name())
}

// DropStoredProcedure implements the interface sql.StoredProcedureDatabase.
func (pdb PrivilegedDatabase) DropStoredProcedure(ctx *sql.Context, name string) error {
	if db, ok := pdb.db.(sql.StoredProcedureDatabase); ok {
		return db.DropStoredProcedure(ctx, name)
	}
	return sql.ErrStoredProceduresNotSupported.New(pdb.db.Name())
}

// CopyTableData implements the interface sql.TableCopierDatabase.
func (pdb PrivilegedDatabase) CopyTableData(ctx *sql.Context, sourceTable string, destinationTable string) (uint64, error) {
	if db, ok := pdb.db.(sql.TableCopierDatabase); ok {
		// Privilege checking is handled in the analyzer
		return db.CopyTableData(ctx, sourceTable, destinationTable)
	}
	return 0, sql.ErrTableCopyingNotSupported.New()
}

// IsReadOnly implements the interface sql.ReadOnlyDatabase.
func (pdb PrivilegedDatabase) IsReadOnly() bool {
	if db, ok := pdb.db.(sql.ReadOnlyDatabase); ok {
		return db.IsReadOnly()
	}
	return false
}

// GetAllTemporaryTables implements the interface sql.TemporaryTableDatabase.
func (pdb PrivilegedDatabase) GetAllTemporaryTables(ctx *sql.Context) ([]sql.Table, error) {
	if db, ok := pdb.db.(sql.TemporaryTableDatabase); ok {
		return db.GetAllTemporaryTables(ctx)
	}
	// All current temp table checks skip if not implemented, same is iterating over an empty slice
	return nil, nil
}

// GetCollation implements the interface sql.CollatedDatabase.
func (pdb PrivilegedDatabase) GetCollation(ctx *sql.Context) sql.CollationID {
	if db, ok := pdb.db.(sql.CollatedDatabase); ok {
		return db.GetCollation(ctx)
	}
	return sql.Collation_Default
}

// SetCollation implements the interface sql.CollatedDatabase.
func (pdb PrivilegedDatabase) SetCollation(ctx *sql.Context, collation sql.CollationID) error {
	if db, ok := pdb.db.(sql.CollatedDatabase); ok {
		return db.SetCollation(ctx, collation)
	}
	return sql.ErrDatabaseCollationsNotSupported.New(pdb.db.Name())
}

// Unwrap returns the wrapped sql.Database.
func (pdb PrivilegedDatabase) Unwrap() sql.Database {
	return pdb.db
}

// usernameFromCtx returns the username from the context, properly formatted for returned errors.
func (pdb PrivilegedDatabase) usernameFromCtx(ctx *sql.Context) string {
	client := ctx.Session.Client()
	return User{User: client.User, Host: client.Address}.UserHostToString("'")
}
