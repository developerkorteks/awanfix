// Files page functionality for RcloneStorage

class FilesPage {
    constructor() {
        this.files = [];
        this.filteredFiles = [];
        this.currentView = 'table';
        this.currentPage = 1;
        this.itemsPerPage = 10;
        this.currentSort = 'name';
        this.currentFilter = '';
        this.searchTerm = '';
        this.deleteModal = null;
        this.previewModal = null;
        this.init();
    }

    async init() {
        // Check authentication
        if (!authManager.isAuthenticated()) {
            window.location.href = '/login.html';
            return;
        }

        // Load user profile and files
        await this.loadUserProfile();
        await this.loadFiles();
        this.setupEventListeners();
    }

    async loadUserProfile() {
        try {
            const result = await authManager.getProfile();
            if (result.success) {
                const user = result.user;
                document.getElementById('userEmail').textContent = user.email;
                
                // Show admin menu if user is admin
                if (user.role === 'admin') {
                    document.getElementById('adminMenuItem').classList.remove('hidden');
                }
            }
        } catch (error) {
            console.error('Failed to load user profile:', error);
        }
    }

    async loadFiles() {
        try {
            document.getElementById('loadingState').classList.remove('hidden');
            document.getElementById('tableView').classList.add('hidden');
            document.getElementById('gridView').classList.add('hidden');
            document.getElementById('emptyState').classList.add('hidden');

            const filesData = await fileManager.getFiles();
            this.files = filesData.files || [];
            
            this.updateStatistics();
            this.applyFiltersAndSort();
            
        } catch (error) {
            console.error('Failed to load files:', error);
            this.showError('Failed to load files');
        } finally {
            document.getElementById('loadingState').classList.add('hidden');
        }
    }

    updateStatistics() {
        const totalFiles = this.files.length;
        const totalSize = this.files.reduce((sum, file) => sum + (file.size || 0), 0);
        const videoFiles = this.files.filter(file => this.getFileType(file) === 'video').length;
        const imageFiles = this.files.filter(file => this.getFileType(file) === 'image').length;

        document.getElementById('totalFilesCount').textContent = totalFiles;
        document.getElementById('totalSizeCount').textContent = fileManager.formatFileSize(totalSize);
        document.getElementById('videoFilesCount').textContent = videoFiles;
        document.getElementById('imageFilesCount').textContent = imageFiles;
    }

    getFileType(file) {
        const ext = file.name.split('.').pop()?.toLowerCase() || '';
        
        if (['mp4', 'mkv', 'avi', 'mov', 'wmv', 'flv', 'webm'].includes(ext)) {
            return 'video';
        } else if (['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp'].includes(ext)) {
            return 'image';
        } else if (['mp3', 'wav', 'flac', 'aac', 'ogg'].includes(ext)) {
            return 'audio';
        } else if (['pdf', 'doc', 'docx', 'txt', 'rtf'].includes(ext)) {
            return 'document';
        } else if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
            return 'archive';
        } else {
            return 'other';
        }
    }

    applyFiltersAndSort() {
        let filtered = [...this.files];

        // Apply search filter
        if (this.searchTerm) {
            filtered = filtered.filter(file => 
                file.name.toLowerCase().includes(this.searchTerm.toLowerCase())
            );
        }

        // Apply type filter
        if (this.currentFilter) {
            filtered = filtered.filter(file => this.getFileType(file) === this.currentFilter);
        }

        // Apply sorting
        filtered.sort((a, b) => {
            switch (this.currentSort) {
                case 'name':
                    return a.name.localeCompare(b.name);
                case 'size':
                    return (b.size || 0) - (a.size || 0);
                case 'date':
                    return new Date(b.modified || 0) - new Date(a.modified || 0);
                case 'type':
                    return this.getFileType(a).localeCompare(this.getFileType(b));
                default:
                    return 0;
            }
        });

        this.filteredFiles = filtered;
        this.currentPage = 1;
        this.renderFiles();
    }

    renderFiles() {
        if (this.filteredFiles.length === 0) {
            document.getElementById('emptyState').classList.remove('hidden');
            document.getElementById('tableView').classList.add('hidden');
            document.getElementById('gridView').classList.add('hidden');
            document.getElementById('paginationContainer').classList.add('hidden');
            return;
        }

        document.getElementById('emptyState').classList.add('hidden');
        
        if (this.currentView === 'table') {
            this.renderTableView();
        } else {
            this.renderGridView();
        }
        
        this.renderPagination();
    }

    renderTableView() {
        document.getElementById('tableView').classList.remove('hidden');
        document.getElementById('gridView').classList.add('hidden');

        const startIndex = (this.currentPage - 1) * this.itemsPerPage;
        const endIndex = startIndex + this.itemsPerPage;
        const pageFiles = this.filteredFiles.slice(startIndex, endIndex);

        const tbody = document.getElementById('filesTableBody');
        tbody.innerHTML = pageFiles.map(file => `
            <tr class="hover:bg-gray-50 transition-colors">
                <td class="px-6 py-4 whitespace-nowrap text-center">
                    <i class="${fileManager.getFileIcon(file.mime_type, file.name)} text-2xl"></i>
                </td>
                <td class="px-6 py-4 whitespace-nowrap">
                    <div class="flex flex-col">
                        <div class="font-medium text-gray-900 truncate max-w-xs" title="${file.name}">${file.name}</div>
                        <div class="text-sm text-gray-500">${this.getFileType(file)}</div>
                    </div>
                </td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${fileManager.formatFileSize(file.size || 0)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${fileManager.formatDate(file.modified)}</td>
                <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                    <div class="flex items-center space-x-2">
                        <button onclick="filesPage.downloadFile('${file.id}')" title="Download" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-50">
                            <i class="fas fa-download"></i>
                        </button>
                        <button onclick="filesPage.previewFile('${file.id}')" title="Preview" class="text-green-600 hover:text-green-900 transition-colors p-2 rounded-md hover:bg-green-50">
                            <i class="fas fa-eye"></i>
                        </button>
                        ${fileManager.isStreamable(file.mime_type, file.name) ? `
                            <button onclick="filesPage.streamFile('${file.id}')" title="Stream" class="text-purple-600 hover:text-purple-900 transition-colors p-2 rounded-md hover:bg-purple-50">
                                <i class="fas fa-play"></i>
                            </button>
                        ` : ''}
                        <button onclick="filesPage.confirmDelete('${file.id}', '${file.name}')" title="Delete" class="text-red-600 hover:text-red-900 transition-colors p-2 rounded-md hover:bg-red-50">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    }

    renderGridView() {
        document.getElementById('tableView').classList.add('hidden');
        document.getElementById('gridView').classList.remove('hidden');

        const startIndex = (this.currentPage - 1) * this.itemsPerPage;
        const endIndex = startIndex + this.itemsPerPage;
        const pageFiles = this.filteredFiles.slice(startIndex, endIndex);

        const container = document.getElementById('filesGridContainer');
        container.innerHTML = pageFiles.map(file => `
            <div class="bg-white rounded-xl shadow-sm border border-gray-200 hover:shadow-md transition-shadow overflow-hidden">
                <div class="p-6 text-center">
                    <div class="mb-4">
                        <i class="${fileManager.getFileIcon(file.mime_type, file.name)} text-4xl text-gray-400"></i>
                    </div>
                    <h3 class="font-medium text-gray-900 mb-2 truncate" title="${file.name}">
                        ${file.name.length > 25 ? file.name.substring(0, 25) + '...' : file.name}
                    </h3>
                    <div class="text-sm text-gray-500 space-y-1">
                        <div>${fileManager.formatFileSize(file.size || 0)}</div>
                        <div>${this.getFileType(file)}</div>
                        <div>${fileManager.formatDate(file.modified)}</div>
                    </div>
                </div>
                <div class="border-t border-gray-200 bg-gray-50 px-4 py-3">
                    <div class="flex justify-center space-x-2">
                        <button onclick="filesPage.downloadFile('${file.id}')" title="Download" class="text-blue-600 hover:text-blue-900 transition-colors p-2 rounded-md hover:bg-blue-100">
                            <i class="fas fa-download"></i>
                        </button>
                        <button onclick="filesPage.previewFile('${file.id}')" title="Preview" class="text-green-600 hover:text-green-900 transition-colors p-2 rounded-md hover:bg-green-100">
                            <i class="fas fa-eye"></i>
                        </button>
                        ${fileManager.isStreamable(file.mime_type, file.name) ? `
                            <button onclick="filesPage.streamFile('${file.id}')" title="Stream" class="text-purple-600 hover:text-purple-900 transition-colors p-2 rounded-md hover:bg-purple-100">
                                <i class="fas fa-play"></i>
                            </button>
                        ` : ''}
                        <button onclick="filesPage.confirmDelete('${file.id}', '${file.name}')" title="Delete" class="text-red-600 hover:text-red-900 transition-colors p-2 rounded-md hover:bg-red-100">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }

    renderPagination() {
        const totalPages = Math.ceil(this.filteredFiles.length / this.itemsPerPage);
        
        if (totalPages <= 1) {
            document.getElementById('paginationContainer').classList.add('hidden');
            return;
        }

        document.getElementById('paginationContainer').classList.remove('hidden');
        
        const pagination = document.getElementById('pagination');
        let paginationHTML = '';

        // Previous button
        paginationHTML += `
            <button onclick="filesPage.goToPage(${this.currentPage - 1})" 
                    ${this.currentPage === 1 ? 'disabled' : ''} 
                    class="relative inline-flex items-center px-4 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-l-md hover:bg-gray-50 ${this.currentPage === 1 ? 'cursor-not-allowed opacity-50' : 'hover:text-gray-700'}">
                <i class="fas fa-chevron-left mr-1"></i>
                Previous
            </button>
        `;

        // Page numbers
        for (let i = 1; i <= totalPages; i++) {
            if (i === 1 || i === totalPages || (i >= this.currentPage - 2 && i <= this.currentPage + 2)) {
                const isActive = i === this.currentPage;
                paginationHTML += `
                    <button onclick="filesPage.goToPage(${i})" 
                            class="relative inline-flex items-center px-4 py-2 text-sm font-medium ${isActive ? 'bg-primary text-white border-primary' : 'text-gray-700 bg-white border-gray-300 hover:bg-gray-50'} border-l-0">
                        ${i}
                    </button>
                `;
            } else if (i === this.currentPage - 3 || i === this.currentPage + 3) {
                paginationHTML += `<span class="relative inline-flex items-center px-4 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 border-l-0">...</span>`;
            }
        }

        // Next button
        paginationHTML += `
            <button onclick="filesPage.goToPage(${this.currentPage + 1})" 
                    ${this.currentPage === totalPages ? 'disabled' : ''} 
                    class="relative inline-flex items-center px-4 py-2 text-sm font-medium text-gray-500 bg-white border border-gray-300 rounded-r-md hover:bg-gray-50 border-l-0 ${this.currentPage === totalPages ? 'cursor-not-allowed opacity-50' : 'hover:text-gray-700'}">
                Next
                <i class="fas fa-chevron-right ml-1"></i>
            </button>
        `;

        pagination.innerHTML = paginationHTML;
    }

    setupEventListeners() {
        // Search input
        document.getElementById('searchInput').addEventListener('input', (e) => {
            this.searchTerm = e.target.value;
            this.applyFiltersAndSort();
        });

        // Type filter
        document.getElementById('typeFilter').addEventListener('change', (e) => {
            this.currentFilter = e.target.value;
            this.applyFiltersAndSort();
        });

        // Sort by
        document.getElementById('sortBy').addEventListener('change', (e) => {
            this.currentSort = e.target.value;
            this.applyFiltersAndSort();
        });

        // Delete confirmation
        document.getElementById('confirmDeleteBtn').addEventListener('click', () => {
            this.deleteFile();
        });
    }

    setView(view) {
        this.currentView = view;
        
        // Update button states
        document.getElementById('tableViewBtn').classList.toggle('active', view === 'table');
        document.getElementById('gridViewBtn').classList.toggle('active', view === 'grid');
        
        this.renderFiles();
    }

    goToPage(page) {
        const totalPages = Math.ceil(this.filteredFiles.length / this.itemsPerPage);
        if (page >= 1 && page <= totalPages) {
            this.currentPage = page;
            this.renderFiles();
        }
    }

    downloadFile(fileId) {
        const url = fileManager.getDownloadUrl(fileId);
        const link = document.createElement('a');
        link.href = url;
        link.download = '';
        
        // Add authorization header for authenticated download
        if (authManager.token) {
            // For downloads, we'll open in new tab with auth header
            window.open(url + '?token=' + authManager.token, '_blank');
        } else {
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        }
    }

    async previewFile(fileId) {
        try {
            const file = this.files.find(f => f.id === fileId);
            if (!file) return;

            document.getElementById('previewModalTitle').textContent = file.name;
            document.getElementById('downloadFromPreview').onclick = () => this.downloadFile(fileId);

            const modalBody = document.getElementById('previewModalBody');
            const fileType = this.getFileType(file);

            if (fileType === 'image') {
                modalBody.innerHTML = `
                    <div class="flex justify-center items-center bg-gray-100 rounded-lg p-4">
                        <img src="${fileManager.getDownloadUrl(fileId)}" 
                             class="max-w-full max-h-96 rounded-lg shadow-lg object-contain" 
                             alt="${file.name}"
                             onload="this.classList.add('animate-fade-in')"
                             onerror="this.parentElement.innerHTML='<div class=\\'text-center text-gray-500\\'>Failed to load image</div>'">
                    </div>
                `;
            } else if (fileType === 'video') {
                modalBody.innerHTML = `
                    <div class="bg-black rounded-lg overflow-hidden">
                        <video controls 
                               class="w-full h-auto max-h-96 rounded-lg" 
                               preload="metadata"
                               poster=""
                               style="background: #000;">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="video/mp4">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="video/webm">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="video/ogg">
                            <div class="text-center text-white p-8">
                                Your browser does not support the video tag.
                                <br>
                                <a href="${fileManager.getDownloadUrl(fileId)}" class="text-blue-400 hover:text-blue-300">Download video</a>
                            </div>
                        </video>
                    </div>
                `;
            } else if (fileType === 'audio') {
                modalBody.innerHTML = `
                    <div class="bg-gradient-to-r from-purple-100 to-pink-100 rounded-lg p-8">
                        <div class="text-center mb-6">
                            <i class="fas fa-music text-6xl text-purple-600 mb-4"></i>
                            <h3 class="text-lg font-semibold text-gray-900">${file.name}</h3>
                        </div>
                        <audio controls class="w-full rounded-lg shadow-md">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="audio/mpeg">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="audio/wav">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="audio/ogg">
                            <div class="text-center text-gray-600 p-4">
                                Your browser does not support the audio tag.
                                <br>
                                <a href="${fileManager.getDownloadUrl(fileId)}" class="text-blue-600 hover:text-blue-800">Download audio</a>
                            </div>
                        </audio>
                    </div>
                `;
            } else {
                modalBody.innerHTML = `
                    <div class="text-center py-8">
                        <div class="mb-6">
                            <i class="${fileManager.getFileIcon(file.mime_type, file.name)} text-6xl text-gray-400 mb-4"></i>
                        </div>
                        <h3 class="text-xl font-semibold text-gray-900 mb-4">${file.name}</h3>
                        <div class="bg-gray-50 rounded-lg p-4 mb-4 text-left max-w-md mx-auto">
                            <div class="space-y-2 text-sm">
                                <div class="flex justify-between">
                                    <span class="font-medium text-gray-600">Size:</span>
                                    <span class="text-gray-900">${fileManager.formatFileSize(file.size)}</span>
                                </div>
                                <div class="flex justify-between">
                                    <span class="font-medium text-gray-600">Type:</span>
                                    <span class="text-gray-900">${fileType}</span>
                                </div>
                                <div class="flex justify-between">
                                    <span class="font-medium text-gray-600">Modified:</span>
                                    <span class="text-gray-900">${fileManager.formatDate(file.modified)}</span>
                                </div>
                            </div>
                        </div>
                        <p class="text-gray-600">Preview not available for this file type.</p>
                        <button onclick="filesPage.downloadFile('${fileId}')" class="mt-4 bg-primary hover:bg-blue-700 text-white px-6 py-2 rounded-lg transition-colors">
                            <i class="fas fa-download mr-2"></i>Download File
                        </button>
                    </div>
                `;
            }

            document.getElementById('previewModal').classList.remove('hidden');
        } catch (error) {
            console.error('Failed to preview file:', error);
            this.showError('Failed to preview file');
        }
    }

    streamFile(fileId) {
        window.open(`/stream.html?id=${fileId}`, '_blank');
    }

    confirmDelete(fileId, fileName) {
        this.fileToDelete = fileId;
        document.getElementById('deleteFileInfo').innerHTML = `
            <strong>File:</strong> ${fileName}
        `;
        document.getElementById('deleteModal').classList.remove('hidden');
    }

    async deleteFile() {
        if (!this.fileToDelete) return;

        try {
            await fileManager.deleteFile(this.fileToDelete);
            document.getElementById('deleteModal').classList.add('hidden');
            this.showSuccess('File deleted successfully');
            await this.loadFiles(); // Reload files
        } catch (error) {
            console.error('Failed to delete file:', error);
            this.showError('Failed to delete file: ' + error.message);
        }
    }

    async refreshFiles() {
        await this.loadFiles();
        this.showSuccess('Files refreshed');
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

function refreshFiles() {
    if (filesPage) {
        filesPage.refreshFiles();
    }
}

function setView(viewType) {
    if (filesPage) {
        filesPage.currentView = viewType;
        filesPage.renderFiles();
        
        // Update button states
        const tableBtn = document.getElementById('tableViewBtn');
        const gridBtn = document.getElementById('gridViewBtn');
        
        if (viewType === 'table') {
            tableBtn.classList.add('bg-gray-100');
            gridBtn.classList.remove('bg-gray-100');
        } else {
            gridBtn.classList.add('bg-gray-100');
            tableBtn.classList.remove('bg-gray-100');
        }
    }
}

function closePreviewModal() {
    document.getElementById('previewModal').classList.add('hidden');
}

function closeDeleteModal() {
    document.getElementById('deleteModal').classList.add('hidden');
}

// Initialize files page
let filesPage;
document.addEventListener('DOMContentLoaded', () => {
    filesPage = new FilesPage();
});