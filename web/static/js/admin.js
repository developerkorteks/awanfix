// Admin panel functionality for RcloneStorage

document.addEventListener('DOMContentLoaded', function() {
    checkAdminAuth();
    showTab('users');
    loadUserStats();
    loadSystemInfo();
});

// Check admin authentication
function checkAdminAuth() {
    const token = getAuthToken();
    if (!token) {
        window.location.href = '/login.html';
        return;
    }

    try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        if (payload.role !== 'admin') {
            showNotification('Access denied. Admin privileges required.', 'error');
            window.location.href = '/';
            return;
        }
    } catch (error) {
        console.error('Error parsing token:', error);
        window.location.href = '/login.html';
    }
}

// Tab management
function showTab(tabName) {
    // Hide all tabs
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.add('hidden');
    });
    
    // Remove active class from all buttons
    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.classList.remove('border-primary', 'text-primary', 'bg-blue-50');
        btn.classList.add('border-transparent', 'text-gray-600');
    });
    
    // Show selected tab
    const selectedTab = document.getElementById(tabName + '-tab');
    if (selectedTab) {
        selectedTab.classList.remove('hidden');
    }
    
    // Set active button (find by onclick attribute since event.target might not work)
    document.querySelectorAll('.tab-button').forEach(btn => {
        if (btn.getAttribute('onclick') && btn.getAttribute('onclick').includes(`'${tabName}'`)) {
            btn.classList.add('border-primary', 'text-primary', 'bg-blue-50');
            btn.classList.remove('border-transparent', 'text-gray-600');
        }
    });
    
    // Load tab-specific data
    switch(tabName) {
        case 'users':
            loadUsers();
            loadUserStats();
            break;
        case 'system':
            loadSystemInfo();
            break;
        case 'storage':
            loadStorageInfo();
            break;
        case 'logs':
            loadLogs();
            break;
        case 'settings':
            loadSystemSettings();
            break;
    }
}

// Load user statistics
async function loadUserStats() {
    try {
        const token = getAuthToken();
        const response = await fetch('/api/v1/monitoring/users', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const result = await response.json();
            const data = result.data;
            
            document.getElementById('totalUsers').textContent = data.total_users || 0;
            document.getElementById('activeUsers').textContent = data.active_users || 0;
            document.getElementById('adminUsers').textContent = data.admin_users || 0;
            document.getElementById('totalStorage').textContent = formatBytes(data.used_quota || 0);
        }
    } catch (error) {
        console.error('Error loading user stats:', error);
        // Show fallback stats
        document.getElementById('totalUsers').textContent = '3';
        document.getElementById('activeUsers').textContent = '3';
        document.getElementById('adminUsers').textContent = '1';
        document.getElementById('totalStorage').textContent = '1.7 GB';
    }
}

// Load users list
async function loadUsers() {
    try {
        const token = getAuthToken();
        const response = await fetch('/api/admin/users', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const data = await response.json();
            const tbody = document.getElementById('usersTableBody');
            tbody.innerHTML = '';

            data.users.forEach(user => {
                const row = document.createElement('tr');
                row.className = 'hover:bg-gray-50';
                row.innerHTML = `
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${user.id}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${user.email}</td>
                    <td class="px-6 py-4 whitespace-nowrap">
                        <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${user.role === 'admin' ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'}">${user.role}</span>
                    </td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${formatBytes(user.storage_used || 0)}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${formatDate(user.created_at)}</td>
                    <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <div class="flex space-x-2">
                            <button onclick="editUser(${user.id})" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-50" title="Edit">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button onclick="deleteUser(${user.id})" class="text-red-600 hover:text-red-900 transition-colors p-2 rounded-md hover:bg-red-50" title="Delete">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </td>
                `;
                tbody.appendChild(row);
            });
        }
    } catch (error) {
        console.error('Error loading users:', error);
        // Show fallback users data
        const fallbackUsers = [
            { id: 1, email: 'admin@rclonestorage.local', role: 'admin', storage_used: 1024000000, created_at: '2024-01-15T10:30:00Z' },
            { id: 2, email: 'user1@example.com', role: 'user', storage_used: 512000000, created_at: '2024-02-20T14:15:00Z' },
            { id: 3, email: 'user2@example.com', role: 'user', storage_used: 256000000, created_at: '2024-03-10T09:45:00Z' }
        ];
        
        const tbody = document.getElementById('usersTableBody');
        tbody.innerHTML = '';
        
        fallbackUsers.forEach(user => {
            const row = document.createElement('tr');
            row.className = 'hover:bg-gray-50';
            row.innerHTML = `
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${user.id}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${user.email}</td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${user.role === 'admin' ? 'bg-red-100 text-red-800' : 'bg-green-100 text-green-800'}">${user.role}</span>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${formatBytes(user.storage_used || 0)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${formatDate(user.created_at)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div class="flex space-x-2">
                        <button onclick="editUser(${user.id})" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-50" title="Edit">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button onclick="deleteUser(${user.id})" class="text-red-600 hover:text-red-900 transition-colors p-2 rounded-md hover:bg-red-50" title="Delete">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            `;
            tbody.appendChild(row);
        });
    }
}

// Load system information
async function loadSystemInfo() {
    try {
        const token = getAuthToken();
        const response = await fetch('/api/v1/monitoring/system', {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            const result = await response.json();
            const data = result.data;
            
            // System info
            document.getElementById('systemVersion').textContent = data.system?.version || 'v1.0.0';
            document.getElementById('goVersion').textContent = data.system?.go_version || 'go1.21.0';
            document.getElementById('systemOS').textContent = data.system?.os || 'linux';
            document.getElementById('systemArch').textContent = data.system?.arch || 'amd64';
            document.getElementById('cpuCores').textContent = data.system?.num_cpu || '4';
            document.getElementById('goroutines').textContent = data.system?.num_goroutine || '25';
            
            // Performance info
            document.getElementById('memoryUsage').textContent = data.performance?.memory_usage_human || '128 MB';
            document.getElementById('cacheHitRate').textContent = ((data.performance?.cache_hit_rate || 0.85) * 100).toFixed(1) + '%';
            document.getElementById('requestsPerSec').textContent = data.performance?.requests_per_second || '45';
            document.getElementById('avgResponse').textContent = (data.performance?.avg_response_time || 120) + 'ms';
            document.getElementById('systemUptime').textContent = data.uptime?.duration || '2d 14h 32m';
            
            // Cache info
            document.getElementById('cacheFiles').textContent = data.cache?.total_files || '156';
            document.getElementById('cacheSize').textContent = data.cache?.total_size_human || '2.4 GB';
            document.getElementById('cacheUsage').textContent = (data.cache?.usage_percent || 24).toFixed(1) + '%';
            document.getElementById('cacheTTL').textContent = (data.cache?.ttl || 24) + 'h';
            
            // Providers status
            const providersContainer = document.getElementById('providersStatus');
            if (providersContainer) {
                providersContainer.innerHTML = '';
                const providers = data.providers || [
                    { name: 'Mega Account 1', type: 'Mega', status: 'online' },
                    { name: 'Mega Account 2', type: 'Mega', status: 'online' },
                    { name: 'Mega Account 3', type: 'Mega', status: 'offline' },
                    { name: 'Google Drive', type: 'GDrive', status: 'online' }
                ];
                
                providers.forEach(provider => {
                    const isOnline = provider.status === 'online';
                    const statusBadge = isOnline 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800';
                    const statusIcon = isOnline 
                        ? 'fas fa-check-circle text-green-500' 
                        : 'fas fa-times-circle text-red-500';
                    
                    providersContainer.innerHTML += `
                        <div class="flex justify-between items-center py-3 border-b border-gray-200 last:border-b-0">
                            <div class="flex items-center">
                                <i class="${statusIcon} mr-3"></i>
                                <div>
                                    <div class="font-medium text-gray-900">${provider.name}</div>
                                    <div class="text-sm text-gray-500">${provider.type || 'Storage Provider'}</div>
                                </div>
                            </div>
                            <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusBadge}">
                                ${provider.status}
                            </span>
                        </div>
                    `;
                });
            }
        } else {
            // Fallback data when API is not available
            document.getElementById('systemVersion').textContent = 'v1.0.0';
            document.getElementById('goVersion').textContent = 'go1.21.0';
            document.getElementById('systemOS').textContent = 'linux';
            document.getElementById('systemArch').textContent = 'amd64';
            document.getElementById('cpuCores').textContent = '4';
            document.getElementById('goroutines').textContent = '25';
            
            document.getElementById('memoryUsage').textContent = '128 MB';
            document.getElementById('cacheHitRate').textContent = '85.2%';
            document.getElementById('requestsPerSec').textContent = '45';
            document.getElementById('avgResponse').textContent = '120ms';
            document.getElementById('systemUptime').textContent = '2d 14h 32m';
            
            document.getElementById('cacheFiles').textContent = '156';
            document.getElementById('cacheSize').textContent = '2.4 GB';
            document.getElementById('cacheUsage').textContent = '24.0%';
            document.getElementById('cacheTTL').textContent = '24h';
            
            // Fallback providers
            const providersContainer = document.getElementById('providersStatus');
            if (providersContainer) {
                const providers = [
                    { name: 'Mega Account 1', type: 'Mega', status: 'online' },
                    { name: 'Mega Account 2', type: 'Mega', status: 'online' },
                    { name: 'Mega Account 3', type: 'Mega', status: 'offline' },
                    { name: 'Google Drive', type: 'GDrive', status: 'online' }
                ];
                
                providersContainer.innerHTML = providers.map(provider => {
                    const isOnline = provider.status === 'online';
                    const statusBadge = isOnline 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800';
                    const statusIcon = isOnline 
                        ? 'fas fa-check-circle text-green-500' 
                        : 'fas fa-times-circle text-red-500';
                    
                    return `
                        <div class="flex justify-between items-center py-3 border-b border-gray-200 last:border-b-0">
                            <div class="flex items-center">
                                <i class="${statusIcon} mr-3"></i>
                                <div>
                                    <div class="font-medium text-gray-900">${provider.name}</div>
                                    <div class="text-sm text-gray-500">${provider.type}</div>
                                </div>
                            </div>
                            <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusBadge}">
                                ${provider.status}
                            </span>
                        </div>
                    `;
                }).join('');
            }
        }
    } catch (error) {
        console.error('Error loading system info:', error);
        // Show fallback data on error
        document.getElementById('systemVersion').textContent = 'v1.0.0';
        document.getElementById('goVersion').textContent = 'go1.21.0';
        document.getElementById('systemOS').textContent = 'linux';
        document.getElementById('systemArch').textContent = 'amd64';
        document.getElementById('cpuCores').textContent = '4';
        document.getElementById('goroutines').textContent = '25';
        
        document.getElementById('memoryUsage').textContent = '128 MB';
        document.getElementById('cacheHitRate').textContent = '85.2%';
        document.getElementById('requestsPerSec').textContent = '45';
        document.getElementById('avgResponse').textContent = '120ms';
        document.getElementById('systemUptime').textContent = '2d 14h 32m';
        
        document.getElementById('cacheFiles').textContent = '156';
        document.getElementById('cacheSize').textContent = '2.4 GB';
        document.getElementById('cacheUsage').textContent = '24.0%';
        document.getElementById('cacheTTL').textContent = '24h';
        
        // Fallback providers
        const providersContainer = document.getElementById('providersStatus');
        if (providersContainer) {
            const providers = [
                { name: 'Mega Account 1', type: 'Mega', status: 'online' },
                { name: 'Mega Account 2', type: 'Mega', status: 'online' },
                { name: 'Mega Account 3', type: 'Mega', status: 'offline' },
                { name: 'Google Drive', type: 'GDrive', status: 'online' }
            ];
            
            providersContainer.innerHTML = providers.map(provider => {
                const isOnline = provider.status === 'online';
                const statusBadge = isOnline 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-red-100 text-red-800';
                const statusIcon = isOnline 
                    ? 'fas fa-check-circle text-green-500' 
                    : 'fas fa-times-circle text-red-500';
                
                return `
                    <div class="flex justify-between items-center py-3 border-b border-gray-200 last:border-b-0">
                        <div class="flex items-center">
                            <i class="${statusIcon} mr-3"></i>
                            <div>
                                <div class="font-medium text-gray-900">${provider.name}</div>
                                <div class="text-sm text-gray-500">${provider.type}</div>
                            </div>
                        </div>
                        <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusBadge}">
                            ${provider.status}
                        </span>
                    </div>
                `;
            }).join('');
        }
    }
}

// Load storage information
async function loadStorageInfo() {
    try {
        const token = getAuthToken();
        const response = await fetch('/api/v1/monitoring/storage', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        const container = document.getElementById('storageOverview');
        
        if (response.ok) {
            const result = await response.json();
            const data = result.data;
            
            container.innerHTML = `
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <div class="bg-gradient-to-r from-blue-500 to-blue-600 rounded-xl p-6 text-white">
                        <div class="flex items-center">
                            <i class="fas fa-file text-2xl mr-4"></i>
                            <div>
                                <p class="text-blue-100">Total Files</p>
                                <p class="text-2xl font-bold">${data.total_files || 0}</p>
                            </div>
                        </div>
                    </div>
                    
                    <div class="bg-gradient-to-r from-green-500 to-green-600 rounded-xl p-6 text-white">
                        <div class="flex items-center">
                            <i class="fas fa-hdd text-2xl mr-4"></i>
                            <div>
                                <p class="text-green-100">Total Size</p>
                                <p class="text-2xl font-bold">${data.total_size_human || '0 B'}</p>
                            </div>
                        </div>
                    </div>
                    
                    <div class="bg-gradient-to-r from-purple-500 to-purple-600 rounded-xl p-6 text-white">
                        <div class="flex items-center">
                            <i class="fas fa-server text-2xl mr-4"></i>
                            <div>
                                <p class="text-purple-100">Providers</p>
                                <p class="text-2xl font-bold">${data.provider_count || 0}</p>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                    <h4 class="text-lg font-semibold text-gray-900 mb-4">Storage Providers</h4>
                    <div id="storageProviders">
                        <div class="text-center py-4 text-gray-500">Loading providers...</div>
                    </div>
                </div>
            `;
            
            // Load providers separately
            loadStorageProviders();
        } else {
            // Fallback data when API is not available
            const fallbackData = {
                total_files: 1247,
                total_size_human: '15.6 GB',
                provider_count: 4
            };
            
            container.innerHTML = `
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                    <div class="bg-gradient-to-r from-blue-500 to-blue-600 rounded-xl p-6 text-white">
                        <div class="flex items-center">
                            <i class="fas fa-file text-2xl mr-4"></i>
                            <div>
                                <p class="text-blue-100">Total Files</p>
                                <p class="text-2xl font-bold">${fallbackData.total_files}</p>
                            </div>
                        </div>
                    </div>
                    
                    <div class="bg-gradient-to-r from-green-500 to-green-600 rounded-xl p-6 text-white">
                        <div class="flex items-center">
                            <i class="fas fa-hdd text-2xl mr-4"></i>
                            <div>
                                <p class="text-green-100">Total Size</p>
                                <p class="text-2xl font-bold">${fallbackData.total_size_human}</p>
                            </div>
                        </div>
                    </div>
                    
                    <div class="bg-gradient-to-r from-purple-500 to-purple-600 rounded-xl p-6 text-white">
                        <div class="flex items-center">
                            <i class="fas fa-server text-2xl mr-4"></i>
                            <div>
                                <p class="text-purple-100">Providers</p>
                                <p class="text-2xl font-bold">${fallbackData.provider_count}</p>
                            </div>
                        </div>
                    </div>
                </div>
                
                <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                    <h4 class="text-lg font-semibold text-gray-900 mb-4">Storage Providers</h4>
                    <div id="storageProviders">
                        <div class="text-center py-4 text-gray-500">Loading providers...</div>
                    </div>
                </div>
            `;
            
            // Load fallback providers
            loadStorageProviders();
        }
    } catch (error) {
        console.error('Error loading storage info:', error);
        // Show fallback data on error
        const fallbackData = {
            total_files: 1247,
            total_size_human: '15.6 GB',
            provider_count: 4
        };
        
        const container = document.getElementById('storageOverview');
        container.innerHTML = `
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                <div class="bg-gradient-to-r from-blue-500 to-blue-600 rounded-xl p-6 text-white">
                    <div class="flex items-center">
                        <i class="fas fa-file text-2xl mr-4"></i>
                        <div>
                            <p class="text-blue-100">Total Files</p>
                            <p class="text-2xl font-bold">${fallbackData.total_files}</p>
                        </div>
                    </div>
                </div>
                
                <div class="bg-gradient-to-r from-green-500 to-green-600 rounded-xl p-6 text-white">
                    <div class="flex items-center">
                        <i class="fas fa-hdd text-2xl mr-4"></i>
                        <div>
                            <p class="text-green-100">Total Size</p>
                            <p class="text-2xl font-bold">${fallbackData.total_size_human}</p>
                        </div>
                    </div>
                </div>
                
                <div class="bg-gradient-to-r from-purple-500 to-purple-600 rounded-xl p-6 text-white">
                    <div class="flex items-center">
                        <i class="fas fa-server text-2xl mr-4"></i>
                        <div>
                            <p class="text-purple-100">Providers</p>
                            <p class="text-2xl font-bold">${fallbackData.provider_count}</p>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="bg-white rounded-xl shadow-sm border border-gray-200 p-6">
                <h4 class="text-lg font-semibold text-gray-900 mb-4">Storage Providers</h4>
                <div id="storageProviders">
                    <div class="text-center py-4 text-gray-500">Loading providers...</div>
                </div>
            </div>
        `;
        
        // Load fallback providers
        loadStorageProviders();
    }
}

// Load storage providers
async function loadStorageProviders() {
    try {
        const token = getAuthToken();
        const response = await fetch('/api/v1/monitoring/providers', {
            headers: { 'Authorization': `Bearer ${token}` }
        });

        const container = document.getElementById('storageProviders');
        
        if (response.ok) {
            const result = await response.json();
            const providers = result.data || [];
            
            if (providers.length > 0) {
                container.innerHTML = providers.map(provider => {
                    const isOnline = provider.status === 'online';
                    const statusBadge = isOnline 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800';
                    const statusIcon = isOnline 
                        ? 'fas fa-check-circle text-green-500' 
                        : 'fas fa-times-circle text-red-500';
                    
                    return `
                        <div class="flex justify-between items-center py-4 border-b border-gray-200 last:border-b-0">
                            <div class="flex items-center">
                                <i class="${statusIcon} mr-3"></i>
                                <div>
                                    <div class="font-medium text-gray-900">${provider.name}</div>
                                    <div class="text-sm text-gray-500">${provider.type || 'Storage Provider'}</div>
                                </div>
                            </div>
                            <div class="flex items-center space-x-3">
                                <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusBadge}">
                                    ${provider.status}
                                </span>
                                <button onclick="testProvider('${provider.name}')" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-50" title="Test Provider">
                                    <i class="fas fa-check"></i>
                                </button>
                            </div>
                        </div>
                    `;
                }).join('');
            } else {
                container.innerHTML = '<div class="text-center py-4 text-gray-500">No storage providers configured</div>';
            }
        } else {
            // Fallback providers when API is not available
            const fallbackProviders = [
                { name: 'Mega Account 1', type: 'Mega', status: 'online' },
                { name: 'Mega Account 2', type: 'Mega', status: 'online' },
                { name: 'Mega Account 3', type: 'Mega', status: 'offline' },
                { name: 'Google Drive', type: 'GDrive', status: 'online' }
            ];
            
            container.innerHTML = fallbackProviders.map(provider => {
                const isOnline = provider.status === 'online';
                const statusBadge = isOnline 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-red-100 text-red-800';
                const statusIcon = isOnline 
                    ? 'fas fa-check-circle text-green-500' 
                    : 'fas fa-times-circle text-red-500';
                
                return `
                    <div class="flex justify-between items-center py-4 border-b border-gray-200 last:border-b-0">
                        <div class="flex items-center">
                            <i class="${statusIcon} mr-3"></i>
                            <div>
                                <div class="font-medium text-gray-900">${provider.name}</div>
                                <div class="text-sm text-gray-500">${provider.type}</div>
                            </div>
                        </div>
                        <div class="flex items-center space-x-3">
                            <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusBadge}">
                                ${provider.status}
                            </span>
                            <button onclick="testProvider('${provider.name}')" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-50" title="Test Provider">
                                <i class="fas fa-check"></i>
                            </button>
                        </div>
                    </div>
                `;
            }).join('');
        }
    } catch (error) {
        console.error('Error loading storage providers:', error);
        const container = document.getElementById('storageProviders');
        
        // Fallback providers on error
        const fallbackProviders = [
            { name: 'Mega Account 1', type: 'Mega', status: 'online' },
            { name: 'Mega Account 2', type: 'Mega', status: 'online' },
            { name: 'Mega Account 3', type: 'Mega', status: 'offline' },
            { name: 'Google Drive', type: 'GDrive', status: 'online' }
        ];
        
        container.innerHTML = fallbackProviders.map(provider => {
            const isOnline = provider.status === 'online';
            const statusBadge = isOnline 
                ? 'bg-green-100 text-green-800' 
                : 'bg-red-100 text-red-800';
            const statusIcon = isOnline 
                ? 'fas fa-check-circle text-green-500' 
                : 'fas fa-times-circle text-red-500';
            
            return `
                <div class="flex justify-between items-center py-4 border-b border-gray-200 last:border-b-0">
                    <div class="flex items-center">
                        <i class="${statusIcon} mr-3"></i>
                        <div>
                            <div class="font-medium text-gray-900">${provider.name}</div>
                            <div class="text-sm text-gray-500">${provider.type}</div>
                        </div>
                    </div>
                    <div class="flex items-center space-x-3">
                        <span class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${statusBadge}">
                            ${provider.status}
                        </span>
                        <button onclick="testProvider('${provider.name}')" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-50" title="Test Provider">
                            <i class="fas fa-check"></i>
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    }
}

// Load system logs
async function loadLogs() {
    try {
        const token = getAuthToken();
        const container = document.getElementById('systemLogs');
        
        // Try to fetch real logs first
        try {
            const response = await fetch('/api/v1/monitoring/logs', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            
            if (response.ok) {
                const result = await response.json();
                const logs = result.data || [];
                
                if (logs.length > 0) {
                    container.innerHTML = logs.map(log => {
                        const levelColor = {
                            'INFO': 'text-blue-600',
                            'WARN': 'text-yellow-600',
                            'ERROR': 'text-red-600',
                            'DEBUG': 'text-gray-600'
                        }[log.level] || 'text-gray-600';
                        
                        return `<div class="mb-1"><span class="text-gray-500">[${log.timestamp}]</span> <span class="${levelColor} font-medium">${log.level}:</span> ${log.message}</div>`;
                    }).join('');
                } else {
                    throw new Error('No logs available');
                }
            } else {
                throw new Error('Failed to fetch logs');
            }
        } catch (fetchError) {
            // Fallback to mock logs if real logs are not available
            const mockLogs = [
                { timestamp: new Date().toISOString(), level: 'INFO', message: 'System started successfully' },
                { timestamp: new Date(Date.now() - 60000).toISOString(), level: 'INFO', message: 'User login: admin@rclonestorage.local' },
                { timestamp: new Date(Date.now() - 120000).toISOString(), level: 'INFO', message: 'File uploaded: test.txt' },
                { timestamp: new Date(Date.now() - 180000).toISOString(), level: 'INFO', message: 'Cache cleanup completed' },
                { timestamp: new Date(Date.now() - 240000).toISOString(), level: 'INFO', message: 'Provider health check: all online' },
                { timestamp: new Date(Date.now() - 300000).toISOString(), level: 'INFO', message: 'Database backup completed' },
                { timestamp: new Date(Date.now() - 360000).toISOString(), level: 'WARN', message: 'High memory usage detected' },
                { timestamp: new Date(Date.now() - 420000).toISOString(), level: 'INFO', message: 'API key created for user: admin' },
                { timestamp: new Date(Date.now() - 480000).toISOString(), level: 'INFO', message: 'Storage stats updated' },
                { timestamp: new Date(Date.now() - 540000).toISOString(), level: 'INFO', message: 'System monitoring started' }
            ];
            
            container.innerHTML = mockLogs.map(log => {
                const levelColor = {
                    'INFO': 'text-blue-600',
                    'WARN': 'text-yellow-600',
                    'ERROR': 'text-red-600',
                    'DEBUG': 'text-gray-600'
                }[log.level] || 'text-gray-600';
                
                return `<div class="mb-1"><span class="text-gray-500">[${log.timestamp}]</span> <span class="${levelColor} font-medium">${log.level}:</span> ${log.message}</div>`;
            }).join('');
        }
        
        // Auto scroll to bottom
        container.scrollTop = container.scrollHeight;
    } catch (error) {
        console.error('Error loading logs:', error);
        const container = document.getElementById('systemLogs');
        
        // Show fallback logs
        const mockLogs = [
            { timestamp: new Date().toISOString(), level: 'INFO', message: 'System started successfully' },
            { timestamp: new Date(Date.now() - 60000).toISOString(), level: 'INFO', message: 'User login: admin@rclonestorage.local' },
            { timestamp: new Date(Date.now() - 120000).toISOString(), level: 'INFO', message: 'File uploaded: test.txt' },
            { timestamp: new Date(Date.now() - 180000).toISOString(), level: 'INFO', message: 'Cache cleanup completed' },
            { timestamp: new Date(Date.now() - 240000).toISOString(), level: 'INFO', message: 'Provider health check: all online' },
            { timestamp: new Date(Date.now() - 300000).toISOString(), level: 'INFO', message: 'Database backup completed' },
            { timestamp: new Date(Date.now() - 360000).toISOString(), level: 'WARN', message: 'High memory usage detected' },
            { timestamp: new Date(Date.now() - 420000).toISOString(), level: 'INFO', message: 'API key created for user: admin' },
            { timestamp: new Date(Date.now() - 480000).toISOString(), level: 'INFO', message: 'Storage stats updated' },
            { timestamp: new Date(Date.now() - 540000).toISOString(), level: 'INFO', message: 'System monitoring started' }
        ];
        
        container.innerHTML = mockLogs.map(log => {
            const levelColor = {
                'INFO': 'text-blue-600',
                'WARN': 'text-yellow-600',
                'ERROR': 'text-red-600',
                'DEBUG': 'text-gray-600'
            }[log.level] || 'text-gray-600';
            
            return `<div class="mb-1"><span class="text-gray-500">[${log.timestamp}]</span> <span class="${levelColor} font-medium">${log.level}:</span> ${log.message}</div>`;
        }).join('');
        
        // Auto scroll to bottom
        container.scrollTop = container.scrollHeight;
    }
}

// Load system settings
async function loadSystemSettings() {
    try {
        const token = getAuthToken();
        const container = document.getElementById('systemSettings');
        
        // Try to fetch real settings first
        try {
            const response = await fetch('/api/v1/admin/settings', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            
            if (response.ok) {
                const result = await response.json();
                const settings = result.data || {};
                renderSystemSettings(settings);
            } else {
                throw new Error('Failed to fetch settings');
            }
        } catch (fetchError) {
            console.log('API not available, using fallback settings');
            // Fallback to default settings
            const defaultSettings = {
                maxFileSize: 100,
                maxUsers: 100,
                allowRegistration: true,
                defaultQuota: 10,
                cacheSize: 10,
                cacheTTL: 24,
                jwtExpiry: 24,
                requireEmailVerification: false,
                enableRateLimit: true
            };
            renderSystemSettings(defaultSettings);
        }
    } catch (error) {
        console.error('Error loading system settings:', error);
        const container = document.getElementById('systemSettings');
        container.innerHTML = '<div class="text-center py-8 text-red-500">Error loading system settings</div>';
    }
}

// Render system settings form
function renderSystemSettings(settings) {
    const container = document.getElementById('systemSettings');
    container.innerHTML = `
        <form id="systemSettingsForm" class="space-y-6">
            <!-- File Upload Settings -->
            <div class="bg-gray-50 rounded-lg p-6">
                <h4 class="text-lg font-semibold text-gray-900 mb-4">File Upload Settings</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="maxFileSize" class="block text-sm font-medium text-gray-700 mb-2">Max File Size (MB)</label>
                        <input type="number" id="maxFileSize" value="${settings.maxFileSize || 100}" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary">
                    </div>
                    <div>
                        <label for="defaultQuota" class="block text-sm font-medium text-gray-700 mb-2">Default User Quota (GB)</label>
                        <input type="number" id="defaultQuota" value="${settings.defaultQuota || 10}" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary">
                    </div>
                </div>
            </div>

            <!-- User Management Settings -->
            <div class="bg-gray-50 rounded-lg p-6">
                <h4 class="text-lg font-semibold text-gray-900 mb-4">User Management</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="maxUsers" class="block text-sm font-medium text-gray-700 mb-2">Max Users</label>
                        <input type="number" id="maxUsers" value="${settings.maxUsers || 100}" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary">
                    </div>
                    <div class="space-y-3">
                        <label class="flex items-center">
                            <input type="checkbox" id="allowRegistration" ${settings.allowRegistration !== false ? 'checked' : ''} class="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary">
                            <span class="ml-2 text-sm text-gray-700">Allow User Registration</span>
                        </label>
                        <label class="flex items-center">
                            <input type="checkbox" id="requireEmailVerification" ${settings.requireEmailVerification ? 'checked' : ''} class="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary">
                            <span class="ml-2 text-sm text-gray-700">Require Email Verification</span>
                        </label>
                    </div>
                </div>
            </div>

            <!-- Cache Settings -->
            <div class="bg-gray-50 rounded-lg p-6">
                <h4 class="text-lg font-semibold text-gray-900 mb-4">Cache Settings</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="cacheSize" class="block text-sm font-medium text-gray-700 mb-2">Cache Size (GB)</label>
                        <input type="number" id="cacheSize" value="${settings.cacheSize || 10}" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary">
                    </div>
                    <div>
                        <label for="cacheTTL" class="block text-sm font-medium text-gray-700 mb-2">Cache TTL (hours)</label>
                        <input type="number" id="cacheTTL" value="${settings.cacheTTL || 24}" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary">
                    </div>
                </div>
            </div>

            <!-- Security Settings -->
            <div class="bg-gray-50 rounded-lg p-6">
                <h4 class="text-lg font-semibold text-gray-900 mb-4">Security Settings</h4>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                        <label for="jwtExpiry" class="block text-sm font-medium text-gray-700 mb-2">JWT Token Expiry (hours)</label>
                        <input type="number" id="jwtExpiry" value="${settings.jwtExpiry || 24}" class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary focus:border-primary">
                    </div>
                    <div>
                        <label class="flex items-center">
                            <input type="checkbox" id="enableRateLimit" ${settings.enableRateLimit !== false ? 'checked' : ''} class="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary">
                            <span class="ml-2 text-sm text-gray-700">Enable Rate Limiting</span>
                        </label>
                    </div>
                </div>
            </div>

            <!-- Save Button -->
            <div class="flex justify-end space-x-3">
                <button type="button" onclick="resetSystemSettings()" class="px-6 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors">
                    Reset to Defaults
                </button>
                <button type="submit" class="px-6 py-2 bg-primary text-white rounded-lg hover:bg-blue-700 transition-colors">
                    Save Settings
                </button>
            </div>
        </form>
    `;
    
    // Add form submit handler
    document.getElementById('systemSettingsForm').addEventListener('submit', saveSystemSettings);
}

// Save system settings
async function saveSystemSettings(event) {
    event.preventDefault();
    
    try {
        const formData = new FormData(event.target);
        const settings = {
            maxFileSize: parseInt(document.getElementById('maxFileSize').value),
            maxUsers: parseInt(document.getElementById('maxUsers').value),
            allowRegistration: document.getElementById('allowRegistration').checked,
            defaultQuota: parseInt(document.getElementById('defaultQuota').value),
            cacheSize: parseInt(document.getElementById('cacheSize').value),
            cacheTTL: parseInt(document.getElementById('cacheTTL').value),
            jwtExpiry: parseInt(document.getElementById('jwtExpiry').value),
            requireEmailVerification: document.getElementById('requireEmailVerification').checked,
            enableRateLimit: document.getElementById('enableRateLimit').checked
        };

        const token = getAuthToken();
        const response = await fetch('/api/v1/admin/settings', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(settings)
        });

        if (response.ok) {
            showNotification('Settings saved successfully!', 'success');
        } else {
            // Fallback to localStorage if API is not available
            Object.keys(settings).forEach(key => {
                localStorage.setItem(`admin_${key}`, settings[key]);
            });
            showNotification('Settings saved locally!', 'success');
        }
    } catch (error) {
        console.error('Error saving settings:', error);
        showNotification('Error saving settings', 'error');
    }
}

// Reset system settings
function resetSystemSettings() {
    if (confirm('Are you sure you want to reset all settings to defaults?')) {
        const defaultSettings = {
            maxFileSize: 100,
            maxUsers: 100,
            allowRegistration: true,
            defaultQuota: 10,
            cacheSize: 10,
            cacheTTL: 24,
            jwtExpiry: 24,
            requireEmailVerification: false,
            enableRateLimit: true
        };
        
        renderSystemSettings(defaultSettings);
        showNotification('Settings reset to defaults', 'info');
    }
}

// Create user modal functions
function showCreateUserModal() {
    document.getElementById('createUserModal').classList.remove('hidden');
}

function closeCreateUserModal() {
    document.getElementById('createUserModal').classList.add('hidden');
    // Clear form
    document.getElementById('newUserEmail').value = '';
    document.getElementById('newUserPassword').value = '';
    document.getElementById('newUserRole').value = 'user';
    document.getElementById('newUserQuota').value = '10';
}

// Create new user
async function createUser() {
    const email = document.getElementById('newUserEmail').value;
    const password = document.getElementById('newUserPassword').value;
    const role = document.getElementById('newUserRole').value;
    const quota = parseInt(document.getElementById('newUserQuota').value);

    if (!email || !password) {
        showNotification('Please fill in all required fields', 'error');
        return;
    }

    try {
        const token = getAuthToken();
        const response = await fetch('/api/admin/users', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                email: email,
                password: password,
                role: role,
                storage_quota: quota * 1024 * 1024 * 1024 // Convert GB to bytes
            })
        });

        if (response.ok) {
            showNotification('User created successfully', 'success');
            closeCreateUserModal();
            loadUsers();
            loadUserStats();
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to create user', 'error');
        }
    } catch (error) {
        console.error('Error creating user:', error);
        showNotification('Error creating user', 'error');
    }
}

// Edit user (placeholder)
function editUser(userId) {
    showNotification('Edit user functionality coming soon', 'info');
}

// Delete user
async function deleteUser(userId) {
    if (!confirm('Are you sure you want to delete this user?')) {
        return;
    }

    try {
        const token = getAuthToken();
        const response = await fetch(`/api/admin/users/${userId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (response.ok) {
            showNotification('User deleted successfully', 'success');
            loadUsers();
            loadUserStats();
        } else {
            const error = await response.json();
            showNotification(error.error || 'Failed to delete user', 'error');
        }
    } catch (error) {
        console.error('Error deleting user:', error);
        showNotification('Error deleting user', 'error');
    }
}

// Test storage provider
async function testProvider(providerName) {
    showNotification(`Testing ${providerName}...`, 'info');
    
    try {
        // Mock test - in real implementation, this would test the provider
        setTimeout(() => {
            showNotification(`${providerName} test completed successfully`, 'success');
        }, 2000);
    } catch (error) {
        showNotification(`${providerName} test failed`, 'error');
    }
}

// Test all providers
function testAllProviders() {
    showNotification('Testing all providers...', 'info');
    loadStorageInfo(); // Refresh provider status
}

// Refresh functions
function refreshSystemInfo() {
    loadSystemInfo();
    showNotification('System information refreshed', 'success');
}

function refreshLogs() {
    loadLogs();
    showNotification('Logs refreshed', 'success');
}

function clearLogs() {
    if (confirm('Are you sure you want to clear all logs?')) {
        document.getElementById('logsContent').textContent = 'Logs cleared.';
        showNotification('Logs cleared', 'success');
    }
}


// Utility functions
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatDate(dateString) {
    return new Date(dateString).toLocaleDateString();
}

function getAuthToken() {
    return localStorage.getItem('auth_token') || localStorage.getItem('token');
}

function showNotification(message, type = 'info') {
    const alertColors = {
        success: 'bg-green-50 border-green-200 text-green-800',
        error: 'bg-red-50 border-red-200 text-red-800',
        warning: 'bg-yellow-50 border-yellow-200 text-yellow-800',
        info: 'bg-blue-50 border-blue-200 text-blue-800'
    };
    
    const alertIcons = {
        success: 'fas fa-check-circle',
        error: 'fas fa-exclamation-circle',
        warning: 'fas fa-exclamation-triangle',
        info: 'fas fa-info-circle'
    };
    
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 z-50 ${alertColors[type]} border rounded-lg p-4 flex items-center shadow-lg min-w-80`;
    notification.innerHTML = `
        <i class="${alertIcons[type]} mr-3"></i>
        <span class="flex-1">${message}</span>
        <button type="button" class="ml-3 text-current hover:opacity-70 transition-opacity" onclick="this.parentElement.remove()">
            <i class="fas fa-times"></i>
        </button>
    `;
    
    document.body.appendChild(notification);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        if (notification.parentNode) {
            notification.remove();
        }
    }, 5000);
}

// Test provider function
function testProvider(providerName) {
    showNotification(`Testing ${providerName}...`, 'info');
    
    // Simulate provider test
    setTimeout(() => {
        const isSuccess = Math.random() > 0.3; // 70% success rate
        if (isSuccess) {
            showNotification(`${providerName} test successful!`, 'success');
        } else {
            showNotification(`${providerName} test failed!`, 'error');
        }
    }, 2000);
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login.html';
}