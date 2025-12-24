/**
 * SIEM API Module
 * Handles all API requests with Basic Authentication
 */

const API = {
    baseUrl: '/api',
    
    /**
     * Get Basic Auth header from stored credentials
     */
    getAuthHeader() {
        const credentials = sessionStorage.getItem('siem_credentials');
        if (!credentials) {
            return null;
        }
        return `Basic ${credentials}`;
    },
    
    /**
     * Make an authenticated API request
     */
    async request(endpoint, options = {}) {
        const authHeader = this.getAuthHeader();
        
        if (!authHeader) {
            throw new Error('Not authenticated');
        }
        
        const config = {
            ...options,
            headers: {
                'Authorization': authHeader,
                'Content-Type': 'application/json',
                ...options.headers
            }
        };
        
        const response = await fetch(`${this.baseUrl}${endpoint}`, config);
        
        if (response.status === 401) {
            // Clear credentials and redirect to login
            sessionStorage.removeItem('siem_credentials');
            window.location.href = 'login.html';
            throw new Error('Unauthorized');
        }
        
        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error || `HTTP ${response.status}`);
        }
        
        return response.json();
    },
    
    /**
     * Test authentication with provided credentials
     */
    async testAuth(username, password) {
        const credentials = btoa(`${username}:${password}`);
        
        const response = await fetch(`${this.baseUrl}/health`, {
            headers: {
                'Authorization': `Basic ${credentials}`
            }
        });
        
        if (response.ok) {
            // Store credentials on success
            sessionStorage.setItem('siem_credentials', credentials);
            return true;
        }
        
        return false;
    },
    
    /**
     * Get dashboard statistics
     */
    async getStats() {
        return this.request('/stats');
    },
    
    /**
     * Get all events (with optional pagination)
     * @param {number} page - Page number (default 1)
     * @param {number} limit - Items per page (default 50)
     */
    async getEvents(page = 1, limit = 50) {
        return this.request(`/events?page=${page}&limit=${limit}`);
    },
    
    /**
     * Get all events without pagination (for filtering client-side)
     */
    async getAllEvents() {
        return this.request('/events?limit=200');
    },
    
    /**
     * Health check
     */
    async health() {
        return this.request('/health');
    }
};

// Export for use in other modules
window.API = API;
