-- Create a sample time-series table
CREATE TABLE sensor_data (
  time        TIMESTAMPTZ NOT NULL,
  sensor_id   INTEGER,
  temperature DOUBLE PRECISION,
  humidity    DOUBLE PRECISION
);

-- Convert to hypertable
SELECT create_hypertable('sensor_data', 'time');

-- Create an index to improve query performance
CREATE INDEX ON sensor_data (sensor_id, time DESC);

-- Set retention policy (30 days)
SELECT add_retention_policy('sensor_data', INTERVAL '30 days');

-- Create a continuous aggregate view
CREATE MATERIALIZED VIEW sensor_daily_stats
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 day', time) AS bucket,
       sensor_id,
       avg(temperature) as avg_temp,
       min(temperature) as min_temp,
       max(temperature) as max_temp,
       avg(humidity) as avg_humidity
FROM sensor_data
GROUP BY bucket, sensor_id; 