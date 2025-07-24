class FootballLeagueApp {
    constructor() {
        this.apiBase = 'http://localhost:8080/api';
        this.currentWeek = 0;
        this.totalWeeks = 0;
        this.isLeagueCreated = false;
        this.scheduleData = null;
        
        this.initializeEventListeners();
        this.updateStatus('Ready to start');
        
        // Initialize predictions
        this.updatePredictions([]);
    }

    initializeEventListeners() {
        document.getElementById('initDbBtn').addEventListener('click', () => this.initializeDatabase());
        document.getElementById('createLeagueBtn').addEventListener('click', () => this.createLeague());
        document.getElementById('clearLeagueBtn').addEventListener('click', () => this.clearLeague());
        document.getElementById('nextWeekBtn').addEventListener('click', () => this.playNextWeek());
        document.getElementById('playAllBtn').addEventListener('click', () => this.playAllWeeks());
        document.getElementById('addTeamBtn').addEventListener('click', () => this.addTeam());

        document.getElementById('loadTeamsBtn').addEventListener('click', () => this.loadTeams());
        document.getElementById('clearTeamsBtn').addEventListener('click', () => this.clearTeams());
    }

    async initializeDatabase() {
        try {
            this.updateStatus('Initializing database...');
            this.showLoading('initDbBtn');

            const response = await fetch(`${this.apiBase}/init-db`, {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            
            this.updateStatus('Database initialized successfully!');
            this.showMessage('Database initialized with default teams!', 'success');
            
            // Load teams list
            await this.loadTeams();
            
        } catch (error) {
            console.error('Error initializing database:', error);
            this.updateStatus('Failed to initialize database');
            this.showMessage('Failed to initialize database: ' + error.message, 'error');
        } finally {
            this.hideLoading('initDbBtn');
        }
    }

    async addTeam() {
        const teamName = document.getElementById('teamName').value.trim();
        const teamStrength = parseInt(document.getElementById('teamStrength').value);

        if (!teamName) {
            this.showMessage('Please enter a team name', 'error');
            return;
        }

        if (isNaN(teamStrength) || teamStrength < 1 || teamStrength > 100) {
            this.showMessage('Please enter a valid strength (1-100)', 'error');
            return;
        }

        try {
            this.updateStatus('Adding team...');
            this.showLoading('addTeamBtn');

            const response = await fetch(`${this.apiBase}/teams`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: teamName,
                    strength: teamStrength
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            
            this.updateStatus('Team added successfully!');
            this.showMessage(`Team ${teamName} added successfully!`, 'success');
            
            // Clear form
            document.getElementById('teamName').value = '';
            document.getElementById('teamStrength').value = '';
            
            // Reload teams list
            await this.loadTeams();
            
        } catch (error) {
            console.error('Error adding team:', error);
            this.updateStatus('Failed to add team');
            this.showMessage('Failed to add team: ' + error.message, 'error');
        } finally {
            this.hideLoading('addTeamBtn');
        }
    }

    async loadTeams() {
        try {
            const response = await fetch(`${this.apiBase}/teams`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            this.updateTeamsList(data.teams);
            this.updateTeamCount(data.count);
        } catch (error) {
            console.error('Error loading teams:', error);
            this.showMessage('Failed to load teams: ' + error.message, 'error');
        }
    }

    updateTeamsList(teams) {
        const teamsList = document.getElementById('teamsList');
        
        if (!teams || teams.length === 0) {
            teamsList.innerHTML = '<div class="no-data">No teams available</div>';
            return;
        }
        
        teamsList.innerHTML = teams.map(team => `
            <div class="team-card">
                <h3>${team.name}</h3>
                <p><strong>ID:</strong> ${team.id}</p>
                <p class="team-strength"><strong>Strength:</strong> ${team.strength}/100</p>
                <div class="team-actions">
                    <button class="btn btn-small btn-secondary edit-team-btn" data-team-id="${team.id}">Edit</button>
                    <button class="btn btn-small btn-danger delete-team-btn" data-team-id="${team.id}">Delete</button>
                </div>
            </div>
        `).join('');
        
        // Add event listeners to the new buttons
        this.addTeamButtonListeners();
    }

    addTeamButtonListeners() {
        // Add event listeners for edit buttons
        document.querySelectorAll('.edit-team-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const teamId = parseInt(e.target.getAttribute('data-team-id'));
                this.editTeam(teamId);
            });
        });
        
        // Add event listeners for delete buttons
        document.querySelectorAll('.delete-team-btn').forEach(button => {
            button.addEventListener('click', (e) => {
                const teamId = parseInt(e.target.getAttribute('data-team-id'));
                this.deleteTeam(teamId);
            });
        });
    }

    updateTeamCount(count) {
        document.getElementById('teamCount').textContent = count;
    }



    async clearTeams() {
        if (!confirm('Are you sure you want to clear all teams? This action cannot be undone.')) {
            return;
        }
        
        try {
            this.updateStatus('Clearing all teams...');
            this.showLoading('clearTeamsBtn');
            
            const response = await fetch(`${this.apiBase}/clear-teams`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            this.updateStatus('All teams cleared successfully!');
            this.showMessage('All teams have been removed from the database', 'success');
            
            // Clear teams list
            this.updateTeamsList([]);
            this.updateTeamCount(0);
            
        } catch (error) {
            console.error('Error clearing teams:', error);
            this.updateStatus('Failed to clear teams');
            this.showMessage('Failed to clear teams: ' + error.message, 'error');
        } finally {
            this.hideLoading('clearTeamsBtn');
        }
    }

    async editTeam(teamId) {
        // For now, we'll show a simple prompt. In a real app, you'd have a modal
        const newName = prompt('Enter new team name:');
        if (!newName) return;
        
        const newStrength = prompt('Enter new team strength (1-100):');
        if (!newStrength) return;
        
        const strength = parseInt(newStrength);
        if (isNaN(strength) || strength < 1 || strength > 100) {
            this.showMessage('Invalid strength value', 'error');
            return;
        }
        
        try {
            this.updateStatus('Updating team...');
            
            const response = await fetch(`${this.apiBase}/teams/${teamId}`, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    name: newName,
                    strength: strength
                })
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            this.updateStatus('Team updated successfully!');
            this.showMessage(`Team ${newName} updated successfully!`, 'success');
            
            // Reload teams list
            await this.loadTeams();
            
        } catch (error) {
            console.error('Error updating team:', error);
            this.updateStatus('Failed to update team');
            this.showMessage('Failed to update team: ' + error.message, 'error');
        }
    }

    async deleteTeam(teamId) {
        if (!confirm('Are you sure you want to delete this team?')) {
            return;
        }
        
        try {
            this.updateStatus('Deleting team...');
            
            const response = await fetch(`${this.apiBase}/teams/${teamId}`, {
                method: 'DELETE'
            });
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            
            const data = await response.json();
            
            this.updateStatus('Team deleted successfully!');
            this.showMessage('Team deleted successfully!', 'success');
            
            // Reload teams list
            await this.loadTeams();
            
        } catch (error) {
            console.error('Error deleting team:', error);
            this.updateStatus('Failed to delete team');
            this.showMessage('Failed to delete team: ' + error.message, 'error');
        }
    }

    async loadMatchSchedule() {
        try {
            this.updateStatus('Loading match schedule...');
            this.showLoading('loadScheduleBtn');

            const response = await fetch(`${this.apiBase}/league/schedule`);
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const schedule = await response.json();
            
            if (Object.keys(schedule).length === 0) {
                this.showMessage('No schedule available. Please click "Create League" first to generate fixtures!', 'warning');
                this.updateScheduleDisplay({});
                return;
            }
            
            this.scheduleData = schedule;
            
            this.updateStatus('Match schedule loaded!');
            this.showMessage('Match schedule loaded successfully!', 'success');
            
            this.updateScheduleDisplay(schedule);
            this.updateWeekFilter(schedule);
            
        } catch (error) {
            console.error('Error loading match schedule:', error);
            this.updateStatus('Failed to load match schedule');
            this.showMessage('Failed to load match schedule: ' + error.message, 'error');
        } finally {
            this.hideLoading('loadScheduleBtn');
        }
    }

    updateScheduleDisplay(schedule) {
        const container = document.getElementById('scheduleContainer');
        
        if (!schedule || Object.keys(schedule).length === 0) {
            container.innerHTML = '<div class="no-data">No schedule available. Click "Create League" to generate fixtures first!</div>';
            return;
        }

        const weeks = Object.keys(schedule).sort((a, b) => parseInt(a) - parseInt(b));
        
        container.innerHTML = weeks.map(week => {
            const matches = schedule[week];
            return `
                <div class="week-schedule" data-week="${week}">
                    <div class="week-header">Week ${week}</div>
                    ${matches.map(match => `
                        <div class="schedule-match upcoming">
                            <div class="match-teams">${match.HomeTeam} vs ${match.AwayTeam}</div>
                            <div class="match-status upcoming">Upcoming</div>
                        </div>
                    `).join('')}
                </div>
            `;
        }).join('');
    }

    updateWeekFilter(schedule) {
        const filter = document.getElementById('weekFilter');
        const weeks = Object.keys(schedule).sort((a, b) => parseInt(a) - parseInt(b));
        
        // Clear existing options except "All Weeks"
        filter.innerHTML = '<option value="all">All Weeks</option>';
        
        // Add week options
        weeks.forEach(week => {
            const option = document.createElement('option');
            option.value = week;
            option.textContent = `Week ${week}`;
            filter.appendChild(option);
        });
    }

    filterSchedule() {
        const selectedWeek = document.getElementById('weekFilter').value;
        const weekSchedules = document.querySelectorAll('.week-schedule');
        
        weekSchedules.forEach(weekSchedule => {
            const week = weekSchedule.getAttribute('data-week');
            if (selectedWeek === 'all' || week === selectedWeek) {
                weekSchedule.style.display = 'block';
            } else {
                weekSchedule.style.display = 'none';
            }
        });
    }

    async createLeague() {
        try {
            this.updateStatus('Creating league...');
            this.showLoading('createLeagueBtn');

            const response = await fetch(`${this.apiBase}/league`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify([])
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            this.isLeagueCreated = true;
            
            // Extract total weeks from response
            this.totalWeeks = data.total_weeks || 0;
            this.currentWeek = data.current_week || 0;
            
            this.updateStatus('League created successfully!');
            this.showMessage('League created successfully!', 'success');
            
            // Load initial data
            await this.loadLeagueStatus();
            await this.loadLeagueTable();
            await this.loadMatches();
            await this.loadPredictions();
            
        } catch (error) {
            console.error('Error creating league:', error);
            this.updateStatus('Failed to create league');
            this.showMessage('Failed to create league: ' + error.message, 'error');
        } finally {
            this.hideLoading('createLeagueBtn');
        }
    }

    async clearLeague() {
        if (!confirm('Are you sure you want to clear the league? This will delete all data including teams.')) {
            return;
        }

        try {
            this.updateStatus('Clearing league...');
            this.showLoading('clearLeagueBtn');

            const response = await fetch(`${this.apiBase}/league/clear`, {
                method: 'DELETE'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            this.isLeagueCreated = false;
            this.currentWeek = 0;
            this.totalWeeks = 0;
            
            this.updateStatus('League cleared successfully!');
            this.showMessage('League cleared successfully! All teams and data have been removed.', 'success');
            
            // Clear UI
            this.clearLeagueTable();
            this.clearMatches();
            this.clearTeamsList();
            this.updateWeekInfo();
            this.updatePredictions([]);
            
        } catch (error) {
            console.error('Error clearing league:', error);
            this.updateStatus('Failed to clear league');
            this.showMessage('Failed to clear league: ' + error.message, 'error');
        } finally {
            this.hideLoading('clearLeagueBtn');
        }
    }

    async playNextWeek() {
        if (!this.isLeagueCreated) {
            this.showMessage('Please create a league first', 'error');
            return;
        }

        try {
            this.updateStatus('Playing next week...');
            this.showLoading('nextWeekBtn');

            const response = await fetch(`${this.apiBase}/league/play-week`, {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            
            // Extract current week from response
            this.currentWeek = data.current_week || 0;
            
            this.updateStatus('Week played successfully!');
            this.showMessage(`Week ${this.currentWeek} played successfully!`, 'success');
            
            // Refresh data
            await this.loadLeagueStatus();
            await this.loadLeagueTable();
            await this.loadMatches();
            await this.loadPredictions();
            
        } catch (error) {
            console.error('Error playing week:', error);
            this.updateStatus('Failed to play week');
            this.showMessage('Failed to play week: ' + error.message, 'error');
        } finally {
            this.hideLoading('nextWeekBtn');
        }
    }

    async playAllWeeks() {
        if (!this.isLeagueCreated) {
            this.showMessage('Please create a league first', 'error');
            return;
        }

        try {
            this.updateStatus('Playing all remaining weeks...');
            this.showLoading('playAllBtn');

            const response = await fetch(`${this.apiBase}/league/play-all`, {
                method: 'POST'
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            
            // Extract current week from response (should be the final week)
            if (data.matches_by_week) {
                const weeks = Object.keys(data.matches_by_week).map(Number);
                this.currentWeek = Math.max(...weeks);
            }
            
            this.updateStatus('All weeks played successfully!');
            this.showMessage('All remaining weeks played successfully!', 'success');
            
            // Refresh data
            await this.loadLeagueStatus();
            await this.loadLeagueTable();
            await this.loadMatches();
            await this.loadPredictions();
            
        } catch (error) {
            console.error('Error playing all weeks:', error);
            this.updateStatus('Failed to play all weeks');
            this.showMessage('Failed to play all weeks: ' + error.message, 'error');
        } finally {
            this.hideLoading('playAllBtn');
        }
    }

    async loadLeagueStatus() {
        try {
            const response = await fetch(`${this.apiBase}/league/status`);
            if (response.ok) {
                const data = await response.json();
                console.log('League status response:', data);
                this.currentWeek = data.current_week || 0;
                this.totalWeeks = data.total_weeks || 0;
                console.log('Updated league status - Current Week:', this.currentWeek, 'Total Weeks:', this.totalWeeks);
                this.updateWeekInfo();
            }
        } catch (error) {
            console.error('Error loading league status:', error);
        }
    }

    async loadLeagueTable() {
        try {
            const response = await fetch(`${this.apiBase}/league/table`);
            if (response.ok) {
                const standings = await response.json();
                this.updateLeagueTable(standings);
                
                // Also load matches and predictions
                await this.loadMatches();
                await this.loadPredictions();
            }
        } catch (error) {
            console.error('Error loading league table:', error);
        }
    }

    async loadMatches() {
        try {
            const response = await fetch(`${this.apiBase}/league/matches`);
            if (response.ok) {
                const data = await response.json();
                // Backend returns matches array directly, not wrapped in an object
                this.updateMatches(data || []);
            }
        } catch (error) {
            console.error('Error loading matches:', error);
        }
    }

    async loadPredictions() {
        try {
            const response = await fetch(`${this.apiBase}/league/predictions`);
            if (response.ok) {
                const predictions = await response.json();
                this.updatePredictionsFromBackend(predictions);
            }
        } catch (error) {
            console.error('Error loading predictions:', error);
            // Fallback to empty predictions
            this.updatePredictionsFromBackend([]);
        }
    }

    updateLeagueTable(standings) {
        const tbody = document.getElementById('leagueTableBody');
        
        if (!standings || standings.length === 0) {
            tbody.innerHTML = '<tr><td colspan="8" class="no-data">No league data available</td></tr>';
            return;
        }

        // Debug logging
        console.log('League table data received:', standings);
        standings.forEach((team, index) => {
            console.log(`Team ${index + 1}: ${team.team_name} - GoalsFor: ${team.goals_for}, GoalDiff: ${team.goal_difference}`);
        });

        tbody.innerHTML = standings.map((team, index) => `
            <tr>
                <td class="team-name">${index + 1}. ${team.team_name}</td>
                <td class="points">${team.points || 0}</td>
                <td>${team.played || 0}</td>
                <td>${team.won || 0}</td>
                <td>${team.drawn || 0}</td>
                <td>${team.lost || 0}</td>
                <td class="goal-diff ${(team.goal_difference || 0) >= 0 ? 'positive' : 'negative'}">${(team.goal_difference || 0) >= 0 ? '+' : ''}${team.goal_difference || 0}</td>
                <td>${team.goals_for || 0}</td>
            </tr>
        `).join('');
    }

    updateMatches(matches) {
        const container = document.getElementById('matchesContainer');
        
        if (!matches || matches.length === 0) {
            container.innerHTML = '<div class="no-data">No matches played yet</div>';
            return;
        }

        // Group matches by week
        const matchesByWeek = {};
        matches.forEach(match => {
            if (!matchesByWeek[match.week]) {
                matchesByWeek[match.week] = [];
            }
            matchesByWeek[match.week].push(match);
        });

        // Sort weeks in ascending order (week 1, 2, 3, etc.)
        const weeks = Object.keys(matchesByWeek).sort((a, b) => parseInt(a) - parseInt(b));
        
        // Build HTML for all weeks
        let allWeeksHTML = '';
        
        weeks.forEach(week => {
            const weekMatches = matchesByWeek[week];
            const weekNumber = parseInt(week);
            
            allWeeksHTML += `
                <div class="week-results">
                    <div class="week-header">
                        <h3>${weekNumber}${this.getOrdinalSuffix(weekNumber)} Week Results</h3>
                    </div>
                    ${weekMatches.map(match => `
                        <div class="match">
                            <div class="match-teams">${match.home_team} vs ${match.away_team}</div>
                            <div class="match-score">${match.home_score} - ${match.away_score}</div>
                        </div>
                    `).join('')}
                </div>
            `;
        });

        container.innerHTML = allWeeksHTML;

        // Update week info with match count
        this.updateWeekInfoWithMatches(matches);

        // Scroll to top of matches container to show latest results
        container.scrollTop = 0;

        // Update predictions after showing matches
        this.updatePredictions(matches);
    }

    getOrdinalSuffix(num) {
        const j = num % 10;
        const k = num % 100;
        if (j == 1 && k != 11) {
            return "st";
        }
        if (j == 2 && k != 12) {
            return "nd";
        }
        if (j == 3 && k != 13) {
            return "rd";
        }
        return "th";
    }

    updatePredictions(matches) {
        const container = document.getElementById('predictionsContainer');
        
        if (!matches || matches.length === 0) {
            container.innerHTML = '<div class="no-data">No predictions available</div>';
            return;
        }

        // Calculate predictions based on current standings
        const predictions = this.calculatePredictions(matches);
        
        container.innerHTML = predictions.map(pred => `
            <div class="prediction-item">
                <div class="prediction-team">${pred.team}</div>
                <div class="prediction-percentage">%${pred.percentage}</div>
            </div>
        `).join('');
    }

    updatePredictionsFromBackend(predictions) {
        const container = document.getElementById('predictionsContainer');
        
        if (!predictions || predictions.length === 0) {
            container.innerHTML = '<div class="no-data">No predictions available</div>';
            return;
        }
        
        container.innerHTML = predictions.map(pred => `
            <div class="prediction-item">
                <div class="prediction-team">${pred.team_name}</div>
                <div class="prediction-percentage">%${pred.percentage.toFixed(1)}</div>
            </div>
        `).join('');
    }

    calculatePredictions(matches) {
        // Simple prediction algorithm based on current performance
        // In a real app, you'd use more sophisticated algorithms
        
        // Get current standings
        const standings = this.getCurrentStandings(matches);
        
        if (standings.length === 0) {
            return [];
        }

        // Calculate total points
        const totalPoints = standings.reduce((sum, team) => sum + team.points, 0);
        
        // Calculate percentages based on points
        const predictions = standings.map(team => {
            const percentage = totalPoints > 0 ? Math.round((team.points / totalPoints) * 100) : 0;
            return {
                team: team.team_name,
                percentage: percentage
            };
        });

        // Sort by percentage (highest first)
        return predictions.sort((a, b) => b.percentage - a.percentage);
    }

    getCurrentStandings(matches) {
        // This is a simplified version - in reality, you'd get this from the API
        // For now, we'll return a mock standings based on matches
        const teams = {};
        
        matches.forEach(match => {
            // Initialize teams if not exists
            if (!teams[match.home_team]) {
                teams[match.home_team] = { team_name: match.home_team, points: 0, played: 0 };
            }
            if (!teams[match.away_team]) {
                teams[match.away_team] = { team_name: match.away_team, points: 0, played: 0 };
            }

            // Calculate points
            if (match.home_score > match.away_score) {
                teams[match.home_team].points += 3;
            } else if (match.home_score < match.away_score) {
                teams[match.away_team].points += 3;
            } else {
                teams[match.home_team].points += 1;
                teams[match.away_team].points += 1;
            }

            teams[match.home_team].played++;
            teams[match.away_team].played++;
        });

        return Object.values(teams).sort((a, b) => b.points - a.points);
    }

    clearLeagueTable() {
        const tbody = document.getElementById('leagueTableBody');
        tbody.innerHTML = '<tr><td colspan="7" class="no-data">No league data available</td></tr>';
    }

    clearMatches() {
        const container = document.getElementById('matchesContainer');
        container.innerHTML = '<div class="no-data">No matches played yet</div>';
    }

    clearTeamsList() {
        const container = document.getElementById('teamsList');
        container.innerHTML = '<div class="no-data">No teams available</div>';
    }

    updateWeekInfo() {
        document.getElementById('currentWeek').textContent = `Week: ${this.currentWeek}`;
        document.getElementById('totalWeeks').textContent = `/ ${this.totalWeeks}`;
    }

    updateWeekInfoWithMatches(matches) {
        // Calculate current week from matches data
        let currentWeek = 0;
        if (matches && matches.length > 0) {
            // Find the highest week number from matches
            currentWeek = Math.max(...matches.map(match => match.week));
            console.log('Calculated current week from matches:', currentWeek, 'Total matches:', matches.length);
        } else {
            // Fallback to stored current week if no matches
            currentWeek = this.currentWeek || 0;
            console.log('Using stored current week:', currentWeek);
        }
        
        console.log('Updating week info - Current Week:', currentWeek, 'Total Weeks:', this.totalWeeks);
        
        document.getElementById('currentWeek').textContent = `Week: ${currentWeek}`;
        document.getElementById('totalWeeks').textContent = `/ ${this.totalWeeks}`;
        
        // Add total matches info if we have matches
        const weekInfoElement = document.querySelector('.week-info');
        if (matches && matches.length > 0) {
            const totalMatchesSpan = document.getElementById('totalMatches');
            if (!totalMatchesSpan) {
                const span = document.createElement('span');
                span.id = 'totalMatches';
                span.style.marginLeft = '10px';
                span.style.color = '#666';
                span.style.fontSize = '0.9em';
                weekInfoElement.appendChild(span);
            }
            document.getElementById('totalMatches').textContent = `(${matches.length} matches played)`;
        } else if (document.getElementById('totalMatches')) {
            document.getElementById('totalMatches').textContent = '';
        }
    }

    updateStatus(message) {
        document.getElementById('statusMessage').textContent = message;
    }

    showMessage(message, type) {
        // Remove existing messages
        const existingMessages = document.querySelectorAll('.message');
        existingMessages.forEach(msg => msg.remove());

        // Create new message
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;
        messageDiv.textContent = message;

        // Insert after header
        const header = document.querySelector('header');
        header.parentNode.insertBefore(messageDiv, header.nextSibling);

        // Auto-remove after 5 seconds
        setTimeout(() => {
            if (messageDiv.parentNode) {
                messageDiv.remove();
            }
        }, 5000);
    }

    showLoading(buttonId) {
        const button = document.getElementById(buttonId);
        button.disabled = true;
        button.textContent = 'Loading...';
    }

    hideLoading(buttonId) {
        const button = document.getElementById(buttonId);
        button.disabled = false;
        
        // Restore original text
        const originalTexts = {
            'initDbBtn': 'Initialize DB',
            'createLeagueBtn': 'Create League',
            'clearLeagueBtn': 'Clear League',
            'nextWeekBtn': 'Next Week',
            'playAllBtn': 'Play All',
            'addTeamBtn': 'Add Team',

            'loadTeamsBtn': 'Refresh Teams',
            'clearTeamsBtn': 'Clear All Teams',
            'loadScheduleBtn': 'Load Schedule'
        };
        button.textContent = originalTexts[buttonId] || 'Button';
    }
}

// Initialize the app when the page loads
document.addEventListener('DOMContentLoaded', () => {
    // Create global app instance
    window.app = new FootballLeagueApp();
    // Also keep the footballApp name for backward compatibility
    window.footballApp = window.app;
    // Load teams on page load
    window.app.loadTeams();
}); 