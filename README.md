# Football League Simulator

A comprehensive football league management and simulation system built with Go backend and vanilla JavaScript frontend. Create teams, generate fixtures, simulate realistic matches, and track complete league statistics.

## Features

- **Team Management**: Add teams with customizable strength ratings (1-100)
- **Automatic Fixture Generation**: Round-robin scheduling algorithm for balanced competition
- **Realistic Match Simulation**: Probabilistic match outcomes considering team strength, home advantage, and recent form
- **Live League Table**: Real-time standings following Premier League rules (3 points for wins, 1 for draws)
- **Comprehensive Statistics**: Goals, wins/draws/losses, goal difference tracking
- **Web Interface**: User-friendly frontend for league management
- **RESTful API**: Complete API for programmatic access

## Technologies

- **Backend**: Go 1.24.4+
- **Database**: PostgreSQL
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Deployment**: Heroku-ready with Procfile and app.json

## Quick Start

### Prerequisites
- Go 1.24.4+ installed
- PostgreSQL running locally

### Local Setup
1. Clone the repository and navigate to the project directory
2. Install dependencies: `go mod tidy`
3. Set up PostgreSQL database and run the schema from `database/schema.sql`
4. Set database connection: `export DATABASE_URL="postgres://username:password@localhost:5432/football_league?sslmode=disable"`
5. Run the application: `go run main.go router.go`
6. Access the web interface at `http://localhost:8080`

### Deployment
The project includes Heroku configuration files (`Procfile`, `app.json`) for easy deployment. Simply connect your repository to Heroku and deploy.

## API Endpoints

### League Operations
- `POST /api/league` - Create new league
- `DELETE /api/league` - Clear league
- `GET /api/league/status` - Get league info

### Match Simulation
- `POST /api/league/play-week` - Play one week
- `POST /api/league/play-all` - Play entire season

### Data Retrieval
- `GET /api/league/table` - League standings
- `GET /api/league/matches` - All match results
- `GET /api/league/matches/week/{week}` - Specific week results

### Team Management
- `GET /api/teams` - List all teams
- `POST /api/teams` - Add team
- `PUT /api/teams/{id}` - Update team
- `DELETE /api/teams/{id}` - Delete team

## Usage

1. **Create League**: Initialize a new league with teams
2. **Add Teams**: Add teams with custom names and strength ratings
3. **Simulate Matches**: Play matches week by week or simulate entire seasons
4. **View Results**: Monitor league table, match results, and statistics
5. **API Access**: Use Postman or any HTTP client to interact with the API

## Database Schema

The application uses PostgreSQL with tables for teams, leagues, matches, and team statistics. Sample data includes popular teams like Arsenal, Chelsea, Liverpool, and Manchester City.

## License

This project is available for educational and demonstration purposes.
