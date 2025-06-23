CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- AmoCRM fields table
CREATE TABLE amocrm_fields (
                               id BIGINT PRIMARY KEY,
                               name VARCHAR(255) NOT NULL,
                               type VARCHAR(50) NOT NULL,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                               updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Dialer schedulers table
CREATE TABLE dialer_schedulers (
                                   id UUID PRIMARY KEY,
                                   name VARCHAR(255) NOT NULL,
                                   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Dialer campaigns table
CREATE TABLE dialer_campaigns (
                                  id UUID PRIMARY KEY,
                                  name VARCHAR(255) NOT NULL,
                                  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Dialer buckets table
CREATE TABLE dialer_buckets (
                                id UUID PRIMARY KEY,
                                campaign_id UUID NOT NULL REFERENCES dialer_campaigns(id),
                                name VARCHAR(255) NOT NULL,
                                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Integration flows table
CREATE TABLE integration_flows (
                                   id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                                   name VARCHAR(255) NOT NULL,
                                   flow_data JSONB NOT NULL,
                                   is_active BOOLEAN DEFAULT true,
                                   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Webhook logs table
CREATE TABLE webhook_logs (
                              id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                              webhook_type VARCHAR(100) NOT NULL,
                              raw_data JSONB NOT NULL,
                              processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                              status VARCHAR(50) NOT NULL
);

-- Create indexes
CREATE INDEX idx_webhook_logs_type ON webhook_logs(webhook_type);
CREATE INDEX idx_webhook_logs_status ON webhook_logs(status);
CREATE INDEX idx_webhook_logs_processed_at ON webhook_logs(processed_at);
CREATE INDEX idx_integration_flows_active ON integration_flows(is_active);

-- Create update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
RETURN NEW;
END;
$ language 'plpgsql';

CREATE TRIGGER update_amocrm_fields_updated_at BEFORE UPDATE ON amocrm_fields
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dialer_schedulers_updated_at BEFORE UPDATE ON dialer_schedulers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dialer_campaigns_updated_at BEFORE UPDATE ON dialer_campaigns
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dialer_buckets_updated_at BEFORE UPDATE ON dialer_buckets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_integration_flows_updated_at BEFORE UPDATE ON integration_flows
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();