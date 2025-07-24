-- Football League Database Schema
-- Run this script in PostgreSQL to create your database

-- Create database (run this first, separately)
-- CREATE DATABASE football_league;
-- \c football_league;

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    strength INTEGER NOT NULL CHECK (strength >= 1 AND strength <= 100)
);

-- Leagues table  
CREATE TABLE IF NOT EXISTS leagues (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    current_week INTEGER DEFAULT 0,
    total_weeks INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'paused'))
);

-- League teams (many-to-many relationship)
CREATE TABLE IF NOT EXISTS league_teams (
    id SERIAL PRIMARY KEY,
    league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    UNIQUE(league_id, team_id)
);

-- Matches table
CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
    week_number INTEGER NOT NULL,
    home_team_id INTEGER REFERENCES teams(id),
    away_team_id INTEGER REFERENCES teams(id),
    home_score INTEGER DEFAULT NULL,
    away_score INTEGER DEFAULT NULL,
    played BOOLEAN DEFAULT FALSE,
    played_at INTEGER DEFAULT NULL,
    CHECK (home_team_id != away_team_id)
);

-- Team statistics table
CREATE TABLE IF NOT EXISTS team_stats (
    id SERIAL PRIMARY KEY,
    league_id INTEGER REFERENCES leagues(id) ON DELETE CASCADE,
    team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
    played INTEGER DEFAULT 0,
    won INTEGER DEFAULT 0,
    drawn INTEGER DEFAULT 0,
    lost INTEGER DEFAULT 0,
    goals_for INTEGER DEFAULT 0,
    goals_against INTEGER DEFAULT 0,
    points INTEGER DEFAULT 0,
    goal_difference INTEGER GENERATED ALWAYS AS (goals_for - goals_against) STORED,
    UNIQUE(league_id, team_id)
);

-- Insert sample teams (4 teams for smaller league)
INSERT INTO teams (name, strength) VALUES 
    ('Arsenal', 70),
    ('Chelsea', 85),
    ('Liverpool', 75),
    ('Manchester City', 92)
ON CONFLICT (name) DO NOTHING;

