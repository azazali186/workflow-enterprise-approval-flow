package handler

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"github.com/aeroxe/approval-flow/internal/config"
)

// setupTestDB creates an in-memory SQLite database with test tables.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Skip if CGO is not available (SQLite requires CGO)
	if out, err := exec.Command("go", "env", "CGO_ENABLED").Output(); err != nil || string(out) == "0\n" {
		t.Skip("Skipping SQLite tests: CGO_ENABLED=0")
	}

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create test tables
	err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			status TEXT DEFAULT 'active',
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			category TEXT,
			is_active BOOLEAN DEFAULT 1,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE templates (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE roles (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE applications (
			id TEXT PRIMARY KEY,
			title TEXT,
			status TEXT DEFAULT 'submitted',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE approvals (
			id TEXT PRIMARY KEY,
			status TEXT DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			deleted_at DATETIME
		)
	`).Error
	require.NoError(t, err)

	return db
}

// insertTestData inserts test data into the database.
func insertTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	// Insert users
	users := []struct {
		ID     string
		Name   string
		Email  string
		Status string
	}{
		{"user-1", "Alice Johnson", "alice@example.com", "active"},
		{"user-2", "Bob Smith", "bob@example.com", "active"},
		{"user-3", "Charlie Brown", "charlie@example.com", "inactive"},
	}

	for _, u := range users {
		err := db.Exec(
			"INSERT INTO users (id, name, email, status) VALUES (?, ?, ?, ?)",
			u.ID, u.Name, u.Email, u.Status,
		).Error
		require.NoError(t, err)
	}

	// Insert workflows
	workflows := []struct {
		ID       string
		Name     string
		Category string
		Active   bool
	}{
		{"wf-1", "Expense Approval", "finance", true},
		{"wf-2", "Leave Request", "hr", true},
		{"wf-3", "Old Workflow", "legacy", false},
	}

	for _, wf := range workflows {
		err := db.Exec(
			"INSERT INTO workflows (id, name, category, is_active) VALUES (?, ?, ?, ?)",
			wf.ID, wf.Name, wf.Category, wf.Active,
		).Error
		require.NoError(t, err)
	}

	// Insert templates
	templates := []struct {
		ID   string
		Name string
	}{
		{"tmpl-1", "Expense Form"},
		{"tmpl-2", "Leave Form"},
	}

	for _, tmpl := range templates {
		err := db.Exec(
			"INSERT INTO templates (id, name) VALUES (?, ?)",
			tmpl.ID, tmpl.Name,
		).Error
		require.NoError(t, err)
	}

	// Insert roles
	roles := []struct {
		ID   string
		Name string
	}{
		{"role-1", "admin"},
		{"role-2", "manager"},
		{"role-3", "user"},
	}

	for _, role := range roles {
		err := db.Exec(
			"INSERT INTO roles (id, name) VALUES (?, ?)",
			role.ID, role.Name,
		).Error
		require.NoError(t, err)
	}

	// Insert applications
	applications := []struct {
		ID     string
		Title  string
		Status string
	}{
		{"app-1", "Travel Expense", "submitted"},
		{"app-2", "Equipment Purchase", "approved"},
		{"app-3", "Training Request", "completed"},
	}

	for _, app := range applications {
		err := db.Exec(
			"INSERT INTO applications (id, title, status) VALUES (?, ?, ?)",
			app.ID, app.Title, app.Status,
		).Error
		require.NoError(t, err)
	}

	// Insert approvals
	approvals := []struct {
		ID     string
		Status string
	}{
		{"apr-1", "pending"},
		{"apr-2", "approved"},
		{"apr-3", "rejected"},
	}

	for _, apr := range approvals {
		err := db.Exec(
			"INSERT INTO approvals (id, status) VALUES (?, ?)",
			apr.ID, apr.Status,
		).Error
		require.NoError(t, err)
	}
}

// createTestHandler creates a handler with test database.
func createTestHandler(db *gorm.DB) *DropdownHandler {
	cfg := &config.Config{
		AppName: "test",
	}

	return NewDropdownHandler(db, cfg)
}

// ==================== Database Integration Tests ====================

func TestDropdownDB_ListUsers(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	type row struct {
		ID   string
		Name string
	}

	var rows []row
	err := db.WithContext(ctx).
		Table("users").
		Select("id, name || ' (' || email || ')' as name").
		Where("deleted_at IS NULL AND status = ?", "active").
		Order("name ASC").
		Find(&rows).Error

	require.NoError(t, err)
	assert.Len(t, rows, 2) // Only active users
	assert.Equal(t, "user-1", rows[0].ID) // Alice comes first alphabetically
	assert.Equal(t, "user-2", rows[1].ID) // Bob comes second
}

func TestDropdownDB_ListWorkflows(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	// Test active workflows only
	var activeWorkflows []struct {
		ID   string
		Name string
	}
	err := db.WithContext(ctx).
		Table("workflows").
		Select("id, name").
		Where("deleted_at IS NULL AND is_active = ?", true).
		Order("name ASC").
		Find(&activeWorkflows).Error

	require.NoError(t, err)
	assert.Len(t, activeWorkflows, 2) // Only active workflows

	// Test all workflows
	var allWorkflows []struct {
		ID   string
		Name string
	}
	err = db.WithContext(ctx).
		Table("workflows").
		Select("id, name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&allWorkflows).Error

	require.NoError(t, err)
	assert.Len(t, allWorkflows, 3) // All workflows
}

func TestDropdownDB_ListTemplates(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	var templates []struct {
		ID   string
		Name string
	}
	err := db.WithContext(ctx).
		Table("templates").
		Select("id, name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&templates).Error

	require.NoError(t, err)
	assert.Len(t, templates, 2)
}

func TestDropdownDB_ListRoles(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	var roles []struct {
		ID   string
		Name string
	}
	err := db.WithContext(ctx).
		Table("roles").
		Select("id, name").
		Where("deleted_at IS NULL").
		Order("name ASC").
		Find(&roles).Error

	require.NoError(t, err)
	assert.Len(t, roles, 3)
}

func TestDropdownDB_ListApplications(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	// Test submitted applications only
	var submitted []struct {
		ID    string
		Title string
	}
	err := db.WithContext(ctx).
		Table("applications").
		Select("id, COALESCE(title, 'Untitled') as title").
		Where("deleted_at IS NULL AND status IN ?", []string{"submitted"}).
		Order("created_at DESC").
		Limit(100).
		Find(&submitted).Error

	require.NoError(t, err)
	assert.Len(t, submitted, 1)

	// Test all statuses
	var all []struct {
		ID    string
		Title string
	}
	err = db.WithContext(ctx).
		Table("applications").
		Select("id, COALESCE(title, 'Untitled') as title").
		Where("deleted_at IS NULL AND status IN ?", []string{"submitted", "approved", "completed"}).
		Order("created_at DESC").
		Limit(100).
		Find(&all).Error

	require.NoError(t, err)
	assert.Len(t, all, 3)
}

func TestDropdownDB_ListApprovals(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	var approvals []struct {
		ID     string
		Status string
	}
	err := db.WithContext(ctx).
		Table("approvals").
		Select("id, status").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(100).
		Find(&approvals).Error

	require.NoError(t, err)
	assert.Len(t, approvals, 3)
}

func TestDropdownDB_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	// Soft delete a user
	err := db.Exec("UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = ?", "user-1").Error
	require.NoError(t, err)

	// Verify soft deleted user is excluded
	var users []struct {
		ID   string
		Name string
	}
	err = db.WithContext(ctx).
		Table("users").
		Select("id, name").
		Where("deleted_at IS NULL AND status = ?", "active").
		Find(&users).Error

	require.NoError(t, err)
	assert.Len(t, users, 1) // Only Bob (Alice is soft deleted)
	assert.Equal(t, "user-2", users[0].ID)
}

func TestDropdownDB_Pagination(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	// Test with limit
	var applications []struct {
		ID    string
		Title string
	}
	err := db.WithContext(ctx).
		Table("applications").
		Select("id, title").
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(2).
		Find(&applications).Error

	require.NoError(t, err)
	assert.Len(t, applications, 2)
}

func TestDropdownDB_SpecialCharacters(t *testing.T) {
	db := setupTestDB(t)

	// Insert user with special characters
	err := db.Exec(
		"INSERT INTO users (id, name, email, status) VALUES (?, ?, ?, ?)",
		"user-special", "O'Brien & Sons <script>", "special@test.com", "active",
	).Error
	require.NoError(t, err)

	ctx := context.Background()

	var users []struct {
		ID   string
		Name string
	}
	err = db.WithContext(ctx).
		Table("users").
		Select("id, name || ' (' || email || ')' as name").
		Where("deleted_at IS NULL AND status = ?", "active").
		Find(&users).Error

	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Contains(t, users[0].Name, "O'Brien")
}

func TestDropdownDB_EmptyResults(t *testing.T) {
	db := setupTestDB(t)
	// Don't insert any data

	ctx := context.Background()

	var users []struct {
		ID   string
		Name string
	}
	err := db.WithContext(ctx).
		Table("users").
		Select("id, name").
		Where("deleted_at IS NULL AND status = ?", "active").
		Find(&users).Error

	require.NoError(t, err)
	assert.Empty(t, users)
}

func TestDropdownDB_ConcurrentQueries(t *testing.T) {
	db := setupTestDB(t)
	insertTestData(t, db)

	ctx := context.Background()

	// Run multiple queries concurrently
	for i := 0; i < 10; i++ {
		go func() {
			var users []struct {
				ID   string
				Name string
			}
			err := db.WithContext(ctx).
				Table("users").
				Select("id, name").
				Where("deleted_at IS NULL AND status = ?", "active").
				Find(&users).Error
			assert.NoError(t, err)
		}()
	}
}

func TestDropdownDB_LargeDataset(t *testing.T) {
	db := setupTestDB(t)

	// Insert many users
	for i := 0; i < 100; i++ {
		err := db.Exec(
			"INSERT INTO users (id, name, email, status) VALUES (?, ?, ?, ?)",
			"user-1", "User 1", "user1@test.com", "active",
		).Error
		require.NoError(t, err)
	}

	ctx := context.Background()

	var users []struct {
		ID   string
		Name string
	}
	err := db.WithContext(ctx).
		Table("users").
		Select("id, name").
		Where("deleted_at IS NULL AND status = ?", "active").
		Order("name ASC").
		Find(&users).Error

	require.NoError(t, err)
	assert.Len(t, users, 100)
}
