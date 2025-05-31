-- name: GetGauge :one
SELECT * FROM gauges WHERE id = ? LIMIT 1;

-- name: ListGauges :many
SELECT * FROM gauges ORDER BY name;

-- name: CreateGauge :one
INSERT INTO gauges (name, description, target, value, unit, icon, frequency, direction)
VALUES (?, ?, ?, 0, ?, ?, ?, ?)
RETURNING *;

-- name: UpdateGauge :exec
UPDATE gauges
SET name = ?,
    description = ?,
    target = ?,
    unit = ?,
    icon = ?,
    frequency = ?,
    direction = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: UpdateGaugeValue :exec
UPDATE gauges
SET value = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteGauge :exec
DELETE FROM gauges WHERE id = ?;

-- name: GetCurrentValue :one
SELECT CAST(COALESCE(
    (SELECT value FROM gauge_values WHERE gauge_id = ? ORDER BY date DESC LIMIT 1),
    0.0
) AS REAL) as value;

-- name: CreateGaugeValue :exec
INSERT INTO gauge_values (gauge_id, value, date)
VALUES (?, CAST(? AS REAL), ?);

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
