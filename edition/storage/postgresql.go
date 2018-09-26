// Copyright 2016 Documize Inc. <legal@documize.com>. All rights reserved.
//
// This software (Documize Community Edition) is licensed under
// GNU AGPL v3 http://www.gnu.org/licenses/agpl-3.0.en.html
//
// You can operate outside the AGPL restrictions by purchasing
// Documize Enterprise Edition and obtaining a commercial license
// by contacting <sales@documize.com>.
//
// https://documize.com

// Package storage sets up database persistence providers.
package storage

import (
	"fmt"
	"strings"

	"github.com/documize/community/core/env"
	"github.com/documize/community/domain"
	account "github.com/documize/community/domain/account"
	activity "github.com/documize/community/domain/activity"
	attachment "github.com/documize/community/domain/attachment"
	audit "github.com/documize/community/domain/audit"
	block "github.com/documize/community/domain/block"
	category "github.com/documize/community/domain/category"
	document "github.com/documize/community/domain/document"
	group "github.com/documize/community/domain/group"
	link "github.com/documize/community/domain/link"
	meta "github.com/documize/community/domain/meta"
	org "github.com/documize/community/domain/organization"
	page "github.com/documize/community/domain/page"
	permission "github.com/documize/community/domain/permission"
	pin "github.com/documize/community/domain/pin"
	search "github.com/documize/community/domain/search"
	setting "github.com/documize/community/domain/setting"
	space "github.com/documize/community/domain/space"
	user "github.com/documize/community/domain/user"
	_ "github.com/lib/pq" // the mysql driver is required behind the scenes
)

// PostgreSQLProvider supports by popular demand.
type PostgreSQLProvider struct {
	// User specified connection string.
	ConnectionString string

	// Unused for this provider.
	Variant string
}

// SetPostgreSQLProvider creates PostgreSQL provider
func SetPostgreSQLProvider(r *env.Runtime, s *domain.Store) {
	// Set up provider specific details and wire up data prividers.
	r.StoreProvider = PostgreSQLProvider{
		ConnectionString: r.Flags.DBConn,
		Variant:          "",
	}

	// Wire up data providers!
	accountStore := account.Store{}
	accountStore.Runtime = r
	s.Account = accountStore

	// Activity
	activityStore := activity.Store{}
	activityStore.Runtime = r
	s.Activity = activityStore

	// Attachment
	attachmentStore := attachment.Store{}
	attachmentStore.Runtime = r
	s.Attachment = attachmentStore

	// Audit
	auditStore := audit.Store{}
	auditStore.Runtime = r
	s.Audit = auditStore

	// Section Template
	blockStore := block.Store{}
	blockStore.Runtime = r
	s.Block = blockStore

	// Category
	categoryStore := category.Store{}
	categoryStore.Runtime = r
	s.Category = categoryStore

	// Document
	documentStore := document.Store{}
	documentStore.Runtime = r
	s.Document = documentStore

	// Group
	groupStore := group.Store{}
	groupStore.Runtime = r
	s.Group = groupStore

	// Link
	linkStore := link.Store{}
	linkStore.Runtime = r
	s.Link = linkStore

	// Meta
	metaStore := meta.Store{}
	metaStore.Runtime = r
	s.Meta = metaStore

	// Organization (tenant)
	orgStore := org.Store{}
	orgStore.Runtime = r
	s.Organization = orgStore

	// Page (section)
	pageStore := page.Store{}
	pageStore.Runtime = r
	s.Page = pageStore

	// Permission
	permissionStore := permission.Store{}
	permissionStore.Runtime = r
	s.Permission = permissionStore

	// Pin
	pinStore := pin.Store{}
	pinStore.Runtime = r
	s.Pin = pinStore

	// Search
	searchStore := search.Store{}
	searchStore.Runtime = r
	s.Search = searchStore

	// Setting
	settingStore := setting.Store{}
	settingStore.Runtime = r
	s.Setting = settingStore

	// Space
	spaceStore := space.Store{}
	spaceStore.Runtime = r
	s.Space = spaceStore

	// User
	userStore := user.Store{}
	userStore.Runtime = r
	s.User = userStore
}

// Type returns name of provider
func (p PostgreSQLProvider) Type() env.StoreType {
	return env.StoreTypePostgreSQL
}

// TypeVariant returns databse flavor
func (p PostgreSQLProvider) TypeVariant() string {
	return p.Variant
}

// DriverName returns database/sql driver name.
func (p PostgreSQLProvider) DriverName() string {
	return "postgres"
}

// Params returns connection string parameters that must be present before connecting to DB.
func (p PostgreSQLProvider) Params() map[string]string {
	// Not used for this provider.
	return map[string]string{}
}

// Example holds storage provider specific connection string format.
// used in error messages
func (p PostgreSQLProvider) Example() string {
	return "database connection string format is 'host=localhost port=5432 sslmode=disable user=admin password=secret dbname=documize'"
}

// DatabaseName holds the SQL database name where Documize tables live.
func (p PostgreSQLProvider) DatabaseName() string {
	bits := strings.Split(p.ConnectionString, " ")
	for _, s := range bits {
		s = strings.TrimSpace(s)
		if strings.Contains(s, "dbname=") {
			s = strings.Replace(s, "dbname=", "", 1)

			return s
		}
	}

	return ""
}

// MakeConnectionString returns provider specific DB connection string
// complete with default parameters.
func (p PostgreSQLProvider) MakeConnectionString() string {
	// No special processing so return as-is.
	return p.ConnectionString
}

// QueryMeta is how to extract version number, collation, character set from database provider.
func (p PostgreSQLProvider) QueryMeta() string {
	// SELECT version() as vstring, current_setting('server_version_num') as vnumber, pg_encoding_to_char(encoding) AS charset FROM pg_database WHERE datname = 'documize';

	return fmt.Sprintf(`SELECT cast(current_setting('server_version_num') AS TEXT) AS version, version() AS comment, pg_encoding_to_char(encoding) AS charset, '' AS collation
        FROM pg_database WHERE datname = '%s'`, p.DatabaseName())
}

// // QueryStartLock locks database tables.
// func (p PostgreSQLProvider) QueryStartLock() string {
// 	return ""
// }

// // QueryFinishLock unlocks database tables.
// func (p PostgreSQLProvider) QueryFinishLock() string {
// 	return ""
// }

// // QueryInsertProcessID returns database specific query that will
// // insert ID of this running process.
// func (p PostgreSQLProvider) QueryInsertProcessID() string {
// 	return ""
// }

// // QueryDeleteProcessID returns database specific query that will
// // delete ID of this running process.
// func (p PostgreSQLProvider) QueryDeleteProcessID() string {
// 	return ""
// }

// QueryRecordVersionUpgrade returns database specific insert statement
// that records the database version number.
func (p PostgreSQLProvider) QueryRecordVersionUpgrade(version int) string {
	// Make record that holds new database version number.
	json := fmt.Sprintf("{\"database\": \"%d\"}", version)

	return fmt.Sprintf(`INSERT INTO dmz_config (c_key,c_config) VALUES ('META','%s')
        ON CONFLICT (c_key) DO UPDATE SET c_config='%s' WHERE dmz_config.c_key='META'`, json, json)
}

// QueryRecordVersionUpgradeLegacy returns database specific insert statement
// that records the database version number.
func (p PostgreSQLProvider) QueryRecordVersionUpgradeLegacy(version int) string {
	// This provider has no legacy schema.
	return p.QueryRecordVersionUpgrade(version)
}

// QueryGetDatabaseVersion returns the schema version number.
func (p PostgreSQLProvider) QueryGetDatabaseVersion() string {
	return "SELECT c_config -> 'database' FROM dmz_config WHERE c_key = 'META';"
}

// QueryGetDatabaseVersionLegacy returns the schema version number before The Great Schema Migration (v25, MySQL).
func (p PostgreSQLProvider) QueryGetDatabaseVersionLegacy() string {
	// This provider has no legacy schema.
	return p.QueryGetDatabaseVersion()
}

// QueryTableList returns a list tables in Documize database.
func (p PostgreSQLProvider) QueryTableList() string {
	return fmt.Sprintf(`select table_name
        FROM information_schema.tables
        WHERE table_type='BASE TABLE' AND table_schema NOT IN ('pg_catalog', 'information_schema') AND table_catalog='%s'`, p.DatabaseName())
}

// JSONEmpty returns empty SQL JSON object.
// Typically used as 2nd parameter to COALESCE().
func (p PostgreSQLProvider) JSONEmpty() string {
	return "'{}'::json"
}

// JSONGetValue returns JSON attribute selection syntax.
// Typically used in SELECT <my_json_field> query.
func (p PostgreSQLProvider) JSONGetValue(column, attribute string) string {
	return fmt.Sprintf("%s -> '%s'", column, attribute)
}

// VerfiyVersion checks to see if actual database meets
// minimum version requirements.``
func (p PostgreSQLProvider) VerfiyVersion(dbVersion string) (bool, string) {
	// All versions supported.
	return true, ""
}

// VerfiyCharacterCollation needs to ensure utf8.
func (p PostgreSQLProvider) VerfiyCharacterCollation(charset, collation string) (charOK bool, requirements string) {
	if strings.ToLower(charset) != "utf8" {
		return false, fmt.Sprintf("PostgreSQL character set needs to be utf8, found %s", charset)
	}

	// Collation check ignored.

	return true, ""
}
