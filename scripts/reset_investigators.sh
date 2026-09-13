#!/usr/bin/env bash
set -euo pipefail

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-gouser}"
DB_NAME="${DB_NAME:-aeons}"
export PGPASSWORD="${PGPASSWORD:-gopassword}"

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -v ON_ERROR_STOP=1 <<'SQL'

BEGIN;

DROP TABLE IF EXISTS plays;
DROP TABLE IF EXISTS sessions;  
DROP TABLE IF EXISTS scenarios;
DROP TABLE IF EXISTS campaigns;
DROP TABLE IF EXISTS investigators;
DROP TABLE IF EXISTS classes;

CREATE TABLE classes (
    class_id SERIAL PRIMARY KEY,
    class_name TEXT NOT NULL
);

CREATE TABLE campaigns (
    campaign_id      SERIAL PRIMARY KEY,
    campaign_name	 TEXT NOT NULL
);

CREATE TABLE investigators (
    investigator_id    SERIAL PRIMARY KEY,
    investigator_name  TEXT NOT NULL,
    investigator_class INTEGER NOT NULL REFERENCES classes(class_id)
);

CREATE TABLE scenarios (
    scenario_id		    	    SERIAL PRIMARY KEY,
    scenario_name		        TEXT NOT NULL,
    scenario_campaign_position 	INTEGER NOT NULL,
    scenario_parent_campaign    INTEGER REFERENCES campaigns(campaign_id)
);

CREATE TABLE sessions (
    session_id			SERIAL PRIMARY KEY,
    session_name        TEXT NOT NULL,
    session_scenario    INTEGER NOT NULL REFERENCES scenarios(scenario_id)
);

CREATE TABLE plays (
    play_id             SERIAL PRIMARY KEY,
    play_investigator   INTEGER NOT NULL REFERENCES investigators(investigator_id),
    play_session        INTEGER NOT NULL REFERENCES sessions(session_id),
    play_self_played    BOOLEAN DEFAULT TRUE
);

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO gouser;

COMMIT;
SQL