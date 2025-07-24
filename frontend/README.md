# Football League Simulator Frontend

A simple, modern web interface for the Football League Simulator API.

##  How to Use

### 1. Start the Go Server
```bash
cd insider-league
go run main.go router.go
```

### 2. Open the Frontend
Open `index.html` in your web browser:
- Double-click the `index.html` file, or
- Right-click → "Open with" → Your preferred browser

### 3. Use the Interface

#### **Create League**
- Click "Create League" to initialize a new league with 8 teams
- This will populate the database and show initial standings

#### **Play Matches**
- **Next Week**: Play one week of matches at a time
- **Play All**: Automatically play all remaining weeks until league completion

#### **View Results**
- **League Table**: Shows current standings with points, wins, draws, losses, and goal difference
- **Match Results**: Displays the latest week's match results

#### **Clear League**
- Click "Clear League" to reset everything and start fresh

##  Features

-  **Real-time Updates**: See results immediately after playing matches
-  **Responsive Design**: Works on desktop and mobile devices
-  **Visual Feedback**: Loading states and success/error messages
-  **Clean Interface**: Modern grey theme with intuitive layout
-  **Database Integration**: All data persists in PostgreSQL

##  Interface Layout

```
┌─────────────────────────────────────────────────────────┐
│                    Football League Simulator            │
├─────────────────────────────────────────────────────────┤
│  [Create League] [Clear League]                         │
├─────────────────────────────────────────────────────────┤
│  League Table                    │  Match Results       │
│  ┌─────────────────────────────┐ │  ┌─────────────────┐ │
│  │ Teams │ PTS │ P │ W │ D │ L │ │  │ Week: 1 / 28    │ │
│  │───────│─────│───│───│───│───│ │  │                 │ │
│  │ 1. Arsenal │ 3 │ 1 │ 1 │ 0 │ 0 │ │  │ Arsenal 2-1 Chelsea │ │
│  │ 2. Chelsea │ 0 │ 1 │ 0 │ 0 │ 1 │ │  │ Liverpool 1-1 City  │ │
│  └─────────────────────────────┘ │  └─────────────────┘ │
│  [Play All]                      │  [Next Week]         │
└─────────────────────────────────────────────────────────┘
```

##  Technical Details

- **Frontend**: HTML5, CSS3, Vanilla JavaScript (ES6+)
- **API**: RESTful Go backend with PostgreSQL
- **CORS**: Enabled for cross-origin requests
- **Responsive**: Mobile-friendly design

##  Troubleshooting

### Frontend won't load data
- Make sure your Go server is running on `http://localhost:8080`
- Check browser console for CORS errors
- Verify database connection

### Buttons not working
- Ensure JavaScript is enabled in your browser
- Check that all files (`index.html`, `styles.css`, `script.js`) are in the same folder

### Database issues
- Make sure PostgreSQL is running
- Check that the database schema is properly set up
- Verify connection settings in `database/connection.go`


**Enjoy simulating your football league!** 