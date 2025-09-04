// Authentication Manager for RcloneStorage Web Interface

class AuthManager {
    constructor() {
        this.token = localStorage.getItem('auth_token');
        this.user = JSON.parse(localStorage.getItem('user_data') || '{}');
        this.apiBaseUrl = '/api';
    }

    // Check if user is authenticated
    isAuthenticated() {
        return !!this.token;
    }

    // Check if user is admin
    isAdmin() {
        return this.user.role === 'admin';
    }

    // Get authentication headers
    getAuthHeaders() {
        const headers = {
            'Content-Type': 'application/json'
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }

    // Login user
    async login(email, password) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/auth/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email, password })
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.token;
                this.user = data.user;
                
                // Store in localStorage
                localStorage.setItem('auth_token', this.token);
                localStorage.setItem('user_data', JSON.stringify(this.user));
                
                return { success: true, user: this.user };
            } else {
                const error = await response.json();
                return { success: false, error: error.error || 'Login failed' };
            }
        } catch (error) {
            return { success: false, error: 'Network error' };
        }
    }

    // Register new user
    async register(email, password, role = 'user') {
        try {
            const response = await fetch(`${this.apiBaseUrl}/auth/register`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ email, password, role })
            });

            if (response.ok) {
                const data = await response.json();
                return { success: true, message: data.message };
            } else {
                const error = await response.json();
                return { success: false, error: error.error || 'Registration failed' };
            }
        } catch (error) {
            return { success: false, error: 'Network error' };
        }
    }

    // Get user profile
    async getProfile() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/user/profile`, {
                headers: this.getAuthHeaders()
            });

            if (response.ok) {
                const user = await response.json();
                this.user = user;
                localStorage.setItem('user_data', JSON.stringify(this.user));
                return { success: true, user };
            } else {
                return { success: false, error: 'Failed to get profile' };
            }
        } catch (error) {
            return { success: false, error: 'Network error' };
        }
    }

    // Logout user
    logout() {
        this.token = null;
        this.user = {};
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_data');
        window.location.href = '/login.html';
    }

    // Refresh token
    async refreshToken() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/auth/refresh`, {
                method: 'POST',
                headers: this.getAuthHeaders()
            });

            if (response.ok) {
                const data = await response.json();
                this.token = data.token;
                localStorage.setItem('auth_token', this.token);
                return true;
            }
        } catch (error) {
            console.error('Token refresh failed:', error);
        }
        
        return false;
    }

    // Handle API errors (401 = unauthorized)
    handleApiError(response) {
        if (response.status === 401) {
            this.logout();
        }
    }

    // Create API key
    async createApiKey(name) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/user/api-keys`, {
                method: 'POST',
                headers: this.getAuthHeaders(),
                body: JSON.stringify({ name })
            });

            if (response.ok) {
                return await response.json();
            } else {
                const error = await response.json();
                throw new Error(error.error || 'Failed to create API key');
            }
        } catch (error) {
            throw error;
        }
    }

    // List API keys
    async listApiKeys() {
        try {
            const response = await fetch(`${this.apiBaseUrl}/user/api-keys`, {
                headers: this.getAuthHeaders()
            });

            if (response.ok) {
                return await response.json();
            } else {
                throw new Error('Failed to list API keys');
            }
        } catch (error) {
            throw error;
        }
    }

    // Delete API key
    async deleteApiKey(keyId) {
        try {
            const response = await fetch(`${this.apiBaseUrl}/user/api-keys/${keyId}`, {
                method: 'DELETE',
                headers: this.getAuthHeaders()
            });

            return response.ok;
        } catch (error) {
            return false;
        }
    }
}

// Global auth manager instance
window.authManager = new AuthManager();