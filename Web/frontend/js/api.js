const API = {
    baseUrl: '/api',
    
    getAuthHeader() {
        const credentials = sessionStorage.getItem('siem_credentials');
        if (!credentials) {
            return null;
        }
        return `Basic ${credentials}`;
    },
    
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
    
    async testAuth(username, password) {
        const credentials = btoa(`${username}:${password}`);
        
        const response = await fetch(`${this.baseUrl}/health`, {
            headers: {
                'Authorization': `Basic ${credentials}`
            }
        });
        
        if (response.ok) {
            sessionStorage.setItem('siem_credentials', credentials);
            return true;
        }
        
        return false;
    },
    
    async getStats() {
        return this.request('/stats');
    },
    
    async getEvents(page = 1, limit = 50) {
        return this.request(`/events?page=${page}&limit=${limit}`);
    },
    
    async getAllEvents() {
        return this.request('/events?limit=200');
    },
    
    async health() {
        return this.request('/health');
    }
};

window.API = API;
