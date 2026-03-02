-- +migrate Up
-- Create alert_events table for storing alert event history
CREATE TABLE alert_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id UUID REFERENCES alert_rules(id) ON DELETE SET NULL,
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,
    metric_value FLOAT NOT NULL,
    threshold FLOAT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    message TEXT,
    triggered_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    acknowledged_by UUID REFERENCES users(id) ON DELETE SET NULL,
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_status CHECK (status IN ('pending', 'acknowledged', 'resolved'))
);

CREATE INDEX ix_alert_events_status ON alert_events(status);
CREATE INDEX ix_alert_events_triggered ON alert_events(triggered_at);
CREATE INDEX ix_alert_events_agent ON alert_events(agent_id);
CREATE INDEX ix_alert_events_rule ON alert_events(rule_id);

-- +migrate Down
DROP INDEX IF EXISTS ix_alert_events_rule;
DROP INDEX IF EXISTS ix_alert_events_agent;
DROP INDEX IF EXISTS ix_alert_events_triggered;
DROP INDEX IF EXISTS ix_alert_events_status;
DROP TABLE IF EXISTS alert_events;
