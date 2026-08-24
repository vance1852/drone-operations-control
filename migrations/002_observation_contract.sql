DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'telemetry_events' AND column_name = 'value'
    ) THEN
        ALTER TABLE telemetry_events RENAME COLUMN value TO risk_score;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'telemetry_events' AND column_name = 'unit'
    ) THEN
        ALTER TABLE telemetry_events RENAME COLUMN unit TO scale;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'telemetry_events' AND column_name = 'limit_value'
    ) THEN
        ALTER TABLE telemetry_events RENAME COLUMN limit_value TO alert_threshold;
    END IF;
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'telemetry_events' AND column_name = 'measured_at'
    ) THEN
        ALTER TABLE telemetry_events RENAME COLUMN measured_at TO observed_at;
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'telemetry_events_risk_score_nonnegative') THEN
        ALTER TABLE telemetry_events
            ADD CONSTRAINT telemetry_events_risk_score_nonnegative CHECK (risk_score >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'telemetry_events_alert_threshold_nonnegative') THEN
        ALTER TABLE telemetry_events
            ADD CONSTRAINT telemetry_events_alert_threshold_nonnegative CHECK (alert_threshold >= 0);
    END IF;
END $$;

ALTER TABLE safety_alerts DROP CONSTRAINT IF EXISTS safety_alerts_kind_check;
UPDATE safety_alerts SET kind = 'reassess' WHERE kind = 'retask';
ALTER TABLE safety_alerts
    ADD CONSTRAINT safety_alerts_kind_check CHECK (kind IN ('reassess', 'repeat_drone', 'safety_adjustment', 'close_record'));
