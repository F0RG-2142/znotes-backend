// script.js
document.addEventListener('DOMContentLoaded', () => {
    const API_BASE_URL = '/api/v1';

    // --- Element Selectors ---
    const authContainer = document.getElementById('auth-container');
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    const authError = document.getElementById('auth-error');

    const appContainer = document.getElementById('app-container');
    const userGreeting = document.getElementById('user-greeting');
    const logoutButton = document.getElementById('logout-button');

    const createPersonalNoteBtn = document.getElementById('create-personal-note-btn');
    const createGroupBtn = document.getElementById('create-group-btn');
    const groupSelector = document.getElementById('group-selector');
    const manageSelectedGroupBtn = document.getElementById('manage-selected-group-btn');
    const createGroupNoteBtn = document.getElementById('create-group-note-btn');
    const currentGroupNameDisplay = document.getElementById('current-group-name-display');


    const noteEditorModal = document.getElementById('note-editor-modal');
    const noteEditorTitle = document.getElementById('note-editor-title');
    const noteIdInput = document.getElementById('note-id');
    const noteTypeInput = document.getElementById('note-type'); // 'personal' or 'group'
    const noteGroupIdInput = document.getElementById('note-group-id');
    const noteTitleInput = document.getElementById('note-title-input');
    const noteContentInput = document.getElementById('note-content-input');
    const saveNoteButton = document.getElementById('save-note-button');
    const editorError = document.getElementById('editor-error');

    const groupCreatorModal = document.getElementById('group-creator-modal');
    const groupNameInput = document.getElementById('group-name-input');
    const saveGroupButton = document.getElementById('save-group-button');
    const groupCreatorError = document.getElementById('group-creator-error');

    const groupManagerModal = document.getElementById('group-manager-modal');
    const managingGroupName = document.getElementById('managing-group-name');
    const managingGroupIdInput = document.getElementById('managing-group-id');
    const updateGroupNameInput = document.getElementById('update-group-name-input');
    const updateGroupNameButton = document.getElementById('update-group-name-button');
    const updateGroupError = document.getElementById('update-group-error');
    const memberCountDisplay = document.getElementById('member-count');
    const groupMembersList = document.getElementById('group-members-list');
    const addMemberEmailInput = document.getElementById('add-member-email-input');
    const addMemberButton = document.getElementById('add-member-button');
    const groupManagerError = document.getElementById('group-manager-error');
    const deleteGroupButton = document.getElementById('delete-group-button');
    const deleteGroupError = document.getElementById('delete-group-error');


    const personalNotesList = document.getElementById('personal-notes-list');
    const groupNotesList = document.getElementById('group-notes-list');

    let currentUser = null;
    let currentGroups = [];
    // let webSocket = null; // WebSocket Placeholder

    // --- Utility Functions ---
    const getAuthToken = () => localStorage.getItem('authToken');
    const setAuthToken = (token) => localStorage.setItem('authToken', token);
    const getRefreshToken = () => localStorage.getItem('refreshToken');
    const setRefreshToken = (token) => localStorage.setItem('refreshToken', token);

    function setCurrentUser(user) {
        currentUser = user;
        localStorage.setItem('currentUser', JSON.stringify(user));
        userGreeting.textContent = `Hello, ${user.name || user.email}!`;
    }

    function getCurrentUser() {
        if (currentUser) return currentUser;
        const userStr = localStorage.getItem('currentUser');
        return userStr ? JSON.parse(userStr) : null;
    }

    function clearAuthData() {
        localStorage.removeItem('authToken');
        localStorage.removeItem('refreshToken');
        localStorage.removeItem('currentUser');
        currentUser = null;
    }

    async function apiRequest(endpoint, method = 'GET', body = null, requiresAuth = true) {
        const headers = { 'Content-Type': 'application/json' };
        let token = requiresAuth ? getAuthToken() : null;

        if (requiresAuth && !token) {
            showAuthView('Session expired or not logged in.');
            return Promise.reject('No auth token');
        }
        if (token) headers['Authorization'] = `Bearer ${token}`;

        const config = { method, headers };
        if (body) config.body = JSON.stringify(body);

        try {
            let response = await fetch(`${API_BASE_URL}${endpoint}`, config);

            if (response.status === 401 && requiresAuth) {
                const refreshed = await attemptTokenRefresh();
                if (refreshed) {
                    headers['Authorization'] = `Bearer ${getAuthToken()}`; // Update header with new token
                    response = await fetch(`${API_BASE_URL}${endpoint}`, config); // Retry original request
                } else {
                    logoutUser(); // Full logout if refresh fails
                    throw new Error('Session expired. Please login again.');
                }
            }

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({ message: response.statusText }));
                throw new Error(errorData.message || `HTTP error ${response.status}`);
            }
            return response.status === 204 ? null : response.json();
        } catch (error) {
            console.error(`API Request Error (${method} ${endpoint}):`, error.message);
            throw error; // Re-throw to be caught by calling function
        }
    }

    async function attemptTokenRefresh() {
        const rToken = getRefreshToken();
        if (!rToken) return false;
        try {
            const data = await apiRequest('/token/refresh', 'POST', { refresh_token: rToken }, false); // refresh is unauthenticated initially
            setAuthToken(data.access_token);
            if (data.refresh_token) setRefreshToken(data.refresh_token); // If backend sends new refresh token
            return true;
        } catch (error) {
            console.error('Token refresh failed:', error.message);
            return false;
        }
    }

    // --- View Management ---
    function showAuthView(message = '') {
        authContainer.style.display = 'block';
        appContainer.style.display = 'none';
        authError.textContent = message;
        loginForm.reset();
        registerForm.reset();
    }

    function showAppView() {
        authContainer.style.display = 'none';
        appContainer.style.display = 'block';
        const user = getCurrentUser();
        if (user) {
            setCurrentUser(user); // Ensures greeting is set
            loadInitialAppData();
        } else {
            showAuthView('User data not found. Please login.');
        }
    }

    // --- Modal Management ---
    function openModal(modalElement) {
        modalElement.style.display = 'block';
    }
    function closeModal(modalElement) {
        modalElement.style.display = 'none';
        const errorP = modalElement.querySelector('.error-message');
        if (errorP) errorP.textContent = '';
    }
    document.querySelectorAll('.modal .close-button').forEach(btn =>
        btn.addEventListener('click', (e) => closeModal(e.target.closest('.modal')))
    );
    window.addEventListener('click', (event) => {
        if (event.target.classList.contains('modal')) closeModal(event.target);
    });


    // --- Authentication ---
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        authError.textContent = '';
        try {
            const data = await apiRequest('/login', 'POST', {
                email: loginForm['login-email'].value,
                password: loginForm['login-password'].value,
            }, false);
            setAuthToken(data.access_token);
            setRefreshToken(data.refresh_token);
            setCurrentUser(data.user); // Backend should return user object
            showAppView();
        } catch (error) {
            authError.textContent = `Login failed: ${error.message}`;
        }
    });

    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        authError.textContent = '';
        try {
            await apiRequest('/register', 'POST', {
                name: registerForm['register-name'].value,
                email: registerForm['register-email'].value,
                password: registerForm['register-password'].value,
            }, false);
            authError.textContent = 'Registration successful! Please login.';
            registerForm.reset();
        } catch (error) {
            authError.textContent = `Registration failed: ${error.message}`;
        }
    });

    logoutButton.addEventListener('click', async () => {
        try {
            const rToken = getRefreshToken();
            if (rToken) await apiRequest('/logout', 'POST', { refresh_token: rToken });
        } catch (error) {
            console.warn('Logout API call failed (token might already be invalid):', error.message);
        } finally {
            logoutUser();
        }
    });

    function logoutUser() {
        clearAuthData();
        // closeWebSocket(); // If using WebSockets
        showAuthView('You have been logged out.');
        personalNotesList.innerHTML = '';
        groupNotesList.innerHTML = '';
        groupSelector.innerHTML = '<option value="">-- Select a Group --</option>';
        currentGroupNameDisplay.textContent = 'No group selected';
        manageSelectedGroupBtn.style.display = 'none';
        createGroupNoteBtn.style.display = 'none';
    }

    // --- Note Rendering & HTML Escaping ---
    function escapeHTML(str) {
        if (str === null || str === undefined) return '';
        return str.toString().replace(/[&<>"']/g, m => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#039;'})[m]);
    }

    function renderNoteItem(note, type, listElement, groupId = null) {
        const noteEl = document.createElement('div');
        noteEl.classList.add('note-item');
        noteEl.dataset.id = note.id;
        if (groupId) noteEl.dataset.groupId = groupId;

        noteEl.innerHTML = `
            <h3>${escapeHTML(note.title)}</h3>
            <p>${escapeHTML(note.content)}</p>
            <div class="meta">
                <small>ID: ${note.id}</small><br>
                <small>Last updated: ${new Date(note.updated_at || note.created_at).toLocaleString()}</small>
            </div>
            <div class="actions">
                <button class="edit-button">Edit</button>
                <button class="delete-button danger-button">Delete</button>
            </div>
        `;
        noteEl.querySelector('.edit-button').addEventListener('click', () => openNoteEditor(note, type, groupId));
        noteEl.querySelector('.delete-button').addEventListener('click', () => {
            if (confirm(`Delete note "${escapeHTML(note.title)}"?`)) {
                type === 'personal' ? deletePersonalNote(note.id) : deleteGroupNote(groupId, note.id);
            }
        });
        listElement.appendChild(noteEl);
    }

    // --- Personal Notes ---
    createPersonalNoteBtn.addEventListener('click', () => openNoteEditor(null, 'personal'));

    async function loadPersonalNotes() {
        personalNotesList.innerHTML = '<p>Loading personal notes...</p>';
        try {
            const notes = await apiRequest('/notes'); // GET /api/v1/notes
            personalNotesList.innerHTML = '';
            if (notes && notes.length) {
                notes.forEach(note => renderNoteItem(note, 'personal', personalNotesList));
            } else {
                personalNotesList.innerHTML = '<p>No personal notes yet. Create one!</p>';
            }
        } catch (error) {
            personalNotesList.innerHTML = `<p class="error-message">Failed to load personal notes: ${error.message}</p>`;
        }
    }

    async function savePersonalNote(id, title, content) {
        editorError.textContent = '';
        const endpoint = id ? `/notes/${id}` : '/notes';
        const method = id ? 'PUT' : 'POST';
        try {
            await apiRequest(endpoint, method, { title, content });
            closeModal(noteEditorModal);
            loadPersonalNotes();
            // WEBSOCKET_EVENT: { type: 'personal_note_changed', userId: currentUser.id }
        } catch (error) {
            editorError.textContent = `Save failed: ${error.message}`;
        }
    }

    async function deletePersonalNote(noteId) {
        try {
            await apiRequest(`/notes/${noteId}`, 'DELETE');
            loadPersonalNotes();
            // WEBSOCKET_EVENT: { type: 'personal_note_deleted', userId: currentUser.id, noteId: noteId }
        } catch (error) {
            alert(`Delete failed: ${error.message}`);
        }
    }

    // --- Group Management ---
    createGroupBtn.addEventListener('click', () => {
        groupNameInput.value = '';
        groupCreatorError.textContent = '';
        openModal(groupCreatorModal);
    });

    saveGroupButton.addEventListener('click', async () => {
        const name = groupNameInput.value.trim();
        if (!name) {
            groupCreatorError.textContent = 'Group name is required.';
            return;
        }
        groupCreatorError.textContent = '';
        try {
            await apiRequest('/groups', 'POST', { name });
            closeModal(groupCreatorModal);
            loadUserGroups(); // Refresh groups
            // WEBSOCKET_EVENT: { type: 'group_created_or_user_added', involvedUserIds: [currentUser.id] } (simplified)
        } catch (error) {
            groupCreatorError.textContent = `Failed to create group: ${error.message}`;
        }
    });

    manageSelectedGroupBtn.addEventListener('click', () => {
        const selectedGroupId = groupSelector.value;
        const group = currentGroups.find(g => g.id === selectedGroupId);
        if (group) openGroupManagerModal(group);
    });

    async function loadUserGroups() {
        groupSelector.disabled = true;
        try {
            const groups = await apiRequest('/groups'); // GET /api/v1/groups (should be user's groups)
            currentGroups = groups || [];
            groupSelector.innerHTML = '<option value="">-- Select a Group --</option>';
            if (currentGroups.length > 0) {
                currentGroups.forEach(group => {
                    const option = document.createElement('option');
                    option.value = group.id;
                    option.textContent = escapeHTML(group.name);
                    groupSelector.appendChild(option);
                });
                groupSelector.disabled = false;
                // If a group was previously selected and still exists, re-select it
                const previouslySelected = groupSelector.dataset.prevValue;
                if (previouslySelected && currentGroups.some(g => g.id === previouslySelected)) {
                    groupSelector.value = previouslySelected;
                } else if (currentGroups.length > 0){
                     groupSelector.value = currentGroups[0].id; // Select first group
                }
                groupSelector.dispatchEvent(new Event('change')); // Trigger note load for selected group
            } else {
                groupNotesList.innerHTML = '<p>You are not part of any groups. Create or join one!</p>';
                currentGroupNameDisplay.textContent = 'No group selected';
                createGroupNoteBtn.style.display = 'none';
                manageSelectedGroupBtn.style.display = 'none';
            }
        } catch (error) {
            console.error("Failed to load groups:", error.message);
            groupNotesList.innerHTML = `<p class="error-message">Could not load groups: ${error.message}</p>`;
        } finally {
            groupSelector.disabled = false;
        }
    }

    groupSelector.addEventListener('change', function() {
        const selectedGroupId = this.value;
        groupSelector.dataset.prevValue = selectedGroupId; // Store selection
        if (selectedGroupId) {
            const selectedGroup = currentGroups.find(g => g.id === selectedGroupId);
            currentGroupNameDisplay.textContent = selectedGroup ? escapeHTML(selectedGroup.name) : 'Unknown Group';
            loadGroupNotes(selectedGroupId);
            createGroupNoteBtn.style.display = 'inline-block';
            manageSelectedGroupBtn.style.display = 'inline-block';
        } else {
            groupNotesList.innerHTML = '<p>Select a group to view its notes.</p>';
            currentGroupNameDisplay.textContent = 'No group selected';
            createGroupNoteBtn.style.display = 'none';
            manageSelectedGroupBtn.style.display = 'none';
        }
    });


    // --- Group Manager Modal Logic ---
    function openGroupManagerModal(group) {
        managingGroupIdInput.value = group.id;
        managingGroupName.textContent = escapeHTML(group.name);
        updateGroupNameInput.value = group.name;
        updateGroupError.textContent = '';
        groupManagerError.textContent = '';
        deleteGroupError.textContent = '';
        addMemberEmailInput.value = '';
        loadGroupMembers(group.id);
        openModal(groupManagerModal);
    }

    updateGroupNameButton.addEventListener('click', async () => {
        const groupId = managingGroupIdInput.value;
        const newName = updateGroupNameInput.value.trim();
        if (!newName) {
            updateGroupError.textContent = 'Group name cannot be empty.';
            return;
        }
        updateGroupError.textContent = '';
        try {
            await apiRequest(`/groups/${groupId}`, 'PUT', { name: newName });
            // Update local group name and selector
            const groupInState = currentGroups.find(g => g.id === groupId);
            if (groupInState) groupInState.name = newName;
            managingGroupName.textContent = escapeHTML(newName);
            const optionInSelector = groupSelector.querySelector(`option[value="${groupId}"]`);
            if (optionInSelector) optionInSelector.textContent = escapeHTML(newName);
            if(groupSelector.value === groupId) currentGroupNameDisplay.textContent = escapeHTML(newName);

            updateGroupError.textContent = 'Name updated!';
            setTimeout(() => updateGroupError.textContent = '', 2000);
            // WEBSOCKET_EVENT: { type: 'group_details_changed', groupId: groupId }
        } catch (error) {
            updateGroupError.textContent = `Failed to update name: ${error.message}`;
        }
    });

    addMemberButton.addEventListener('click', async () => {
        const groupId = managingGroupIdInput.value;
        const email = addMemberEmailInput.value.trim();
        if (!email) {
            groupManagerError.textContent = 'Enter an email to add.';
            return;
        }
        groupManagerError.textContent = '';
        try {
            // Backend needs to handle finding user by email and adding their ID.
            // The request might be { user_email: email } or { user_id: ... } if frontend did a lookup first.
            // For simplicity, sending email.
            await apiRequest(`/groups/${groupId}/members`, 'POST', { email: email });
            addMemberEmailInput.value = '';
            loadGroupMembers(groupId); // Refresh list
            // WEBSOCKET_EVENT: { type: 'group_member_added', groupId: groupId, newMemberEmail: email (or ID) }
        } catch (error) {
            groupManagerError.textContent = `Failed to add member: ${error.message}`;
        }
    });

    async function loadGroupMembers(groupId) {
        groupMembersList.innerHTML = '<li>Loading members...</li>';
        memberCountDisplay.textContent = '...';
        try {
            const members = await apiRequest(`/groups/${groupId}/members`); // Expects array of users {id, name, email}
            groupMembersList.innerHTML = '';
            memberCountDisplay.textContent = members ? members.length : 0;
            if (members && members.length) {
                members.forEach(member => {
                    const li = document.createElement('li');
                    li.textContent = `${escapeHTML(member.name || member.email)} (ID: ${member.id})`;
                    // Add remove button if API allows (e.g., if current user is owner/admin)
                    // And if not trying to remove self (usually not allowed this way)
                    if (member.id !== getCurrentUser().id) {
                        const removeBtn = document.createElement('button');
                        removeBtn.textContent = 'Remove';
                        removeBtn.onclick = async () => {
                            if (confirm(`Remove ${escapeHTML(member.name || member.email)} from the group?`)) {
                                try {
                                    await apiRequest(`/groups/${groupId}/members/${member.id}`, 'DELETE');
                                    loadGroupMembers(groupId); // Refresh
                                    // WEBSOCKET_EVENT: { type: 'group_member_removed', groupId: groupId, memberId: member.id }
                                } catch (err) {
                                    groupManagerError.textContent = `Failed to remove: ${err.message}`;
                                }
                            }
                        };
                        li.appendChild(removeBtn);
                    }
                    groupMembersList.appendChild(li);
                });
            } else {
                groupMembersList.innerHTML = '<li>No other members in this group.</li>';
            }
        } catch (error) {
            groupMembersList.innerHTML = `<li><p class="error-message">Failed to load members: ${error.message}</p></li>`;
        }
    }

    deleteGroupButton.addEventListener('click', async () => {
        const groupId = managingGroupIdInput.value;
        const groupName = managingGroupName.textContent;
        if (confirm(`ARE YOU SURE you want to permanently delete the group "${groupName}" and all its notes? This cannot be undone.`)) {
            deleteGroupError.textContent = '';
            try {
                await apiRequest(`/groups/${groupId}`, 'DELETE');
                closeModal(groupManagerModal);
                loadUserGroups(); // Reload group list
                groupNotesList.innerHTML = '<p>Group deleted. Select another group.</p>';
                currentGroupNameDisplay.textContent = 'No group selected';
                // WEBSOCKET_EVENT: { type: 'group_deleted', groupId: groupId }
            } catch (error) {
                deleteGroupError.textContent = `Failed to delete group: ${error.message}`;
            }
        }
    });


    // --- Group Notes ---
    createGroupNoteBtn.addEventListener('click', () => {
        const selectedGroupId = groupSelector.value;
        if (!selectedGroupId) {
            alert('Select a group first!');
            return;
        }
        openNoteEditor(null, 'group', selectedGroupId);
    });

    async function loadGroupNotes(groupId) {
        groupNotesList.innerHTML = `<p>Loading notes for group ${escapeHTML(currentGroups.find(g=>g.id===groupId)?.name || groupId)}...</p>`;
        try {
            const notes = await apiRequest(`/groups/${groupId}/notes`);
            groupNotesList.innerHTML = '';
            if (notes && notes.length) {
                notes.forEach(note => renderNoteItem(note, 'group', groupNotesList, groupId));
            } else {
                groupNotesList.innerHTML = '<p>No notes in this group yet. Create one!</p>';
            }
        } catch (error) {
            groupNotesList.innerHTML = `<p class="error-message">Failed to load group notes: ${error.message}</p>`;
        }
    }

    async function saveGroupNote(noteId, groupId, title, content) {
        editorError.textContent = '';
        const endpoint = noteId ? `/groups/${groupId}/notes/${noteId}` : `/groups/${groupId}/notes`;
        const method = noteId ? 'PUT' : 'POST';
        try {
            await apiRequest(endpoint, method, { title, content });
            closeModal(noteEditorModal);
            loadGroupNotes(groupId); // Reload notes for the current group
            // WEBSOCKET_EVENT: { type: 'group_note_changed', groupId: groupId }
        } catch (error) {
            editorError.textContent = `Save failed: ${error.message}`;
        }
    }

    async function deleteGroupNote(groupId, noteId) {
        try {
            await apiRequest(`/groups/${groupId}/notes/${noteId}`, 'DELETE');
            loadGroupNotes(groupId); // Reload notes for the current group
            // WEBSOCKET_EVENT: { type: 'group_note_deleted', groupId: groupId, noteId: noteId }
        } catch (error) {
            alert(`Delete failed: ${error.message}`);
        }
    }

    // --- Note Editor (Common) ---
    function openNoteEditor(note = null, type = 'personal', groupId = null) {
        editorError.textContent = '';
        noteIdInput.value = note ? note.id : '';
        noteTypeInput.value = type;
        noteGroupIdInput.value = (type === 'group' && groupId) ? groupId : '';
        noteTitleInput.value = note ? note.title : '';
        noteContentInput.value = note ? note.content : '';

        if (type === 'personal') {
            noteEditorTitle.textContent = note ? 'Edit Personal Note' : 'Create Personal Note';
        } else if (type === 'group') {
            const groupName = currentGroups.find(g => g.id === groupId)?.name || 'Selected Group';
            noteEditorTitle.textContent = note ? `Edit Note in "${escapeHTML(groupName)}"` : `Create Note in "${escapeHTML(groupName)}"`;
        }
        openModal(noteEditorModal);
        noteTitleInput.focus();
    }

    saveNoteButton.addEventListener('click', () => {
        const id = noteIdInput.value;
        const type = noteTypeInput.value;
        const groupId = noteGroupIdInput.value;
        const title = noteTitleInput.value.trim();
        const content = noteContentInput.value.trim();

        if (!title) {
            editorError.textContent = 'Title is required.';
            return;
        }
        if (!content) {
            editorError.textContent = 'Content is required.';
            return;
        }

        if (type === 'personal') {
            savePersonalNote(id, title, content);
        } else if (type === 'group' && groupId) {
            saveGroupNote(id, groupId, title, content);
        }
    });

    // --- WebSocket Placeholder & Initialization ---
    function initializeWebSocket() {
        // const token = getAuthToken();
        // if (!token || webSocket) return; // No token or already connected

        // const wsScheme = window.location.protocol === "https:" ? "wss:" : "ws:";
        // const wsURL = `${wsScheme}//${window.location.host}/ws/connect?token=${token}`; // Example path
        // webSocket = new WebSocket(wsURL);

        // webSocket.onopen = () => console.log("WebSocket Connected");
        // webSocket.onclose = (event) => {
        //     console.log("WebSocket Disconnected:", event.reason || "No reason provided");
        //     webSocket = null;
        //     // Optional: implement reconnection logic here, perhaps with a delay
        //     // setTimeout(() => initializeWebSocket(), 5000); // Reconnect after 5s
        // };
        // webSocket.onerror = (error) => console.error("WebSocket Error:", error);
        // webSocket.onmessage = (event) => {
        //     try {
        //         const message = JSON.parse(event.data);
        //         console.log("WS Message:", message);
        //         handleWebSocketMessage(message);
        //     } catch (e) {
        //         console.error("Failed to parse WS message:", e);
        //     }
        // };
    }

    function closeWebSocket() {
        // if (webSocket) {
        //     webSocket.close(1000, "User logout"); // 1000 is normal closure
        //     webSocket = null;
        // }
    }

    function handleWebSocketMessage(message) {
        // const { type, data } = message;
        // console.log(`Received WebSocket message: type=${type}, data=`, data);

        // // Personal notes updates
        // if (type === 'personal_note_created' || type === 'personal_note_updated' || type === 'personal_note_deleted') {
        //     if (data.user_id === currentUser.id) { // Ensure it's for the current user
        //         console.log("Refreshing personal notes due to WS message.");
        //         loadPersonalNotes();
        //     }
        // }

        // // Group related updates
        // if (type === 'group_created' || type === 'group_deleted' || type === 'group_member_added' || type === 'group_member_removed' || type === 'group_details_changed') {
        //      // Check if current user is affected (e.g., for member changes) or if it's a group they care about
        //     console.log("Refreshing groups list due to WS message for groups.");
        //     loadUserGroups(); // This can be broad; more targeted updates are better
        //     // If current group details changed, and it's the one being viewed in manager
        //     if (type === 'group_details_changed' && managingGroupIdInput.value === data.group_id) {
        //          const group = currentGroups.find(g => g.id === data.group_id);
        //          if(group) openGroupManagerModal(group); // Re-open/refresh manager
        //     }
        // }

        // // Group notes updates
        // if (type === 'group_note_created' || type === 'group_note_updated' || type === 'group_note_deleted') {
        //     const selectedGroupId = groupSelector.value;
        //     if (data.group_id === selectedGroupId) { // Only refresh if the affected group is currently selected
        //         console.log(`Refreshing notes for group ${data.group_id} due to WS message.`);
        //         loadGroupNotes(data.group_id);
        //     }
        // }
        // Add more specific handlers as needed
    }


    // --- Initial Application Load ---
    async function loadInitialAppData() {
        userGreeting.textContent = `Hello, ${currentUser.name || currentUser.email}!`;
        await Promise.all([
            loadPersonalNotes(),
            loadUserGroups() // This will also trigger loading notes for the default/first group if any
        ]);
        // initializeWebSocket(); // Initialize WebSocket after initial data load and user is authenticated
    }

    async function checkLoginStatus() {
        const token = getAuthToken();
        const rToken = getRefreshToken();
        const user = getCurrentUser();

        if (token && rToken && user) {
            // Optimistically show app view while token is validated/refreshed
            setCurrentUser(user); // Set current user from localStorage first
            showAppView(); // Load the app UI quickly

            try {
                // Validate token by fetching user data or a dedicated 'verify' endpoint
                const freshUser = await apiRequest('/user/me'); // GET current user
                setCurrentUser(freshUser); // Update with fresh data from server
                loadInitialAppData(); // Re-ensure data is loaded after user confirmation
            } catch (error) { // This catch is important if /user/me fails (e.g. token truly expired)
                console.warn("Initial token validation/refresh failed:", error.message);
                logoutUser(); // If validation fails, proper logout
            }
        } else {
            showAuthView();
        }
    }

    checkLoginStatus();
});