// Settings page functionality for RcloneStorage

document.addEventListener('DOMContentLoaded', function() {
    checkAuth();
    loadUserProfile();
    loadSettings();
});

// Load user profile data
async function loadUserProfile() {
    try {
        const token = getAuthToken();
        const response = await fetch('/api/user/profile', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            document.getElementById('email').value = data.email;
            
            // Update storage info
            const storageUsed = formatBytes(data.storage_used || 0);
            const storageQuota = data.storage_quota === -1 ? 'Unlimited' : formatBytes(data.storage_quota);
            
            document.getElementById('storageUsed').textContent = storageUsed;
            document.getElementById('storageQuota').textContent = storageQuota;
            
            // Update progress bar
            if (data.storage_quota > 0) {
                const percentage = (data.storage_used / data.storage_quota) * 100;
                document.getElementById('storageProgress').style.width = `${percentage}%`;
            } else {
                document.getElementById('storageProgress').style.width = '0%';
            }
        }
    } catch (error) {
        console.error('Error loading profile:', error);
        showNotification('Error loading profile data', 'error');
    }
}

// Load user settings
function loadSettings() {
    // Load settings from localStorage or defaults
    const settings = {
        defaultProvider: localStorage.getItem('defaultProvider') || 'union',
        autoCleanup: localStorage.getItem('autoCleanup') === 'true',
        twoFactor: localStorage.getItem('twoFactor') === 'true',
        loginNotifications: localStorage.getItem('loginNotifications') === 'true',
        sessionTimeout: localStorage.getItem('sessionTimeout') || '24',
        theme: localStorage.getItem('theme') || 'light',
        language: localStorage.getItem('language') || 'en',
        dateFormat: localStorage.getItem('dateFormat') || 'MM/DD/YYYY',
        timezone: localStorage.getItem('timezone') || 'UTC',
        uploadNotifications: localStorage.getItem('uploadNotifications') === 'true',
        storageNotifications: localStorage.getItem('storageNotifications') === 'true',
        securityNotifications: localStorage.getItem('securityNotifications') === 'true'
    };

    // Apply settings to form
    Object.keys(settings).forEach(key => {
        const element = document.getElementById(key);
        if (element) {
            if (element.type === 'checkbox') {
                element.checked = settings[key];
            } else {
                element.value = settings[key];
            }
        }
    });

    // Apply theme
    applyTheme(settings.theme);
}

// Change password
async function changePassword() {
    const currentPassword = document.getElementById('currentPassword').value;
    const newPassword = document.getElementById('newPassword').value;
    const confirmPassword = document.getElementById('confirmPassword').value;

    if (!currentPassword || !newPassword || !confirmPassword) {
        showNotification('Please fill in all password fields', 'error');
        return;
    }

    if (newPassword !== confirmPassword) {
        showNotification('New passwords do not match', 'error');
        return;
    }

    if (newPassword.length < 8) {
        showNotification('Password must be at least 8 characters long', 'error');
        return;
    }

    try {
        const token = getAuthToken();
        const response = await fetch('/api/user/change-password', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                current_password: currentPassword,
                new_password: newPassword
            })
        });

        if (response.ok) {
            showNotification('Password changed successfully', 'success');
            // Clear password fields
            document.getElementById('currentPassword').value = '';
            document.getElementById('newPassword').value = '';
            document.getElementById('confirmPassword').value = '';
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to change password', 'error');
        }
    } catch (error) {
        console.error('Error changing password:', error);
        showNotification('Error changing password', 'error');
    }
}

// Save settings
function saveSettings() {
    const settings = {
        defaultProvider: document.getElementById('defaultProvider').value,
        autoCleanup: document.getElementById('autoCleanup').checked,
        twoFactor: document.getElementById('twoFactor').checked,
        loginNotifications: document.getElementById('loginNotifications').checked,
        sessionTimeout: document.getElementById('sessionTimeout').value,
        theme: document.getElementById('theme').value,
        language: document.getElementById('language').value,
        dateFormat: document.getElementById('dateFormat').value,
        timezone: document.getElementById('timezone').value,
        uploadNotifications: document.getElementById('uploadNotifications').checked,
        storageNotifications: document.getElementById('storageNotifications').checked,
        securityNotifications: document.getElementById('securityNotifications').checked
    };

    // Save to localStorage
    Object.keys(settings).forEach(key => {
        localStorage.setItem(key, settings[key]);
    });

    // Apply theme immediately
    applyTheme(settings.theme);

    showNotification('Settings saved successfully', 'success');
}

// Reset settings to defaults
function resetSettings() {
    if (confirm('Are you sure you want to reset all settings to defaults?')) {
        // Clear localStorage
        const settingsKeys = [
            'defaultProvider', 'autoCleanup', 'twoFactor', 'loginNotifications',
            'sessionTimeout', 'theme', 'language', 'dateFormat', 'timezone',
            'uploadNotifications', 'storageNotifications', 'securityNotifications'
        ];

        settingsKeys.forEach(key => {
            localStorage.removeItem(key);
        });

        // Reload settings
        loadSettings();
        showNotification('Settings reset to defaults', 'success');
    }
}

// Apply theme
function applyTheme(theme) {
    const body = document.body;
    body.classList.remove('theme-light', 'theme-dark');

    if (theme === 'dark') {
        body.classList.add('theme-dark');
    } else if (theme === 'light') {
        body.classList.add('theme-light');
    } else if (theme === 'auto') {
        // Use system preference
        if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
            body.classList.add('theme-dark');
        } else {
            body.classList.add('theme-light');
        }
    }
}

// Format bytes to human readable
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// Show notification
function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `notification notification-${type}`;
    notification.innerHTML = `
        <i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i>
        <span>${message}</span>
        <button onclick="this.parentElement.remove()">
            <i class="fas fa-times"></i>
        </button>
    `;

    // Add to page
    document.body.appendChild(notification);

    // Auto remove after 5 seconds
    setTimeout(() => {
        if (notification.parentElement) {
            notification.remove();
        }
    }, 5000);
}

// Get auth token
function getAuthToken() {
    return localStorage.getItem('auth_token') || localStorage.getItem('token');
}

// Check authentication
function checkAuth() {
    const token = getAuthToken();
    if (!token) {
        console.log('No token found, redirecting to login');
        window.location.href = '/login.html';
        return;
    }

    // Check if user is admin
    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        if (payload.role === 'admin') {
            document.querySelectorAll('.admin-only').forEach(el => {
                el.style.display = 'block';
            });
        }
    } catch (error) {
        console.error('Error parsing token:', error);
        // Don't redirect on token parse error, just log it
    }
}

// Logout function
function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login.html';
}