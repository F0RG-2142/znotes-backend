const API_URL = 'http://localhost:8080';
const WS_URL = 'ws://localhost:8080/ws';
let token = 'your-jwt-token'; // Replace with actual JWT token from authentication
let currentGroupId = null;
let ws = null;

// Tab switching
document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
        document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
        tab.classList.add('active');
        document.getElementById(tab.dataset.section).classList.add('active');
        if (tab.dataset.section === 'personal') {
            loadPersonalNotes();
            if (ws) ws.close();
        } else if (tab.dataset.section === 'groups') {
            loadGroups();
            document.getElementById('group-note-editor').style.display = 'none';
        } else if (tab.dataset.section === 'stats') {
            loadStats();
            if (ws) ws.close();
        }
    });
});

// Load personal notes
async function loadPersonalNotes() {
    try {
        const response = await fetch(`${API_URL}/notes`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to fetch notes');
        const notes = await response.json();
        const noteList = document.getElementById('personal-notes');
        noteList.innerHTML = notes.map(note => `
            <div class="note-item">
                <div>
                    <strong>${note.title}</strong>
                    <p>${note.content.substring(0, 100)}${note.content.length > 100 ? '...' : ''}</p>
                    <p>Category: ${note.category || 'None'}</p>
                </div>
                <div>
                    <button onclick="editPersonalNote(${note.id})">Edit</button>
                    <button onclick="deletePersonalNote(${note.id})">Delete</button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        showError(error.message);
    }
}

// Create personal note
async function createPersonalNote() {
    const title = document.getElementById('personal-title').value;
    const content = document.getElementById('personal-content').value;
    if (!title || !content) {
        showError('Title and content are required');
        return;
    }
    try {
        const response = await fetch(`${API_URL}/notes`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ title, content, isGroupNote: false })
        });
        if (!response.ok) throw new Error('Failed to create note');
        document.getElementById('personal-title').value = '';
        document.getElementById('personal-content').value = '';
        loadPersonalNotes();
    } catch (error) {
        showError(error.message);
    }
}

// Edit personal note (load into editor)
async function editPersonalNote(id) {
    try {
        const response = await fetch(`${API_URL}/notes/${id}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to fetch note');
        const note = await response.json();
        document.getElementById('personal-title').value = note.title;
        document.getElementById('personal-content').value = note.content;
        const createButton = document.querySelector('#personal button');
        createButton.textContent = 'Update Note';
        createButton.onclick = async () => {
            try {
                const response = await fetch(`${API_URL}/notes/${id}`, {
                    method: 'PUT',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        title: document.getElementById('personal-title').value,
                        content: document.getElementById('personal-content').value
                    })
                });
                if (!response.ok) throw new Error('Failed to update note');
                document.getElementById('personal-title').value = '';
                document.getElementById('personal-content').value = '';
                createButton.textContent = 'Create Note';
                createButton.onclick = createPersonalNote;
                loadPersonalNotes();
            } catch (error) {
                showError(error.message);
            }
        };
    } catch (error) {
        showError(error.message);
    }
}

// Delete personal note
async function deletePersonalNote(id) {
    try {
        const response = await fetch(`${API_URL}/notes/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to delete note');
        loadPersonalNotes();
    } catch (error) {
        showError(error.message);
    }
}

// Load groups
async function loadGroups() {
    try {
        const response = await fetch(`${API_URL}/groups`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to fetch groups');
        const groups = await response.json();
        const groupList = document.getElementById('group-list');
        groupList.innerHTML = groups.map(group => `
            <div class="group-item">
                <div>
                    <strong>${group.name}</strong>
                </div>
                <div>
                    <button onclick="loadGroupNotes(${group.id}, '${group.name}')">View Notes</button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        showError(error.message);
    }
}

// Create group
async function createGroup() {
    const name = document.getElementById('group-name').value;
    if (!name) {
        showError('Group name is required');
        return;
    }
    try {
        const response = await fetch(`${API_URL}/groups`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ name })
        });
        if (!response.ok) throw new Error('Failed to create group');
        document.getElementById('group-name').value = '';
        loadGroups();
    } catch (error) {
        showError(error.message);
    }
}

// Load group notes and connect WebSocket
async function loadGroupNotes(groupId, groupName) {
    currentGroupId = groupId;
    document.getElementById('group-note-editor').style.display = 'block';
    document.getElementById('group-title').textContent = `Notes for ${groupName}`;
    try {
        const response = await fetch(`${API_URL}/groups/${groupId}/notes`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to fetch group notes');
        const notes = await response.json();
        const noteList = document.getElementById('group-list');
        noteList.innerHTML = notes.map(note => `
            <div class="note-item">
                <div>
                    <strong>${note.title}</strong>
                    <p>${note.content.substring(0, 100)}${note.content.length > 100 ? '...' : ''}</p>
                    <p>Category: ${note.category || 'None'}</p>
                </div>
                <div>
                    <button onclick="editGroupNote(${note.id})">Edit</button>
                    <button onclick="deleteGroupNote(${note.id})">Delete</button>
                </div>
            </div>
        `).join('');

        // Connect WebSocket
        if (ws) ws.close();
        ws = new WebSocket(`${WS_URL}?groupId=${groupId}&token=${token}`);
        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            if (data.type === 'noteUpdate') {
                loadGroupNotes(groupId, groupName); // Refresh notes
            }
        };
        ws.onerror = () => showError('WebSocket connection failed');
    } catch (error) {
        showError(error.message);
    }
}

// Create group note
async function createGroupNote() {
    const title = document.getElementById('group-note-title').value;
    const content = document.getElementById('group-note-content').value;
    if (!title || !content || !currentGroupId) {
        showError('Title, content, and group selection are required');
        return;
    }
    try {
        const response = await fetch(`${API_URL}/groups/${currentGroupId}/notes`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ title, content, isGroupNote: true })
        });
        if (!response.ok) throw new Error('Failed to create group note');
        document.getElementById('group-note-title').value = '';
        document.getElementById('group-note-content').value = '';
        loadGroupNotes(currentGroupId, document.getElementById('group-title').textContent.replace('Notes for ', ''));
    } catch (error) {
        showError(error.message);
    }
}

// Edit group note (load into editor)
async function editGroupNote(id) {
    try {
        const response = await fetch(`${API_URL}/groups/${currentGroupId}/notes/${id}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to fetch group note');
        const note = await response.json();
        document.getElementById('group-note-title').value = note.title;
        document.getElementById('group-note-content').value = note.content;
        const createButton = document.querySelector('#group-note-editor button');
        createButton.textContent = 'Update Group Note';
        createButton.onclick = async () => {
            try {
                const response = await fetch(`${API_URL}/groups/${currentGroupId}/notes/${id}`, {
                    method: 'PUT',
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        title: document.getElementById('group-note-title').value,
                        content: document.getElementById('group-note-content').value
                    })
                });
                if (!response.ok) throw new Error('Failed to update group note');
                document.getElementById('group-note-title').value = '';
                document.getElementById('group-note-content').value = '';
                createButton.textContent = 'Create Group Note';
                createButton.onclick = createGroupNote;
                loadGroupNotes(currentGroupId, document.getElementById('group-title').textContent.replace('Notes for ', ''));
            } catch (error) {
                showError(error.message);
            }
        };
    } catch (error) {
        showError(error.message);
    }
}

// Delete group note
async function deleteGroupNote(id) {
    try {
        const response = await fetch(`${API_URL}/groups/${currentGroupId}/notes/${id}`, {
            method: 'DELETE',
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to delete group note');
        loadGroupNotes(currentGroupId, document.getElementById('group-title').textContent.replace('Notes for ', ''));
    } catch (error) {
        showError(error.message);
    }
}

// Load stats
async function loadStats() {
    try {
        const response = await fetch(`${API_URL}/dashboard_data`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        if (!response.ok) throw new Error('Failed to fetch stats');
        const data = await response.json();
        const statsContent = document.getElementById('stats-content');
        statsContent.innerHTML = `
            <h3>Note Categories</h3>
            ${data.noteStats.map(stat => `<p>${stat.category}: ${stat.count} notes</p>`).join('')}
            <h3>User Activity</h3>
            ${data.userStats.map(stat => `<p>${stat.user}: ${stat.requests} requests</p>`).join('')}
        `;
    } catch (error) {
        showError(error.message);
    }
}

// Show error message
function showError(message) {
    const errorDiv = document.getElementById('error-message');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
    setTimeout(() => { errorDiv.style.display = 'none'; }, 3000);
}

// Initial load
loadPersonalNotes();