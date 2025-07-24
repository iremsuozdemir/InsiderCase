-- VIEW LEAGUE TABLE
-- ==========================================

-- League table sorted by points, goal difference, goals scored
SELECT 
    ROW_NUMBER() OVER (ORDER BY ts.points DESC, ts.goal_difference DESC, ts.goals_for DESC) as position,
    t.name as team_name,
    ts.played,
    ts.won,
    ts.drawn,
    ts.lost,
    ts.goals_for,
    ts.goals_against,
    ts.goal_difference,
    ts.points
FROM team_stats ts
JOIN teams t ON ts.team_id = t.id
WHERE ts.league_id = 1
ORDER BY ts.points DESC, ts.goal_difference DESC, ts.goals_for DESC;

-- ==========================================
-- VIEW MATCH RESULTS
-- ==========================================

-- All matches with team names
SELECT 
    m.week_number,
    ht.name as home_team,
    m.home_score,
    m.away_score,
    at.name as away_team,
    m.played_at
FROM matches m
JOIN teams ht ON m.home_team_id = ht.id
JOIN teams at ON m.away_team_id = at.id  
WHERE m.league_id = 1
ORDER BY m.week_number, m.played_at;


-- CLEANUP (IF NEEDED)
-- ==========================================

-- Delete test data (uncomment if you want to clean up)
-- DELETE FROM matches WHERE league_id = 1;
-- DELETE FROM team_stats WHERE league_id = 1;
-- DELETE FROM league_teams WHERE league_id = 1;
-- DELETE FROM leagues WHERE id = 1; 