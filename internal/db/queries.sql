-- Gauge Templates CRUD Operations
-- name: GetGaugeTemplate :one
SELECT * FROM gauge_templates WHERE id = ? LIMIT 1;

-- name: ListGaugeTemplates :many
SELECT * FROM gauge_templates ORDER BY name;

-- name: ListActiveGaugeTemplates :many
SELECT * FROM gauge_templates WHERE active = true ORDER BY name;

-- name: CreateGaugeTemplate :one
INSERT INTO gauge_templates (name, description, target, unit, icon, frequency, direction, active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateGaugeTemplate :exec
UPDATE gauge_templates
SET name = ?,
    description = ?,
    target = ?,
    unit = ?,
    icon = ?,
    frequency = ?,
    direction = ?,
    active = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteGaugeTemplate :exec
DELETE FROM gauge_templates WHERE id = ?;

-- Gauge Instances CRUD Operations
-- name: GetGaugeInstance :one
SELECT * FROM gauge_instances WHERE id = ? LIMIT 1;

-- name: ListGaugeInstances :many
SELECT * FROM gauge_instances ORDER BY period_start DESC;

-- name: ListGaugeInstancesByTemplate :many
SELECT * FROM gauge_instances WHERE template_id = ? ORDER BY period_start DESC;

-- name: CreateGaugeInstance :one
INSERT INTO gauge_instances (template_id, period_start, value)
VALUES (?, ?, 0)
RETURNING *;

-- name: UpdateGaugeInstanceValue :exec
UPDATE gauge_instances
SET value = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteGaugeInstance :exec
DELETE FROM gauge_instances WHERE id = ?;

-- name: InstanceExistsForPeriod :one
SELECT EXISTS(SELECT 1 FROM gauge_instances 
WHERE template_id = ? AND period_start = ?) as instance_exists;

-- Dashboard and Current Period Queries
-- name: ListCurrentPeriodGaugeInstances :many
SELECT gi.*, gt.name, gt.description, gt.target, gt.unit, gt.icon, gt.frequency, gt.direction
FROM gauge_instances gi
JOIN gauge_templates gt ON gi.template_id = gt.id
WHERE gt.active = true
  AND ((gt.frequency = 'weekly' AND gi.period_start = ?) 
    OR (gt.frequency = 'bi-weekly' AND gi.period_start = ?)
    OR (gt.frequency = 'monthly' AND gi.period_start = ?))
ORDER BY gt.name;

-- name: GetCurrentValue :one
SELECT CAST(COALESCE(
    (SELECT value FROM gauge_values WHERE gauge_id = ? ORDER BY date DESC LIMIT 1),
    0.0
) AS REAL) as value;

-- name: CreateGaugeValue :exec
INSERT INTO gauge_values (gauge_id, value, date)
VALUES (?, ?, ?);

-- name: GetGaugeValues :many
SELECT * FROM gauge_values 
WHERE gauge_id = ?
ORDER BY date DESC;

-- name: GetGaugeHistory :many
SELECT strftime('%Y-%m', date) as month,
       CAST(AVG(value) AS REAL) as average_value
FROM gauge_values
WHERE gauge_id = ?
GROUP BY strftime('%Y-%m', date)
ORDER BY month DESC;
