// Profile page functionality for RcloneStorage

class ProfilePage {
    constructor() {
        this.user = null;
        this.apiKeys = [];
        this.storageChart = null;
        this.createApiKeyModal = null;
        this.apiKeyCreatedModal = null;
        this.init();
    }

    async init() {
        // Check authentication
        if (!authManager.isAuthenticated()) {
            window.location.href = '/login.html';
            return;
        }

        // Load data
        await this.loadUserProfile();
        await this.loadApiKeys();
        await this.loadRecentActivity();
    }

    async loadUserProfile() {
        try {
            const result = await authManager.getProfile();
            if (result.success) {
                this.user = result.user;
                this.updateProfileUI();
                this.renderStorageChart();
            }
        } catch (error) {
            console.error('Failed to load user profile:', error);
            this.showError('Failed to load profile data');
        }
    }

    updateProfileUI() {
        const user = this.user;
        
        // Profile card
        document.getElementById('profileName').textContent = user.email.split('@')[0];
        document.getElementById('profileEmail').textContent = user.email;
        document.getElementById('profileRole').textContent = user.role.charAt(0).toUpperCase() + user.role.slice(1);
        document.getElementById('profileRole').className = `badge ${user.role === 'admin' ? 'bg-danger' : 'bg-primary'}`;
        
        // Navigation
        document.getElementById('userEmail').textContent = user.email;
        if (user.role === 'admin') {
            document.getElementById('adminMenuItem').classList.remove('d-none');
        }
        
        // Storage info
        const usedText = fileManager.formatFileSize(user.storage_used || 0);
        const quotaText = user.storage_quota === -1 ? 'Unlimited' : fileManager.formatFileSize(user.storage_quota);
        const usagePercent = user.usage_percent || 0;
        
        document.getElementById('profileStorageUsed').textContent = usedText;
        document.getElementById('profileUsagePercent').textContent = `${usagePercent.toFixed(1)}%`;
        
        // Account information
        document.getElementById('accountEmail').textContent = user.email;
        document.getElementById('accountRole').textContent = user.role.charAt(0).toUpperCase() + user.role.slice(1);
        document.getElementById('accountCreated').textContent = new Date(user.created_at).toLocaleDateString();
        document.getElementById('accountQuota').textContent = quotaText;
        document.getElementById('accountUsed').textContent = usedText;
        
        // Storage details
        document.getElementById('storageUsedText').textContent = usedText;
        document.getElementById('storageTotalText').textContent = quotaText;
        
        if (user.storage_quota === -1) {
            document.getElementById('storageAvailableText').textContent = 'Unlimited';
        } else {
            const available = user.storage_quota - (user.storage_used || 0);
            document.getElementById('storageAvailableText').textContent = fileManager.formatFileSize(Math.max(0, available));
        }
    }

    renderStorageChart() {
        const ctx = document.getElementById('storageChart').getContext('2d');
        
        if (this.storageChart) {
            this.storageChart.destroy();
        }

        const used = this.user.storage_used || 0;
        const quota = this.user.storage_quota === -1 ? used * 2 : this.user.storage_quota;
        const available = Math.max(0, quota - used);

        this.storageChart = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: ['Used', 'Available'],
                datasets: [{
                    data: [used, available],
                    backgroundColor: ['#dc3545', '#28a745'],
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'bottom'
                    },
                    tooltip: {
                        callbacks: {
                            label: function(context) {
                                const label = context.label || '';
                                const value = fileManager.formatFileSize(context.raw);
                                return `${label}: ${value}`;
                            }
                        }
                    }
                }
            }
        });
    }

    async loadApiKeys() {
        try {
            const result = await authManager.listApiKeys();
            this.apiKeys = result.api_keys || [];
            this.renderApiKeys();
        } catch (error) {
            console.error('Failed to load API keys:', error);
            document.getElementById('apiKeysLoading').style.display = 'none';
            document.getElementById('noApiKeys').style.display = 'block';
        }
    }

    renderApiKeys() {
        document.getElementById('apiKeysLoading').style.display = 'none';
        
        if (this.apiKeys.length === 0) {
            document.getElementById('noApiKeys').style.display = 'block';
            document.getElementById('apiKeysList').style.display = 'none';
            return;
        }

        document.getElementById('noApiKeys').style.display = 'none';
        document.getElementById('apiKeysList').style.display = 'block';

        const container = document.getElementById('apiKeysList');
        container.innerHTML = this.apiKeys.map(key => `
            <div class="border rounded p-3 mb-3">
                <div class="d-flex justify-content-between align-items-start">
                    <div>
                        <h6 class="mb-1">${key.name}</h6>
                        <small class="text-muted">
                            Created: ${new Date(key.created_at).toLocaleDateString()}
                            ${key.last_used ? `• Last used: ${new Date(key.last_used).toLocaleDateString()}` : '• Never used'}
                        </small>
                    </div>
                    <button class="btn btn-outline-danger btn-sm" onclick="profilePage.deleteApiKey(${key.id}, '${key.name}')">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    }

    async loadRecentActivity() {
        try {
            // Get recent files as activity
            const filesData = await fileManager.getFiles();
            const recentFiles = (filesData.files || []).slice(0, 5);
            
            const container = document.getElementById('recentActivity');
            
            if (recentFiles.length === 0) {
                container.innerHTML = `
                    <div class="text-center py-4">
                        <i class="fas fa-history fa-2x text-muted mb-3"></i>
                        <p class="text-muted">No recent activity</p>
                    </div>
                `;
                return;
            }

            container.innerHTML = recentFiles.map(file => `
                <div class="d-flex align-items-center mb-3 pb-3 border-bottom">
                    <div class="me-3">
                        <i class="${fileManager.getFileIcon(file.mime_type, file.name)}"></i>
                    </div>
                    <div class="flex-grow-1">
                        <div class="fw-medium">${file.name}</div>
                        <small class="text-muted">
                            Uploaded ${fileManager.formatDate(file.modified)} • ${fileManager.formatFileSize(file.size)}
                        </small>
                    </div>
                    <div class="ms-2">
                        <button class="btn btn-outline-primary btn-sm" onclick="profilePage.downloadFile('${file.id}')">
                            <i class="fas fa-download"></i>
                        </button>
                    </div>
                </div>
            `).join('');
            
            // Update file count
            document.getElementById('profileFileCount').textContent = filesData.total || 0;
            document.getElementById('accountFiles').textContent = filesData.total || 0;
            
        } catch (error) {
            console.error('Failed to load recent activity:', error);
            document.getElementById('recentActivity').innerHTML = `
                <div class="text-center py-4 text-muted">
                    <i class="fas fa-exclamation-triangle"></i>
                    Failed to load recent activity
                </div>
            `;
        }
    }

    showCreateApiKeyModal() {
        document.getElementById('apiKeyName').value = '';
        document.getElementById('createApiKeyModal').classList.remove('hidden');
    }

    async createApiKey() {
        const name = document.getElementById('apiKeyName').value.trim();
        
        if (!name) {
            this.showError('Please enter a name for the API key');
            return;
        }

        try {
            const result = await authManager.createApiKey(name);
            
            // Show the created API key
            document.getElementById('newApiKey').value = result.key;
            document.getElementById('apiKeyExample').textContent = result.key;
            
            document.getElementById('createApiKeyModal').classList.add('hidden');
            document.getElementById('apiKeyCreatedModal').classList.remove('hidden');
            
            // Reload API keys list
            await this.loadApiKeys();
            
        } catch (error) {
            console.error('Failed to create API key:', error);
            this.showError('Failed to create API key: ' + error.message);
        }
    }

    async deleteApiKey(keyId, keyName) {
        if (!confirm(`Are you sure you want to delete the API key "${keyName}"? This action cannot be undone.`)) {
            return;
        }

        try {
            const success = await authManager.deleteApiKey(keyId);
            if (success) {
                this.showSuccess('API key deleted successfully');
                await this.loadApiKeys();
            } else {
                this.showError('Failed to delete API key');
            }
        } catch (error) {
            console.error('Failed to delete API key:', error);
            this.showError('Failed to delete API key: ' + error.message);
        }
    }

    copyApiKey() {
        const apiKeyInput = document.getElementById('newApiKey');
        apiKeyInput.select();
        document.execCommand('copy');
        this.showSuccess('API key copied to clipboard!');
    }

    downloadFile(fileId) {
        const url = fileManager.getDownloadUrl(fileId);
        window.open(url, '_blank');
    }

    showSuccess(message) {
        this.showAlert(message, 'success');
    }

    showError(message) {
        this.showAlert(message, 'danger');
    }

    showAlert(message, type) {
        const alertColors = {
            success: 'bg-green-50 border-green-200 text-green-800',
            danger: 'bg-red-50 border-red-200 text-red-800',
            warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
            info: 'bg-blue-50 border-blue-200 text-blue-800'
        };
        
        const alertIcons = {
            success: 'fas fa-check-circle',
            danger: 'fas fa-exclamation-circle',
            warning: 'fas fa-exclamation-triangle',
            info: 'fas fa-info-circle'
        };
        
        const alert = document.createElement('div');
        alert.className = `fixed top-4 right-4 z-50 ${alertColors[type]} border rounded-lg p-4 flex items-center shadow-lg min-w-80`;
        alert.innerHTML = `
            <i class="${alertIcons[type]} mr-3"></i>
            <span class="flex-1">${message}</span>
            <button type="button" class="ml-3 text-current hover:opacity-70 transition-opacity" onclick="this.parentElement.remove()">
                <i class="fas fa-times"></i>
            </button>
        `;
        
        document.body.appendChild(alert);
        
        setTimeout(() => {
            if (alert.parentNode) {
                alert.remove();
            }
        }, 5000);
    }
}

// Global functions
function logout() {
    authManager.logout();
}

function showCreateApiKeyModal() {
    if (profilePage) {
        profilePage.showCreateApiKeyModal();
    }
}

function closeCreateApiKeyModal() {
    document.getElementById('createApiKeyModal').classList.add('hidden');
}

function closeApiKeyCreatedModal() {
    document.getElementById('apiKeyCreatedModal').classList.add('hidden');
}

function createApiKey() {
    if (profilePage) {
        profilePage.createApiKey();
    }
}

function copyApiKey() {
    if (profilePage) {
        profilePage.copyApiKey();
    }
}

// Initialize profile page
let profilePage;
document.addEventListener('DOMContentLoaded', () => {
    profilePage = new ProfilePage();
});