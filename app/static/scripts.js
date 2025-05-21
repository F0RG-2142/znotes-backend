// Global State
const state = {
    user: null,
    token: null,
    refreshToken: null,
    currentNote: null,
    currentTeam: null,
    currentTeamNote: null,
    personalNotes: [],
    teams: [],
    teamNotes: [],
    teamMembers: []
};

// Auth Functions
async function login(email, password) {
    try {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Login failed');
        }
        
        const userData = await response.json();
        
        // Save auth data
        state.user = {
            id: userData.id,
            email: userData.email,
            hasPremium: userData.has_notes_premium
        };
        state.token = userData.token;
        state.refreshToken = userData.refresh_token;
        
        // Save tokens to localStorage
        localStorage.setItem('token', userData.token);
        localStorage.setItem('refreshToken', userData.refresh_token);
        localStorage.setItem('userId', userData.id);
        
        // Update UI
        showMainContent();
        updateUserInfo();
        
        // Load user data
        await loadUserData();
    } catch (error) {
        console.error('Login error:', error);
        elements.loginError.textContent = error.message;
    }
}

async function register(email, password) {
    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Registration failed');
        }
        
        // Switch to login tab after successful registration
        elements.authTabs[0].click();
        elements.loginError.textContent = 'Registration successful. Please login.';
        elements.loginError.style.color = 'green';
    } catch (error) {
        console.error('Registration error:', error);
        elements.registerError.textContent = error.message;
    }
}

async function logout() {
    try {
        await fetch(`${API_BASE_URL}/logout`, {
            method: 'POST',
            headers: {
                'Authorization': state.token
            }
        });
    } catch (error) {
        console.error('Logout error:', error);
    } finally {
        // Clear state and localStorage
        state.user = null;
        state.token = null;
        state.refreshToken = null;
        state.currentNote = null;
        state.currentTeam = null;
        state.personalNotes = [];
        state.teams = [];
        
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('userId');
        
        // Show auth section
        showAuthSection();
    }
}

async function refreshAccessToken() {
    try {
        const response = await fetch(`${API_BASE_URL}/token/refresh`, {
            method: 'POST',
            headers: {
                'Authorization': state.refreshToken
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to refresh token');
        }
        
        const data = await response.json();
        state.token = data.token;
        localStorage.setItem('token', data.token);
        
        return true;
    } catch (error) {
        console.error('Token refresh error:', error);
        logout(); // Force logout if token refresh fails
        return false;
    }
}

// API Request Helper
async function apiRequest(url, method = 'GET', body = null, needsAuth = true) {
    const headers = {
        'Content-Type': 'application/json'
    };
    
    if (needsAuth && state.token) {
        headers['Authorization'] = state.token;
    }
    
    const options = {
        method,
        headers
    };
    
    if (body && (method === 'POST' || method === 'PUT')) {
        options.body = JSON.stringify(body);
    }
    
    try {
        let response = await fetch(`${API_BASE_URL}${url}`, options);
        
        // Handle token expiration
        if (response.status === 401 && state.refreshToken) {
            const refreshed = await refreshAccessToken();
            if (refreshed) {
                // Retry request with new token
                headers['Authorization'] = state.token;
                options.headers = headers;
                response = await fetch(`${API_BASE_URL}${url}`, options);
            }
        }
        
        if (!response.ok) {
            if (response.headers.get('Content-Type')?.includes('application/json')) {
                const errorData = await response.json();
                throw new Error(errorData.error || `Request failed with status ${response.status}`);
            } else {
                throw new Error(`Request failed with status ${response.status}`);
            }
        }
        
        if (response.status === 204) {
            return null; // No content
        }
        
        if (response.headers.get('Content-Type')?.includes('application/json')) {
            return await response.json();
        }
        
        return null;
    } catch (error) {
        console.error(`API Request Error (${url}):`, error);
        throw error;
    }
}

// Notes Functions
async function fetchPersonalNotes() {
    try {
        const notes = await apiRequest(`/notes?authorId=${state.user.id}`);
        state.personalNotes = notes || [];
        renderPersonalNotesList();
    } catch (error) {
        console.error('Fetch notes error:', error);
    }
}

async function createNote(body) {
    try {
        await apiRequest('/notes', 'POST', {
            body,
            user_id: state.user.id
        });
        
        // Refresh notes list
        await fetchPersonalNotes();
        
        // Clear editor
        resetNoteEditor();
    } catch (error) {
        console.error('Create note error:', error);
        alert(`Failed to create note: ${error.message}`);
    }
}

async function updateNote(noteId, body) {
    try {
        await apiRequest(`/notes/${noteId}`, 'PUT', {
            noteID: noteId,
            body
        });
        
        // Update local state
        const noteIndex = state.personalNotes.findIndex(note => note.id === noteId);
        if (noteIndex !== -1) {
            state.personalNotes[noteIndex].body = body;
            state.personalNotes[noteIndex].updated_at = new Date().toISOString();
        }
        
        // Refresh notes list
        renderPersonalNotesList();
    } catch (error) {
        console.error('Update note error:', error);
        alert(`Failed to update note: ${error.message}`);
    }
}

// Event Listeners
function setupEventListeners() {
    // Auth tabs
    elements.authTabs.forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.tab;
            
            // Update active tab
            elements.authTabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            
            // Show corresponding form
            elements.authForms.forEach(form => {
                form.classList.remove('active');
                if (form.id === `${tabName}-form`) {
                    form.classList.add('active');
                }
            });
            
            // Clear error messages
            elements.loginError.textContent = '';
            elements.registerError.textContent = '';
        });
    });
    
    // Login form
    elements.loginForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const email = document.getElementById('login-email').value;
        const password = document.getElementById('login-password').value;
        login(email, password);
    });
    
    // Register form
    elements.registerForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const email = document.getElementById('register-email').value;
        const password = document.getElementById('register-password').value;
        register(email, password);
    });
    
    // Logout button
    elements.logoutBtn.addEventListener('click', logout);
    
    // Create note
    elements.createNoteBtn.addEventListener('click', () => {
        resetNoteEditor();
    });
    
    // Save note
    elements.saveNoteBtn.addEventListener('click', () => {
        const noteText = elements.noteEditor.value.trim();
        
        if (!noteText) {
            alert('Note cannot be empty');
            return;
        }
        
        if (state.currentNote) {
            // Update existing note
            updateNote(state.currentNote.id, noteText);
        } else {
            // Create new note
            createNote(noteText);
        }
    });
    
    // Delete note
    elements.deleteNoteBtn.addEventListener('click', () => {
        if (!state.currentNote) {
            alert('No note selected');
            return;
        }
        
        deleteNote(state.currentNote.id);
    });
    
    // Create team button
    elements.createTeamBtn.addEventListener('click', () => {
        elements.createTeamForm.reset();
        showModal(elements.createTeamModal);
    });
    
    // Create team form
    elements.createTeamForm.addEventListener('submit', (e) => {
        e.preventDefault();
        const teamName = elements.teamNameInput.value.trim();
        const isPrivate = elements.teamPrivateCheckbox.checked;
        
        if (!teamName) {
            alert('Team name cannot be empty');
            return;
        }
        
        createTeam(teamName, isPrivate);
    });
    
    // Create team note button
    elements.createTeamNoteBtn.addEventListener('click', () => {
        if (!state.currentTeam) {
            alert('No team selected');
            return;
        }
        
        resetTeamNoteEditor();
        elements.teamNoteEditorContainer.style.display = 'flex';
    });
    
    // Save team note
    elements.saveTeamNoteBtn.addEventListener('click', () => {
        if (!state.currentTeam) {
            alert('No team selected');
            return;
        }
        
        const noteText = elements.teamNoteEditor.value.trim();
        
        if (!noteText) {
            alert('Note cannot be empty');
            return;
        }
        
        if (state.currentTeamNote) {
            // Update existing team note
            updateTeamNote(state.currentTeam.team_id, state.currentTeamNote.note_id, noteText);
        } else {
            // Create new team note
            createTeamNote(state.currentTeam.team_id, noteText);
        }
    });
    
    // Delete team note
    elements.deleteTeamNoteBtn.addEventListener('click', () => {
        if (!state.currentTeam || !state.currentTeamNote) {
            alert('No team note selected');
            return;
        }
        
        deleteTeamNote(state.currentTeam.team_id, state.currentTeamNote.note_id);
    });
    
    // Manage team button
    elements.manageTeamBtn.addEventListener('click', () => {
        if (!state.currentTeam) {
            alert('No team selected');
            return;
        }
        
        // Fetch team members
        fetchTeamMembers(state.currentTeam.team_id);
        
        // Reset form
        elements.addMemberForm.reset();
        
        // Show modal
        showModal(elements.manageTeamModal);
    });
    
    // Add member form
    elements.addMemberForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        
        if (!state.currentTeam) {
            alert('No team selected');
            return;
        }
        
        const email = elements.memberEmailInput.value.trim();
        const role = elements.memberRoleInput.value;
        
        if (!email) {
            alert('Email cannot be empty');
            return;
        }
        
        // TODO: Add proper user lookup by email
        // For now, we'll just use the email as the user ID
        addTeamMember(state.currentTeam.team_id, email, role);
    });
    
    // Close modal buttons
    elements.closeModalBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            const modal = btn.closest('.modal');
            closeModal(modal);
        });
    });
    
    // Close modal when clicking outside
    elements.modalOverlay.addEventListener('click', (e) => {
        if (e.target === elements.modalOverlay) {
            document.querySelectorAll('.modal').forEach(modal => {
                closeModal(modal);
            });
        }
    });
}

// Initialize Application
function init() {
    setupEventListeners();
    checkAuthStatus();
}

// Start the application
document.addEventListener('DOMContentLoaded', init);// UI Functions
function showAuthSection() {
    elements.authSection.style.display = 'flex';
    elements.mainContent.style.display = 'none';
    elements.logoutBtn.style.display = 'none';
}

function showMainContent() {
    elements.authSection.style.display = 'none';
    elements.mainContent.style.display = 'flex';
    elements.logoutBtn.style.display = 'block';
}

function updateUserInfo() {
    if (state.user) {
        elements.userInfo.textContent = state.user.email;
    } else {
        elements.userInfo.textContent = 'Not logged in';
    }
}

function renderPersonalNotesList() {
    elements.personalNotesList.innerHTML = '';
    
    if (state.personalNotes.length === 0) {
        const emptyItem = document.createElement('li');
        emptyItem.textContent = 'No notes yet';
        emptyItem.classList.add('empty-list-message');
        elements.personalNotesList.appendChild(emptyItem);
        return;
    }
    
    state.personalNotes.forEach(note => {
        const listItem = document.createElement('li');
        listItem.classList.add('note-item');
        
        if (state.currentNote && state.currentNote.id === note.id) {
            listItem.classList.add('active');
        }
        
        // Create preview title from first line of note body
        const previewTitle = note.body.split('\n')[0] || 'Untitled Note';
        
        // Format date
        const date = new Date(note.updated_at);
        const formattedDate = `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
        
        listItem.innerHTML = `
            <div class="note-info">
                <div class="note-title">${previewTitle}</div>
                <div class="note-date">Updated: ${formattedDate}</div>
            </div>
        `;
        
        listItem.addEventListener('click', () => {
            selectNote(note);
        });
        
        elements.personalNotesList.appendChild(listItem);
    });
}

function renderTeamsList() {
    elements.teamsList.innerHTML = '';
    
    if (state.teams.length === 0) {
        const emptyItem = document.createElement('li');
        emptyItem.textContent = 'No teams yet';
        emptyItem.classList.add('empty-list-message');
        elements.teamsList.appendChild(emptyItem);
        return;
    }
    
    state.teams.forEach(team => {
        const listItem = document.createElement('li');
        listItem.classList.add('team-item');
        
        if (state.currentTeam && state.currentTeam.team_id === team.team_id) {
            listItem.classList.add('active');
        }
        
        // Format date
        const date = new Date(team.created_at);
        const formattedDate = date.toLocaleDateString();
        
        listItem.innerHTML = `
            <div class="team-info">
                <div class="team-title">${team.team_name}</div>
                <div class="team-date">Created: ${formattedDate}</div>
            </div>
        `;
        
        listItem.addEventListener('click', () => {
            selectTeam(team);
        });
        
        elements.teamsList.appendChild(listItem);
    });
}

function renderTeamNotesList() {
    elements.teamNotesList.innerHTML = '';
    
    if (!state.teamNotes || state.teamNotes.length === 0) {
        const emptyItem = document.createElement('li');
        emptyItem.textContent = 'No team notes yet';
        emptyItem.classList.add('empty-list-message');
        elements.teamNotesList.appendChild(emptyItem);
        return;
    }
    
    state.teamNotes.forEach(note => {
        const listItem = document.createElement('li');
        listItem.classList.add('note-item');
        
        if (state.currentTeamNote && state.currentTeamNote.note_id === note.note_id) {
            listItem.classList.add('active');
        }
        
        // Create preview title from first line of note body
        const previewTitle = note.body.split('\n')[0] || 'Untitled Note';
        
        // Format date
        const date = new Date(note.updated_at);
        const formattedDate = `${date.toLocaleDateString()} ${date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
        
        listItem.innerHTML = `
            <div class="note-info">
                <div class="note-title">${previewTitle}</div>
                <div class="note-date">Updated: ${formattedDate}</div>
            </div>
        `;
        
        listItem.addEventListener('click', () => {
            selectTeamNote(note);
        });
        
        elements.teamNotesList.appendChild(listItem);
    });
}

function renderTeamMembersList() {
    elements.teamMembersList.innerHTML = '';
    
    if (!state.teamMembers || state.teamMembers.length === 0) {
        const emptyItem = document.createElement('li');
        emptyItem.textContent = 'No members yet';
        emptyItem.classList.add('empty-list-message');
        elements.teamMembersList.appendChild(emptyItem);
        return;
    }
    
    state.teamMembers.forEach(member => {
        const listItem = document.createElement('li');
        listItem.classList.add('team-member-item');
        
        listItem.innerHTML = `
            <div class="member-info">
                <div class="member-email">${member.user_id}</div>
                <div class="member-role">${member.role}</div>
            </div>
            <button class="remove-member-btn" data-member-id="${member.user_id}">Remove</button>
        `;
        
        // Add event listener for remove button
        const removeBtn = listItem.querySelector('.remove-member-btn');
        removeBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            const memberId = removeBtn.dataset.memberId;
            removeTeamMember(state.currentTeam.team_id, memberId);
        });
        
        elements.teamMembersList.appendChild(listItem);
    });
}

function selectNote(note) {
    state.currentNote = note;
    elements.noteEditor.value = note.body;
    elements.currentNoteInfo.textContent = `Editing Note (${new Date(note.updated_at).toLocaleString()})`;
    
    // Update active class in list
    const noteItems = elements.personalNotesList.querySelectorAll('.note-item');
    noteItems.forEach(item => item.classList.remove('active'));
    
    const selectedItem = Array.from(noteItems).find(item => {
        const titleEl = item.querySelector('.note-title');
        return titleEl && titleEl.textContent === note.body.split('\n')[0];
    });
    
    if (selectedItem) {
        selectedItem.classList.add('active');
    }
}

function selectTeam(team) {
    state.currentTeam = team;
    elements.teamName.textContent = team.team_name;
    elements.teamView.style.display = 'block';
    
    // Fetch team notes
    fetchTeamNotes(team.team_id);
    
    // Update active class in list
    const teamItems = elements.teamsList.querySelectorAll('.team-item');
    teamItems.forEach(item => item.classList.remove('active'));
    
    const selectedItem = Array.from(teamItems).find(item => {
        const titleEl = item.querySelector('.team-title');
        return titleEl && titleEl.textContent === team.team_name;
    });
    
    if (selectedItem) {
        selectedItem.classList.add('active');
    }
}

function selectTeamNote(note) {
    state.currentTeamNote = note;
    elements.teamNoteEditor.value = note.body;
    elements.currentTeamNoteInfo.textContent = `Editing Team Note (${new Date(note.updated_at).toLocaleString()})`;
    elements.teamNoteEditorContainer.style.display = 'flex';
    
    // Update active class in list
    const noteItems = elements.teamNotesList.querySelectorAll('.note-item');
    noteItems.forEach(item => item.classList.remove('active'));
    
    const selectedItem = Array.from(noteItems).find(item => {
        const titleEl = item.querySelector('.note-title');
        return titleEl && titleEl.textContent === note.body.split('\n')[0];
    });
    
    if (selectedItem) {
        selectedItem.classList.add('active');
    }
}

function resetNoteEditor() {
    state.currentNote = null;
    elements.noteEditor.value = '';
    elements.currentNoteInfo.textContent = 'New Note';
}

function resetTeamNoteEditor() {
    state.currentTeamNote = null;
    elements.teamNoteEditor.value = '';
    elements.currentTeamNoteInfo.textContent = 'New Team Note';
}

function showNotesView() {
    elements.teamView.style.display = 'none';
}

// Modal Functions
function showModal(modal) {
    elements.modalOverlay.style.display = 'flex';
    modal.style.display = 'block';
    modal.classList.add('active');
}

function closeModal(modal) {
    elements.modalOverlay.style.display = 'none';
    modal.style.display = 'none';
    modal.classList.remove('active');
}

// Helper Functions
async function loadUserData() {
    try {
        await Promise.all([
            fetchPersonalNotes(),
            fetchTeams()
        ]);
    } catch (error) {
        console.error('Load user data error:', error);
    }
}

// Check if user is already logged in
function checkAuthStatus() {
    const token = localStorage.getItem('token');
    const refreshToken = localStorage.getItem('refreshToken');
    const userId = localStorage.getItem('userId');
    
    if (token && refreshToken && userId) {
        // Restore session
        state.token = token;
        state.refreshToken = refreshToken;
        state.user = { id: userId };
        
        // Validate token by loading user data
        loadUserData()
            .then(() => {
                showMainContent();
                updateUserInfo();
            })
            .catch(() => {
                // Token invalid, clear session
                logout();
            });
    } else {
        showAuthSection();
    }
}// Teams Functions
async function fetchTeams() {
    try {
        const teams = await apiRequest('/teams');
        state.teams = teams || [];
        renderTeamsList();
    } catch (error) {
        console.error('Fetch teams error:', error);
    }
}

async function createTeam(teamName, isPrivate) {
    try {
        await apiRequest('/teams', 'POST', {
            team_name: teamName,
            user_id: state.user.id,
            is_private: isPrivate
        });
        
        // Refresh teams list
        await fetchTeams();
        
        // Close modal
        closeModal(elements.createTeamModal);
    } catch (error) {
        console.error('Create team error:', error);
        alert(`Failed to create team: ${error.message}`);
    }
}

async function fetchTeamNotes(teamId) {
    try {
        const notes = await apiRequest(`/teams/${teamId}/notes`);
        state.teamNotes = notes || [];
        renderTeamNotesList();
    } catch (error) {
        console.error('Fetch team notes error:', error);
    }
}

async function createTeamNote(teamId, body) {
    try {
        await apiRequest(`/teams/${teamId}/notes`, 'POST', {
            body,
            user_id: state.user.id
        });
        
        // Refresh team notes list
        await fetchTeamNotes(teamId);
        
        // Clear editor
        resetTeamNoteEditor();
    } catch (error) {
        console.error('Create team note error:', error);
        alert(`Failed to create team note: ${error.message}`);
    }
}

async function updateTeamNote(teamId, noteId, body) {
    try {
        await apiRequest(`/teams/${teamId}/notes/${noteId}`, 'PUT', {
            body
        });
        
        // Update local state
        const noteIndex = state.teamNotes.findIndex(note => note.note_id === noteId);
        if (noteIndex !== -1) {
            state.teamNotes[noteIndex].body = body;
            state.teamNotes[noteIndex].updated_at = new Date().toISOString();
        }
        
        // Refresh team notes list
        renderTeamNotesList();
    } catch (error) {
        console.error('Update team note error:', error);
        alert(`Failed to update team note: ${error.message}`);
    }
}

async function deleteTeamNote(teamId, noteId) {
    if (!confirm('Are you sure you want to delete this team note?')) {
        return;
    }
    
    try {
        await apiRequest(`/teams/${teamId}/notes/${noteId}`, 'DELETE');
        
        // Remove from local state
        state.teamNotes = state.teamNotes.filter(note => note.note_id !== noteId);
        
        if (state.currentTeamNote && state.currentTeamNote.note_id === noteId) {
            state.currentTeamNote = null;
            resetTeamNoteEditor();
        }
        
        // Refresh team notes list
        renderTeamNotesList();
    } catch (error) {
        console.error('Delete team note error:', error);
        alert(`Failed to delete team note: ${error.message}`);
    }
}

async function fetchTeamMembers(teamId) {
    try {
        const members = await apiRequest(`/teams/${teamId}/members`);
        state.teamMembers = members || [];
        renderTeamMembersList();
    } catch (error) {
        console.error('Fetch team members error:', error);
    }
}

async function addTeamMember(teamId, userId, role) {
    try {
        await apiRequest(`/teams/${teamId}/members`, 'POST', {
            user_id: userId,
            role
        });
        
        // Refresh team members list
        await fetchTeamMembers(teamId);
    } catch (error) {
        console.error('Add team member error:', error);
        alert(`Failed to add team member: ${error.message}`);
    }
}

async function removeTeamMember(teamId, memberId) {
    if (!confirm('Are you sure you want to remove this member from the team?')) {
        return;
    }
    
    try {
        await apiRequest(`/teams/${teamId}/members/${memberId}`, 'DELETE');
        
        // Refresh team members list
        await fetchTeamMembers(teamId);
    } catch (error) {
        console.error('Remove team member error:', error);
        alert(`Failed to remove team member: ${error.message}`);
    }
}

async function deleteTeam(teamId) {
    if (!confirm('Are you sure you want to delete this team? All team notes will be deleted.')) {
        return;
    }
    
    try {
        await apiRequest(`/teams/${teamId}`, 'DELETE');
        
        // Remove from local state
        state.teams = state.teams.filter(team => team.team_id !== teamId);
        
        if (state.currentTeam && state.currentTeam.team_id === teamId) {
            state.currentTeam = null;
            showNotesView();
        }
        
        // Refresh teams list
        renderTeamsList();
    } catch (error) {
        console.error('Delete team error:', error);
        alert(`Failed to delete team: ${error.message}`);
    }
}

// API Base URL - Update this with your actual API URL
const API_BASE_URL = '/api/v1';

// DOM Elements
const elements = {
    // Auth elements
    authSection: document.getElementById('auth-section'),
    loginForm: document.getElementById('login-form'),
    registerForm: document.getElementById('register-form'),
    authTabs: document.querySelectorAll('.auth-tab'),
    authForms: document.querySelectorAll('.auth-form'),
    loginError: document.getElementById('login-error'),
    registerError: document.getElementById('register-error'),
    
    // Main content elements
    mainContent: document.getElementById('main-content'),
    userInfo: document.getElementById('user-info'),
    logoutBtn: document.getElementById('logout-btn'),
    
    // Notes elements
    personalNotesList: document.getElementById('personal-notes-list'),
    createNoteBtn: document.getElementById('create-note-btn'),
    noteEditor: document.getElementById('note-editor'),
    saveNoteBtn: document.getElementById('save-note-btn'),
    deleteNoteBtn: document.getElementById('delete-note-btn'),
    currentNoteInfo: document.getElementById('current-note-info'),
    
    // Teams elements
    teamsList: document.getElementById('teams-list'),
    createTeamBtn: document.getElementById('create-team-btn'),
    teamView: document.getElementById('team-view'),
    teamName: document.getElementById('team-name'),
    createTeamNoteBtn: document.getElementById('create-team-note-btn'),
    manageTeamBtn: document.getElementById('manage-team-btn'),
    teamNotesList: document.getElementById('team-notes-list'),
    teamNoteEditor: document.getElementById('team-note-editor'),
    saveTeamNoteBtn: document.getElementById('save-team-note-btn'),
    deleteTeamNoteBtn: document.getElementById('delete-team-note-btn'),
    currentTeamNoteInfo: document.getElementById('current-team-note-info'),
    teamNoteEditorContainer: document.getElementById('team-note-editor-container'),
    
    // Modal elements
    modalOverlay: document.getElementById('modal-overlay'),
    createTeamModal: document.getElementById('create-team-modal'),
    createTeamForm: document.getElementById('create-team-form'),
    teamNameInput: document.getElementById('team-name-input'),
    teamPrivateCheckbox: document.getElementById('team-private-checkbox'),
    manageTeamModal: document.getElementById('manage-team-modal'),
    teamMembersList: document.getElementById('team-members-list'),
    addMemberForm: document.getElementById('add-member-form'),
    memberEmailInput: document.getElementById('member-email'),
    memberRoleInput: document.getElementById('member-role'),
    closeModalBtns: document.querySelectorAll('.close-modal')
};

// Auth Functions
async function login(email, password) {
    try {
        const response = await fetch(`${API_BASE_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Login failed');
        }
        
        const userData = await response.json();
        
        // Save auth data
        state.user = {
            id: userData.id,
            email: userData.email,
            hasPremium: userData.has_notes_premium
        };
        state.token = userData.token;
        state.refreshToken = userData.refresh_token;
        
        // Save tokens to localStorage
        localStorage.setItem('token', userData.token);
        localStorage.setItem('refreshToken', userData.refresh_token);
        localStorage.setItem('userId', userData.id);
        
        // Update UI
        showMainContent();
        updateUserInfo();
        
        // Load user data
        await loadUserData();
    } catch (error) {
        console.error('Login error:', error);
        elements.loginError.textContent = error.message;
    }
}

async function register(email, password) {
    try {
        const response = await fetch(`${API_BASE_URL}/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ email, password })
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Registration failed');
        }
        
        // Switch to login tab after successful registration
        elements.authTabs[0].click();
        elements.loginError.textContent = 'Registration successful. Please login.';
        elements.loginError.style.color = 'green';
    } catch (error) {
        console.error('Registration error:', error);
        elements.registerError.textContent = error.message;
    }
}

async function logout() {
    try {
        await fetch(`${API_BASE_URL}/logout`, {
            method: 'POST',
            headers: {
                'Authorization': state.token
            }
        });
    } catch (error) {
        console.error('Logout error:', error);
    } finally {
        // Clear state and localStorage
        state.user = null;
        state.token = null;
        state.refreshToken = null;
        state.currentNote = null;
        state.currentTeam = null;
        state.personalNotes = [];
        state.teams = [];
        
        localStorage.removeItem('token');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('userId');
        
        // Show auth section
        showAuthSection();
    }
}

async function refreshAccessToken() {
    try {
        const response = await fetch(`${API_BASE_URL}/token/refresh`, {
            method: 'POST',
            headers: {
                'Authorization': state.refreshToken
            }
        });
        
        if (!response.ok) {
            throw new Error('Failed to refresh token');
        }
        
        const data = await response.json();
        state.token = data.token;
        localStorage.setItem('token', data.token);
        
        return true;
    } catch (error) {
        console.error('Token refresh error:', error);
        logout(); // Force logout if token refresh fails
        return false;
    }
}

// API Request Helper
async function apiRequest(url, method = 'GET', body = null, needsAuth = true) {
    const headers = {
        'Content-Type': 'application/json'
    };
    
    if (needsAuth && state.token) {
        headers['Authorization'] = state.token;
    }
    
    const options = {
        method,
        headers
    };
    
    if (body && (method === 'POST' || method === 'PUT')) {
        options.body = JSON.stringify(body);
    }
    
    try {
        let response = await fetch(`${API_BASE_URL}${url}`, options);
        
        // Handle token expiration
        if (response.status === 401 && state.refreshToken) {
            const refreshed = await refreshAccessToken();
            if (refreshed) {
                // Retry request with new token
                headers['Authorization'] = state.token;
                options.headers = headers;
                response = await fetch(`${API_BASE_URL}${url}`, options);
            }
        }
        
        if (!response.ok) {
            if (response.headers.get('Content-Type')?.includes('application/json')) {
                const errorData = await response.json();
                throw new Error(errorData.error || `Request failed with status ${response.status}`);
            } else {
                throw new Error(`Request failed with status ${response.status}`);
            }
        }
        
        if (response.status === 204) {
            return null; // No content
        }
        
        if (response.headers.get('Content-Type')?.includes('application/json')) {
            return await response.json();
        }
        
        return null;
    } catch (error) {
        console.error(`API Request Error (${url}):`, error);
        throw error;
    }
}

// Notes Functions
async function fetchPersonalNotes() {
    try {
        const notes = await apiRequest(`/notes?authorId=${state.user.id}`);
        state.personalNotes = notes || [];
        renderPersonalNotesList();
    } catch (error) {
        console.error('Fetch notes error:', error);
    }
}

async function createNote(body) {
    try {
        await apiRequest('/notes', 'POST', {
            body,
            user_id: state.user.id
        });
        
        // Refresh notes list
        await fetchPersonalNotes();
        
        // Clear editor
        resetNoteEditor();
    } catch (error) {
        console.error('Create note error:', error);
        alert(`Failed to create note: ${error.message}`);
    }
}

async function updateNote(noteId, body) {
    try {
        await apiRequest(`/notes/${noteId}`, 'PUT', {
            noteID: noteId,
            body
        });
        
        // Update local state
        const noteIndex = state.personalNotes.findIndex(note => note.id === noteId);
        if (noteIndex !== -1) {
            state.personalNotes[noteIndex].body = body;
            state.personalNotes[noteIndex].updated_at = new Date().toISOString();
        }
        
        // Refresh notes list
        renderPersonalNotesList();
    } catch (error) {
        console.error('Update note error:', error);
        alert(`Failed to update note: ${error.message}`);
    }
}

async function deleteNote(noteId) {
    if (!confirm('Are you sure you want to delete this note?')) {
        return;
    }
    
    try {
        await apiRequest(`/notes/${noteId}`, 'DELETE');
        
        // Remove from local state
        state.personalNotes = state.personalNotes.filter(note => note.id !== noteId);
        
        if (state.currentNote && state.currentNote.id === noteId) {
            state.currentNote = null;
            resetNoteEditor();
        }
        
        // Refresh notes list
        renderPersonalNotesList();
    } catch (error) {
        console.error('Delete note error:', error);
        alert(`Failed to delete note: ${error.message}`);
    }
}