# Football League Database Setup Guide

##  Complete Setup Steps

### 1. First Time PostgreSQL Setup

#### Connect to PostgreSQL (Terminal/Command Prompt):
```bash
# Connect as postgres user
psql -U postgres

#### Create Database:
```sql
-- Create the database
CREATE DATABASE football_league;

-- Switch to the database
\c football_league;

-- Check you're in the right database
SELECT current_database();
```

### 2. DBeaver Connection Setup

#### Open DBeaver and Create New Connection:

1. **Click**: "New Database Connection" (plug icon)
2. **Choose**: PostgreSQL
3. **Enter Connection Details**:
   - **Host**: `localhost`
   - **Port**: `5432`
   - **Database**: `football_league`
   - **Username**: `postgres`
   - **Password**: `[your password]`

#### Test Connection:
- Click **"Test Connection"**
- Should show: "Connected" ✅
- Click **"Finish"**

### 3. Run Schema Script

#### In DBeaver:
1. **Right-click** your database connection → **"SQL Editor"** → **"Open SQL Script"**
2. **Navigate** to your project: `insider-case/insider-league/database/schema.sql`
3. **Click** the **"Execute Script"** button (▶️)
4. **Verify**: Should see "Script executed successfully"

#### Or via Terminal:
```bash
# Navigate to your project
cd /Users/iremsuozdemir/Desktop/insider-case/insider-league

# Run the schema
psql -U postgres -d football_league -f database/schema.sql
```

### 4. Verify Setup

#### Check Tables in DBeaver:
1. **Expand** your database connection
2. **Expand** "Schemas" → "public" → "Tables"
3. **You should see**:
   - `teams` (8 sample teams)
   - `leagues`
   - `matches`
   - `team_stats`
   - `league_teams`

#### Test with SQL:
```sql
-- View sample teams
SELECT * FROM teams;

-- Check team count
SELECT COUNT(*) FROM teams;

-- View leagues
SELECT * FROM leagues;
```









