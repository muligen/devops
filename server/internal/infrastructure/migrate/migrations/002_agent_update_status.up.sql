-- +migrate Up
-- Create agent_update_status table for tracking agent update progress
CREATE TABLE agent_update_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL,
    version_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_update_status_agent FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE,
    CONSTRAINT fk_update_status_version FOREIGN KEY (version_id) REFERENCES versions(id) ON DELETE CASCADE,
    CONSTRAINT chk_status CHECK (status IN ('pending', 'downloading', 'installing', 'success', 'failed'))
);

CREATE INDEX ix_update_status_agent ON agent_update_status(agent_id);
CREATE INDEX ix_update_status_version ON agent_update_status(version_id);
CREATE INDEX ix_update_status_status ON agent_update_status(status);

-- +migrate Down
DROP TABLE IF EXISTS agent_update_status;
