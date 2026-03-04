// Package e2e provides end-to-end testing infrastructure
package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	agentdomain "github.com/agentteams/server/internal/modules/agent/domain"
	authdomain "github.com/agentteams/server/internal/modules/auth/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TestUsers contains predefined test users
var TestUsers = struct {
	Admin    *authdomain.User
	Operator *authdomain.User
	Viewer   *authdomain.User
}{
	Admin: &authdomain.User{
		ID:       uuid.NewString(),
		Username: "test-admin",
		Email:    "admin@test.com",
		Role:     authdomain.RoleAdmin,
	},
	Operator: &authdomain.User{
		ID:       uuid.NewString(),
		Username: "test-operator",
		Email:    "operator@test.com",
		Role:     authdomain.RoleOperator,
	},
	Viewer: &authdomain.User{
		ID:       uuid.NewString(),
		Username: "test-viewer",
		Email:    "viewer@test.com",
		Role:     authdomain.RoleViewer,
	},
}

// TestAgents contains predefined test agents
var TestAgents = struct {
	Online  *agentdomain.Agent
	Offline *agentdomain.Agent
}{
	Online: &agentdomain.Agent{
		ID:        uuid.NewString(),
		Name:      "test-agent-online",
		TokenHash: hashToken("test-token-online"),
		Status:    agentdomain.StatusOnline,
		IPAddress: "192.168.1.100",
		OSInfo:    "Windows 10",
		Version:   "1.0.0",
	},
	Offline: &agentdomain.Agent{
		ID:        uuid.NewString(),
		Name:      "test-agent-offline",
		TokenHash: hashToken("test-token-offline"),
		Status:    agentdomain.StatusOffline,
		IPAddress: "192.168.1.101",
		OSInfo:    "Windows 11",
		Version:   "1.0.0",
	},
}

// hashToken creates a SHA256 hash of the token
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// FixtureLoader loads test fixtures into the database
type FixtureLoader struct {
	db *gorm.DB
}

// NewFixtureLoader creates a new fixture loader
func NewFixtureLoader(db *gorm.DB) *FixtureLoader {
	return &FixtureLoader{db: db}
}

// LoadAll loads all fixtures
func (l *FixtureLoader) LoadAll(ctx context.Context) error {
	if err := l.LoadUsers(ctx); err != nil {
		return err
	}
	if err := l.LoadAgents(ctx); err != nil {
		return err
	}
	return nil
}

// LoadUsers loads user fixtures
func (l *FixtureLoader) LoadUsers(ctx context.Context) error {
	users := []*authdomain.User{
		TestUsers.Admin,
		TestUsers.Operator,
		TestUsers.Viewer,
	}

	for _, user := range users {
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		// Set a default password hash for test users
		user.PasswordHash = hashToken("test-password")

		if err := l.db.WithContext(ctx).Create(user).Error; err != nil {
			return err
		}
	}
	return nil
}

// LoadAgents loads agent fixtures
func (l *FixtureLoader) LoadAgents(ctx context.Context) error {
	agents := []*agentdomain.Agent{
		TestAgents.Online,
		TestAgents.Offline,
	}

	for _, agent := range agents {
		agent.CreatedAt = time.Now()
		agent.UpdatedAt = time.Now()

		if err := l.db.WithContext(ctx).Create(agent).Error; err != nil {
			return err
		}
	}
	return nil
}

// Cleanup removes all test data
func (l *FixtureLoader) Cleanup(ctx context.Context) error {
	return l.db.WithContext(ctx).Exec("TRUNCATE users, agents CASCADE").Error
}
