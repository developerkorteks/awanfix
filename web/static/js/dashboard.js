// Dashboard functionality for RcloneStorage

class Dashboard {
    constructor() {
        this.storageChart = null;
        this.init();
    }

    async init() {
        // Check authentication
        if (!authManager.isAuthenticated()) {
            window.location.href = '/login.html';
            return;
        }

        // Load user data and dashboard content
        await this.loadUserProfile();
        await this.loadDashboardData();
        this.setupEventListeners();
    }

    async loadUserProfile() {
        try {
            const result = await authManager.getProfile();
            if (result.success) {
                const user = result.user;
                document.getElementById('userEmail').textContent = user.email;
                document.getElementById('userName').textContent = user.email.split('@')[0];
                
                // Show admin menu if user is admin
                if (user.role === 'admin') {
                    document.getElementById('adminMenuItem').classList.remove('d-none');
                }
            }
        } catch (error) {
            console.error('Failed to load user profile:', error);
        }
    }

    async loadDashboardData() {
        try {
            // Load files
            const filesData = await fileManager.getFiles();
            this.updateFileStats(filesData);
            this.renderRecentFiles(filesData.files || []);

            // Load user profile for storage info
            const profileResult = await authManager.getProfile();
            if (profileResult.success) {
                this.updateStorageStats(profileResult.user);
                this.renderStorageChart(profileResult.user);
            }

        } catch (error) {
            console.error('Failed to load dashboard data:', error);
            this.showError('Failed to load dashboard data');
        }
    }

    updateFileStats(filesData) {
        const totalFiles = filesData.total || 0;
        const totalSize = filesData.total_size || 0;
        
        document.getElementById('totalFiles').textContent = totalFiles;
        document.getElementById('storageUsed').textContent = fileManager.formatFileSize(totalSize);
    }

    updateStorageStats(user) {
        const usagePercent = user.usage_percent || 0;
        const quotaText = user.storage_quota === -1 ? 'Unlimited' : fileManager.formatFileSize(user.storage_quota);
        
        document.getElementById('usagePercent').textContent = `${usagePercent.toFixed(1)}%`;
        
        // Update storage used with quota info
        const storageUsedElement = document.getElementById('storageUsed');
        const currentText = storageUsedElement.textContent;
        storageUsedElement.innerHTML = `${currentText}<br><small class="text-muted">of ${quotaText}</small>`;
    }

    renderStorageChart(user) {
        const ctx = document.getElementById('storageChart').getContext('2d');
        
        // Destroy existing chart if it exists
        if (this.storageChart) {
            this.storageChart.destroy();
        }

        const used = user.storage_used || 0;
        const quota = user.storage_quota === -1 ? used * 2 : user.storage_quota; // Show some available space for unlimited
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

    renderRecentFiles(files) {
        const container = document.getElementById('recentFiles');
        
        if (!files || files.length === 0) {
            container.innerHTML = `
                <div class="text-center text-muted">
                    <i class="fas fa-folder-open fa-2x mb-2"></i>
                    <p>No files yet. <a href="/upload.html">Upload your first file</a></p>
                </div>
            `;
            return;
        }

        // Show only the 5 most recent files
        const recentFiles = files.slice(0, 5);
        
        container.innerHTML = recentFiles.map(file => `
            <div class="d-flex align-items-center mb-2 p-2 border rounded">
                <div class="me-3">
                    <i class="${fileManager.getFileIcon(file.mime_type, file.name)}"></i>
                </div>
                <div class="flex-grow-1">
                    <div class="fw-medium">${file.name}</div>
                    <small class="text-muted">${fileManager.formatFileSize(file.size)} • ${fileManager.formatDate(file.modified)}</small>
                </div>
                <div class="ms-2">
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-primary" onclick="downloadFile('${file.id}')" title="Download">
                            <i class="fas fa-download"></i>
                        </button>
                        ${fileManager.isStreamable(file.mime_type, file.name) ? `
                            <button class="btn btn-outline-success" onclick="streamFile('${file.id}')" title="Stream">
                                <i class="fas fa-play"></i>
                            </button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `).join('');
    }

    setupEventListeners() {
        // Auto-refresh every 5 minutes
        setInterval(() => {
            this.loadDashboardData();
        }, 5 * 60 * 1000);
    }

    showError(message) {
        // You can implement a toast notification system here
        console.error(message);
    }
}

// Global functions
function logout() {
    authManager.logout();
}

function refreshData() {
    window.location.reload();
}

function downloadFile(fileId) {
    const url = fileManager.getDownloadUrl(fileId);
    const link = document.createElement('a');
    link.href = url;
    link.download = '';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
}

function streamFile(fileId) {
    window.open(`/stream.html?id=${fileId}`, '_blank');
}

// Initialize dashboard when page loads
document.addEventListener('DOMContentLoaded', () => {
    new Dashboard();
});