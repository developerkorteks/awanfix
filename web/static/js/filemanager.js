// File Manager for RcloneStorage Web Interface

class FileManager {
    constructor(authManager) {
        this.auth = authManager;
        this.apiBaseUrl = '/api/v1';
    }

    // Upload file with progress callback
    async uploadFile(file, onProgress = null) {
        return new Promise((resolve, reject) => {
            const formData = new FormData();
            formData.append('file', file);

            const xhr = new XMLHttpRequest();

            // Progress tracking
            if (onProgress) {
                xhr.upload.addEventListener('progress', (e) => {
                    if (e.lengthComputable) {
                        const percentComplete = (e.loaded / e.total) * 100;
                        onProgress(percentComplete);
                    }
                });
            }

            // Response handling
            xhr.onload = () => {
                if (xhr.status === 200) {
                    try {
                        const response = JSON.parse(xhr.responseText);
                        resolve(response);
                    } catch (error) {
                        reject(new Error('Invalid response format'));
                    }
                } else {
                    try {
                        const error = JSON.parse(xhr.responseText);
                        reject(new Error(error.error || 'Upload failed'));
                    } catch (e) {
                        reject(new Error(`Upload failed with status ${xhr.status}`));
                    }
                }
            };

            xhr.onerror = () => {
                reject(new Error('Network error during upload'));
            };

            // Send request
            xhr.open('POST', `${this.apiBaseUrl}/upload`);
            xhr.setRequestHeader('Authorization', `Bearer ${this.auth.token}`);
            xhr.send(formData);
        });
    }

    // Get list of files
    async getFiles() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/files`, {
                headers: this.auth.getAuthHeaders()
            });

            if (response.ok) {
                return await response.json();
            } else {
                this.auth.handleApiError(response);
                throw new Error('Failed to get files');
            }
        } catch (error) {
            throw error;
        }
    }

    // Get file information
    async getFileInfo(fileId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/files/${fileId}`, {
                headers: this.auth.getAuthHeaders()
            });

            if (response.ok) {
                return await response.json();
            } else {
                throw new Error('Failed to get file info');
            }
        } catch (error) {
            throw error;
        }
    }

    // Delete file
    async deleteFile(fileId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/files/${fileId}`, {
                method: 'DELETE',
                headers: this.auth.getAuthHeaders()
            });

            if (response.ok) {
                return await response.json();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to delete file');
            }
        } catch (error) {
            throw error;
        }
    }

    // Get download URL
    getDownloadUrl(fileId) {
        return `${this.apiBaseUrl}/download/${fileId}`;
    }

    // Get streaming URL
    getStreamUrl(fileId) {
        return `${this.apiBaseUrl}/stream/${fileId}`;
    }

    // Get stream info
    async getStreamInfo(fileId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/stream/${fileId}/info`, {
                headers: this.auth.getAuthHeaders()
            });

            if (response.ok) {
                return await response.json();
            } else {
                throw new Error('Failed to get stream info');
            }
        } catch (error) {
            throw error;
        }
    }

    // Format file size
    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // Get file icon based on MIME type
    getFileIcon(mimeType, filename = '') {
        const ext = filename.split('.').pop()?.toLowerCase() || '';
        
        if (mimeType?.startsWith('video/') || ['mp4', 'mkv', 'avi', 'mov', 'wmv'].includes(ext)) {
            return 'fas fa-video file-icon-video';
        } else if (mimeType?.startsWith('image/') || ['jpg', 'jpeg', 'png', 'gif', 'bmp'].includes(ext)) {
            return 'fas fa-image file-icon-image';
        } else if (mimeType?.startsWith('audio/') || ['mp3', 'wav', 'flac', 'aac'].includes(ext)) {
            return 'fas fa-music file-icon-audio';
        } else if (mimeType?.includes('pdf') || ['pdf', 'doc', 'docx', 'txt'].includes(ext)) {
            return 'fas fa-file-alt file-icon-document';
        } else if (['zip', 'rar', '7z', 'tar', 'gz'].includes(ext)) {
            return 'fas fa-file-archive file-icon-archive';
        } else {
            return 'fas fa-file file-icon-default';
        }
    }

    // Check if file is streamable
    isStreamable(mimeType, filename = '') {
        const ext = filename.split('.').pop()?.toLowerCase() || '';
        const streamableTypes = ['video/', 'audio/'];
        const streamableExts = ['mp4', 'mkv', 'avi', 'mov', 'mp3', 'wav', 'flac'];
        
        return streamableTypes.some(type => mimeType?.startsWith(type)) || 
               streamableExts.includes(ext);
    }

    // Format date
    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diffTime = Math.abs(now - date);
        const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

        if (diffDays === 1) {
            return 'Today';
        } else if (diffDays === 2) {
            return 'Yesterday';
        } else if (diffDays <= 7) {
            return `${diffDays - 1} days ago`;
        } else {
            return date.toLocaleDateString();
        }
    }
}

// Global file manager instance
window.fileManager = new FileManager(window.authManager);