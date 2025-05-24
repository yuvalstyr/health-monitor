-- Add frequency column to gauges table
ALTER TABLE gauges ADD COLUMN frequency TEXT NOT NULL DEFAULT 'monthly';

-- Add direction column to gauges table
ALTER TABLE gauges ADD COLUMN direction TEXT NOT NULL DEFAULT 'under';
