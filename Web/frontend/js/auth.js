/**
 * SIEM Authentication Module
 * Handles login, logout, and auth state
 */

const Auth = {
    /**
     * Check if user is authenticated
     */
    isAuthenticated() {
        return sessionStorage.getItem('siem_credentials') !== null;
    },
    
    /**
     * Login with username and password
     */
    async login(username, password) {
        try {
            return await API.testAuth(username, password);
        } catch (error) {
            console.error('Login error:', error);
            return false;
        }
    },
    
    /**
     * Logout and redirect to login page
     */
    logout() {
        sessionStorage.removeItem('siem_credentials');
        window.location.href = 'login.html';
    },
    
    /**
     * Require authentication - redirect to login if not authenticated
     */
    requireAuth() {
        if (!this.isAuthenticated()) {
            window.location.href = 'login.html';
            return false;
        }
        return true;
    }
};

// Setup logout button handler
document.addEventListener('DOMContentLoaded', () => {
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => Auth.logout());
    }
});

// Export for use in other modules
window.Auth = Auth;
