# Postman Testing Guide for Football League API

### 1. **Import Collection**
1. Open Postman
2. Click **"Import"** button
3. Select the `postman_collection.json` file from your project
4. The collection will appear in your Postman workspace

### 2. **Verify Server is Running**
Your Go server should be running on `http://localhost:8080`

---

##  Testing Sequence

### **Step 1: Health Check**
- **Request**: `GET http://localhost:8080/health`
- **Expected Response**:
```json
{
    "status": "OK",
    "timestamp": "2024-06-09T15:30:00Z"
}
```

### **Step 2: Create League**
- **Request**: `POST http://localhost:8080/api/league`
- **Headers**: `Content-Type: application/json`
- **Body**:
```json
[
  {
    "name": "Arsenal",
    "strength": 90
  },
  {
    "name": "Chelsea", 
    "strength": 85
  },
  {
    "name": "Liverpool",
    "strength": 88
  },
  {
    "name": "Manchester City",
    "strength": 92
  },
  {
    "name": "Manchester United",
    "strength": 82
  },
  {
    "name": "Tottenham",
    "strength": 80
  }
]
```
- **Expected Response**:
```json
{
    "current_week": 0,
    "total_weeks": 15,
    "status": "League created successfully"
}
```

### **Step 3: Get League Table (Initial)**
- **Request**: `GET http://localhost:8080/api/league/table`
- **Expected Response**: Empty table with all teams at 0 points

### **Step 4: Play Week 1**
- **Request**: `POST http://localhost:8080/api/league/play-week`
- **Expected Response**:
```json
{
    "current_week": 1,
    "total_weeks": 15,
    "status": "Week played successfully"
}
```

### **Step 5: Get League Table (After Week 1)**
- **Request**: `GET http://localhost:8080/api/league/table`
- **Expected Response**: Teams with updated stats and points

### **Step 6: Get All Matches**
- **Request**: `GET http://localhost:8080/api/league/matches`
- **Expected Response**: Array of all matches with results

### **Step 7: Get Week 1 Matches**
- **Request**: `GET http://localhost:8080/api/league/matches/1`
- **Expected Response**: Matches from week 1 only

### **Step 8: Get League Status**
- **Request**: `GET http://localhost:8080/api/league/status`
- **Expected Response**:
```json
{
    "current_week": 1,
    "total_weeks": 15,
    "status": "active",
    "progress": "6.67%"
}
```

---

##  Continue Testing

### **Play More Weeks**
Repeat **Step 4** multiple times to play through the season:
- Week 2: `POST /api/league/play-week`
- Week 3: `POST /api/league/play-week`
- etc.

### **Check Progress**
After each week:
1. Get league table to see standings
2. Get matches to see results
3. Get status to see progress

---

##  Expected Results

### **League Table Format**
```json
[
  {
    "team_name": "Manchester City",
    "played": 1,
    "won": 1,
    "drawn": 0,
    "lost": 0,
    "goals_for": 3,
    "goals_against": 0,
    "points": 3,
    "goal_diff": 3,
    "position": 1
  }
]
```

### **Match Format**
```json
{
  "home_team": "Arsenal",
  "away_team": "Chelsea", 
  "home_score": 2,
  "away_score": 1,
  "week": 1
}
```

---
