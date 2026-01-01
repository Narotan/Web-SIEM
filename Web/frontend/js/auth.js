const Auth = {
    isAuthenticated() {
        return sessionStorage.getItem('siem_credentials') !== null;
    },
    
    async login(username, password) {
        try {
            return await API.testAuth(username, password);
        } catch (error) {
            console.error('Login error:', error);
            return false;
        }
    },
    
    logout() {
        sessionStorage.removeItem('siem_credentials');
        window.location.href = 'login.html';
    },
    
    requireAuth() {
        if (!this.isAuthenticated()) {
            window.location.href = 'login.html';
            return false;
        }
        return true;
    }
};

document.addEventListener('DOMContentLoaded', () => {
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => Auth.logout());
    }
});

window.Auth = Auth;
