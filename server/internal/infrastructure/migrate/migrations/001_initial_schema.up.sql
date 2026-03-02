-- +migrate Up
-- Create users table for authentication and authorization
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_users_username UNIQUE (username),
    CONSTRAINT uq_users_email UNIQUE (email)
);

CREATE INDEX ix_users_status ON users(status);
CREATE INDEX ix_users_deleted ON users(deleted_at);

-- Create agents table for registered agents
CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'offline',
    version VARCHAR(20) NOT NULL DEFAULT '0.0.0',
    hostname VARCHAR(255),
    ip_address VARCHAR(45),
    os_info VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    last_seen_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_agents_name UNIQUE (name)
);

CREATE INDEX ix_agents_status ON agents(status);
CREATE INDEX ix_agents_deleted ON agents(deleted_at);
CREATE INDEX ix_agents_last_seen ON agents(last_seen_at);

-- Create tasks table for command execution
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    params JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    priority INTEGER NOT NULL DEFAULT 0,
    timeout INTEGER NOT NULL DEFAULT 300,
    result JSONB,
    output TEXT,
    exit_code INTEGER,
    duration FLOAT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_tasks_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE INDEX ix_tasks_agent ON tasks(agent_id);
CREATE INDEX ix_tasks_status ON tasks(status);
CREATE INDEX ix_tasks_created ON tasks(created_at);
CREATE INDEX ix_tasks_agent_created ON tasks(agent_id, created_at);

-- Create agent_metrics table for time-series metrics
CREATE TABLE agent_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    cpu_usage FLOAT NOT NULL DEFAULT 0,
    memory_total BIGINT NOT NULL DEFAULT 0,
    memory_used BIGINT NOT NULL DEFAULT 0,
    memory_percent FLOAT NOT NULL DEFAULT 0,
    disk_total BIGINT NOT NULL DEFAULT 0,
    disk_used BIGINT NOT NULL DEFAULT 0,
    disk_percent FLOAT NOT NULL DEFAULT 0,
    uptime BIGINT NOT NULL DEFAULT 0,
    collected_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_metrics_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);

CREATE INDEX ix_metrics_agent ON agent_metrics(agent_id);
CREATE INDEX ix_metrics_collected ON agent_metrics(collected_at);
CREATE INDEX ix_metrics_agent_collected ON agent_metrics(agent_id, collected_at);

-- Create versions table for agent updates
CREATE TABLE versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version VARCHAR(20) NOT NULL,
    platform VARCHAR(50) NOT NULL DEFAULT 'windows',
    file_url VARCHAR(500),
    file_hash VARCHAR(64),
    file_size BIGINT,
    signature TEXT,
    release_notes TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID REFERENCES users(id),

    CONSTRAINT uq_versions_version_platform UNIQUE (version, platform)
);

CREATE INDEX ix_versions_active ON versions(is_active);
CREATE INDEX ix_versions_created ON versions(created_at);

-- Create alert_rules table for monitoring alerts
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    metric_type VARCHAR(50) NOT NULL,
    condition VARCHAR(10) NOT NULL,
    threshold FLOAT NOT NULL,
    duration INTEGER NOT NULL DEFAULT 60,
    severity VARCHAR(20) NOT NULL DEFAULT 'warning',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_condition CHECK (condition IN ('>', '>=', '<', '<=', '==', '!='))
);

CREATE INDEX ix_alert_rules_enabled ON alert_rules(enabled);

-- Create audit_logs table for user action tracking
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX ix_audit_logs_user ON audit_logs(user_id);
CREATE INDEX ix_audit_logs_action ON audit_logs(action);
CREATE INDEX ix_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX ix_audit_logs_created ON audit_logs(created_at);

-- Insert default admin user
-- Password: 'admin123' (bcrypt hash)
INSERT INTO users (username, password_hash, email, role, status)
VALUES (
    'admin',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3jFpVmPjvPv.FvPvPvPv',
    'admin@agentteams.local',
    'admin',
    'active'
);

-- +migrate Down
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS alert_rules;
DROP TABLE IF EXISTS versions;
DROP TABLE IF EXISTS agent_metrics;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS users;
