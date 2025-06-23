ALTER TABLE amocrm_fields ADD COLUMN entity_type VARCHAR(50) NOT NULL DEFAULT 'leads';

-- Create index on entity_type for faster queries
CREATE INDEX idx_amocrm_fields_entity_type ON amocrm_fields(entity_type);