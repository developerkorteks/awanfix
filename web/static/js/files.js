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

        // Initialize modals
        this.deleteModal = new bootstrap.Modal(document.getElementById('deleteModal'));
        this.previewModal = new bootstrap.Modal(document.getElementById('previewModal'));

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
                    document.getElementById('adminMenuItem').classList.remove('d-none');
                }
            }
        } catch (error) {
            console.error('Failed to load user profile:', error);
        }
    }

    async loadFiles() {
        try {
            document.getElementById('loadingState').style.display = 'block';
            document.getElementById('tableView').style.display = 'none';
            document.getElementById('gridView').style.display = 'none';
            document.getElementById('emptyState').style.display = 'none';

            const filesData = await fileManager.getFiles();
            this.files = filesData.files || [];
            
            this.updateStatistics();
            this.applyFiltersAndSort();
            
        } catch (error) {
            console.error('Failed to load files:', error);
            this.showError('Failed to load files');
        } finally {
            document.getElementById('loadingState').style.display = 'none';
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
            document.getElementById('emptyState').style.display = 'block';
            document.getElementById('tableView').style.display = 'none';
            document.getElementById('gridView').style.display = 'none';
            document.getElementById('paginationContainer').style.display = 'none';
            return;
        }

        document.getElementById('emptyState').style.display = 'none';
        
        if (this.currentView === 'table') {
            this.renderTableView();
        } else {
            this.renderGridView();
        }
        
        this.renderPagination();
    }

    renderTableView() {
        document.getElementById('tableView').style.display = 'block';
        document.getElementById('gridView').style.display = 'none';

        const startIndex = (this.currentPage - 1) * this.itemsPerPage;
        const endIndex = startIndex + this.itemsPerPage;
        const pageFiles = this.filteredFiles.slice(startIndex, endIndex);

        const tbody = document.getElementById('filesTableBody');
        tbody.innerHTML = pageFiles.map(file => `
            <tr>
                <td>
                    <i class="${fileManager.getFileIcon(file.mime_type, file.name)}"></i>
                </td>
                <td>
                    <div class="d-flex align-items-center">
                        <div>
                            <div class="fw-medium">${file.name}</div>
                            <small class="text-muted">${this.getFileType(file)}</small>
                        </div>
                    </div>
                </td>
                <td>${fileManager.formatFileSize(file.size || 0)}</td>
                <td>
                    <small>${fileManager.formatDate(file.modified)}</small>
                </td>
                <td>
                    <div class="btn-group btn-group-sm">
                        <button class="btn btn-outline-primary" onclick="filesPage.downloadFile('${file.id}')" title="Download">
                            <i class="fas fa-download"></i>
                        </button>
                        <button class="btn btn-outline-info" onclick="filesPage.previewFile('${file.id}')" title="Preview">
                            <i class="fas fa-eye"></i>
                        </button>
                        ${fileManager.isStreamable(file.mime_type, file.name) ? `
                            <button class="btn btn-outline-success" onclick="filesPage.streamFile('${file.id}')" title="Stream">
                                <i class="fas fa-play"></i>
                            </button>
                        ` : ''}
                        <button class="btn btn-outline-danger" onclick="filesPage.confirmDelete('${file.id}', '${file.name}')" title="Delete">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `).join('');
    }

    renderGridView() {
        document.getElementById('tableView').style.display = 'none';
        document.getElementById('gridView').style.display = 'block';

        const startIndex = (this.currentPage - 1) * this.itemsPerPage;
        const endIndex = startIndex + this.itemsPerPage;
        const pageFiles = this.filteredFiles.slice(startIndex, endIndex);

        const container = document.getElementById('filesGridContainer');
        container.innerHTML = pageFiles.map(file => `
            <div class="col-lg-3 col-md-4 col-sm-6 mb-4">
                <div class="card h-100">
                    <div class="card-body text-center">
                        <div class="mb-3">
                            <i class="${fileManager.getFileIcon(file.mime_type, file.name)} fa-3x"></i>
                        </div>
                        <h6 class="card-title" title="${file.name}">${file.name.length > 20 ? file.name.substring(0, 20) + '...' : file.name}</h6>
                        <p class="card-text">
                            <small class="text-muted">
                                ${fileManager.formatFileSize(file.size || 0)}<br>
                                ${fileManager.formatDate(file.modified)}
                            </small>
                        </p>
                    </div>
                    <div class="card-footer bg-transparent">
                        <div class="btn-group w-100" role="group">
                            <button class="btn btn-outline-primary btn-sm" onclick="filesPage.downloadFile('${file.id}')" title="Download">
                                <i class="fas fa-download"></i>
                            </button>
                            <button class="btn btn-outline-info btn-sm" onclick="filesPage.previewFile('${file.id}')" title="Preview">
                                <i class="fas fa-eye"></i>
                            </button>
                            ${fileManager.isStreamable(file.mime_type, file.name) ? `
                                <button class="btn btn-outline-success btn-sm" onclick="filesPage.streamFile('${file.id}')" title="Stream">
                                    <i class="fas fa-play"></i>
                                </button>
                            ` : ''}
                            <button class="btn btn-outline-danger btn-sm" onclick="filesPage.confirmDelete('${file.id}', '${file.name}')" title="Delete">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `).join('');
    }

    renderPagination() {
        const totalPages = Math.ceil(this.filteredFiles.length / this.itemsPerPage);
        
        if (totalPages <= 1) {
            document.getElementById('paginationContainer').style.display = 'none';
            return;
        }

        document.getElementById('paginationContainer').style.display = 'block';
        
        const pagination = document.getElementById('pagination');
        let paginationHTML = '';

        // Previous button
        paginationHTML += `
            <li class="page-item ${this.currentPage === 1 ? 'disabled' : ''}">
                <a class="page-link" href="#" onclick="filesPage.goToPage(${this.currentPage - 1})">Previous</a>
            </li>
        `;

        // Page numbers
        for (let i = 1; i <= totalPages; i++) {
            if (i === 1 || i === totalPages || (i >= this.currentPage - 2 && i <= this.currentPage + 2)) {
                paginationHTML += `
                    <li class="page-item ${i === this.currentPage ? 'active' : ''}">
                        <a class="page-link" href="#" onclick="filesPage.goToPage(${i})">${i}</a>
                    </li>
                `;
            } else if (i === this.currentPage - 3 || i === this.currentPage + 3) {
                paginationHTML += `<li class="page-item disabled"><span class="page-link">...</span></li>`;
            }
        }

        // Next button
        paginationHTML += `
            <li class="page-item ${this.currentPage === totalPages ? 'disabled' : ''}">
                <a class="page-link" href="#" onclick="filesPage.goToPage(${this.currentPage + 1})">Next</a>
            </li>
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
                    <div class="text-center">
                        <img src="${fileManager.getDownloadUrl(fileId)}" class="img-fluid" alt="${file.name}">
                    </div>
                `;
            } else if (fileType === 'video') {
                modalBody.innerHTML = `
                    <div class="text-center">
                        <video controls class="w-100" style="max-height: 400px;">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="video/mp4">
                            Your browser does not support the video tag.
                        </video>
                    </div>
                `;
            } else if (fileType === 'audio') {
                modalBody.innerHTML = `
                    <div class="text-center">
                        <audio controls class="w-100">
                            <source src="${fileManager.getStreamUrl(fileId)}" type="audio/mpeg">
                            Your browser does not support the audio tag.
                        </audio>
                    </div>
                `;
            } else {
                modalBody.innerHTML = `
                    <div class="text-center">
                        <i class="${fileManager.getFileIcon(file.mime_type, file.name)} fa-5x mb-3"></i>
                        <h5>${file.name}</h5>
                        <p class="text-muted">
                            Size: ${fileManager.formatFileSize(file.size)}<br>
                            Type: ${fileType}<br>
                            Modified: ${fileManager.formatDate(file.modified)}
                        </p>
                        <p>Preview not available for this file type.</p>
                    </div>
                `;
            }

            this.previewModal.show();
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
        this.deleteModal.show();
    }

    async deleteFile() {
        if (!this.fileToDelete) return;

        try {
            await fileManager.deleteFile(this.fileToDelete);
            this.deleteModal.hide();
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
        const alert = document.createElement('div');
        alert.className = `alert alert-${type} alert-dismissible fade show position-fixed`;
        alert.style.cssText = 'top: 20px; right: 20px; z-index: 1050; min-width: 300px;';
        alert.innerHTML = `
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        `;
        document.body.appendChild(alert);
        
        setTimeout(() => {
            if (alert.parentNode) {
                alert.parentNode.removeChild(alert);
            }
        }, 5000);
    }
}

// Global functions
function logout() {
    authManager.logout();
}

// Initialize files page
let filesPage;
document.addEventListener('DOMContentLoaded', () => {
    filesPage = new FilesPage();
});