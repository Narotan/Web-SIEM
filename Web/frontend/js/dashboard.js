const charts = {
    eventsType: null,
    severity: null,
    processes: null,
    timeline: null
};

function initChartDefaults() {
    if (typeof Chart !== 'undefined') {
        Chart.defaults.color = '#94a3b8';
        Chart.defaults.borderColor = '#1e293b';
        Chart.defaults.font.family = "'JetBrains Mono', monospace";
    }
}

const colors = {
    info: '#06b6d4',
    success: '#10b981',
    warning: '#f59e0b',
    danger: '#ef4444',
    purple: '#8b5cf6',
    cyan: '#06b6d4',
    pink: '#ec4899',
    orange: '#f97316'
};

const severityColors = {
    low: '#10b981',
    medium: '#f59e0b',
    high: '#f97316',
    critical: '#ef4444'
};

async function initDashboard() {
    if (!Auth.requireAuth()) {
        return;
    }
    
    await waitForChartJs();
    initChartDefaults();
    
    await loadDashboard();
    
    setInterval(loadDashboard, 30000);
}

function waitForChartJs() {
    return new Promise((resolve) => {
        if (typeof Chart !== 'undefined') {
            resolve();
            return;
        }
        const checkInterval = setInterval(() => {
            if (typeof Chart !== 'undefined') {
                clearInterval(checkInterval);
                resolve();
            }
        }, 50);
        setTimeout(() => {
            clearInterval(checkInterval);
            resolve();
        }, 10000);
    });
}

async function loadDashboard() {
    const updateIndicator = document.getElementById('updateIndicator');
    updateIndicator?.classList.add('loading');
    
    try {
        const stats = await API.getStats();
        renderDashboard(stats);
        updateLastRefresh();
    } catch (error) {
        console.error('Failed to load dashboard:', error);
        showErrorState();
    } finally {
        updateIndicator?.classList.remove('loading');
    }
}

function updateLastRefresh() {
    const lastUpdate = document.getElementById('lastUpdate');
    if (lastUpdate) {
        const now = new Date();
        lastUpdate.textContent = `Обновление: ${now.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}`;
    }
}

function renderDashboard(stats) {
    renderAgents(stats.active_agents);
    renderLastLogins(stats.last_logins);
    renderEventsTypeChart(stats.events_by_type);
    renderSeverityChart(stats.severity_distribution);
    renderHostsList(stats.active_agents);
    renderProcessesChart(stats.top_processes);
    renderUsersList(stats.top_users);
    renderTimelineChart(stats.events_per_hour);
}

function renderAgents(agents) {
    const container = document.getElementById('agentsList');
    if (!container) return;
    
    if (!agents || Object.keys(agents).length === 0) {
        container.innerHTML = '<div class="no-data">Нет активных агентов</div>';
        return;
    }
    
    const now = new Date();
    const agentEntries = Object.entries(agents).sort((a, b) => new Date(b[1]) - new Date(a[1]));
    
    const html = `
        <div class="agent-list">
            ${agentEntries.map(([id, lastSeen]) => {
                const lastSeenDate = new Date(lastSeen);
                const isOnline = (now - lastSeenDate) < 5 * 60 * 1000;
                const timeAgo = formatTimeAgo(lastSeenDate);
                
                return `
                    <div class="agent-item">
                        <span class="agent-status ${isOnline ? 'online' : 'offline'}"></span>
                        <div class="agent-info">
                            <div class="agent-id">${escapeHtml(id)}</div>
                            <div class="agent-time">${timeAgo}</div>
                        </div>
                    </div>
                `;
            }).join('')}
        </div>
    `;
    
    container.innerHTML = html;
}

function renderLastLogins(logins) {
    const tbody = document.getElementById('loginsBody');
    if (!tbody) return;
    
    if (!logins || logins.length === 0) {
        tbody.innerHTML = '<tr><td colspan="4" class="no-data">Нет данных для анализа</td></tr>';
        return;
    }
    
    const html = logins.map(login => {
        const timestamp = formatTimestamp(login.timestamp);
        const user = login.user || '-';
        const isSuccess = login.event_type === 'user_login';
        const ip = login.source_ip || login.ip || '-';
        
        return `
            <tr>
                <td class="mono">${timestamp}</td>
                <td class="mono">${escapeHtml(user)}</td>
                <td><span class="status-badge ${isSuccess ? 'success' : 'failure'}">${isSuccess ? 'SUCCESS' : 'FAILURE'}</span></td>
                <td class="mono">${escapeHtml(ip)}</td>
            </tr>
        `;
    }).join('');
    
    tbody.innerHTML = html;
}

function renderEventsTypeChart(eventsByType) {
    const canvas = document.getElementById('eventsTypeChart');
    const noData = document.getElementById('eventsTypeNoData');
    if (!canvas || !noData) return;
    
    if (!eventsByType || Object.keys(eventsByType).length === 0) {
        canvas.classList.add('hidden');
        noData.classList.remove('hidden');
        return;
    }
    
    canvas.classList.remove('hidden');
    noData.classList.add('hidden');
    
    const labels = Object.keys(eventsByType);
    const data = Object.values(eventsByType);
    const chartColors = [colors.info, colors.success, colors.warning, colors.danger, colors.purple, colors.cyan, colors.pink, colors.orange];
    
    if (charts.eventsType) {
        charts.eventsType.data.labels = labels;
        charts.eventsType.data.datasets[0].data = data;
        charts.eventsType.update();
    } else {
        charts.eventsType = new Chart(canvas, {
            type: 'doughnut',
            data: {
                labels: labels,
                datasets: [{
                    data: data,
                    backgroundColor: chartColors.slice(0, labels.length),
                    borderWidth: 0
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        position: 'right',
                        labels: {
                            padding: 12,
                            usePointStyle: true,
                            pointStyle: 'circle'
                        }
                    }
                }
            }
        });
    }
}

function renderSeverityChart(severityDist) {
    const canvas = document.getElementById('severityChart');
    const noData = document.getElementById('severityNoData');
    if (!canvas || !noData) return;
    
    if (!severityDist || Object.keys(severityDist).length === 0) {
        canvas.classList.add('hidden');
        noData.classList.remove('hidden');
        return;
    }
    
    canvas.classList.remove('hidden');
    noData.classList.add('hidden');
    
    const orderedKeys = ['low', 'medium', 'high', 'critical'];
    const labels = orderedKeys.filter(k => severityDist[k] !== undefined);
    const data = labels.map(k => severityDist[k]);
    const backgroundColors = labels.map(k => severityColors[k]);
    
    if (charts.severity) {
        charts.severity.data.labels = labels.map(l => l.charAt(0).toUpperCase() + l.slice(1));
        charts.severity.data.datasets[0].data = data;
        charts.severity.data.datasets[0].backgroundColor = backgroundColors;
        charts.severity.update();
    } else {
        charts.severity = new Chart(canvas, {
            type: 'bar',
            data: {
                labels: labels.map(l => l.charAt(0).toUpperCase() + l.slice(1)),
                datasets: [{
                    data: data,
                    backgroundColor: backgroundColors,
                    borderWidth: 0,
                    borderRadius: 4
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: '#21262d'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        }
                    }
                }
            }
        });
    }
}

function renderHostsList(agents) {
    const container = document.getElementById('hostsList');
    if (!container) return;
    
    if (!agents || Object.keys(agents).length === 0) {
        container.innerHTML = '<div class="no-data">Нет активных хостов</div>';
        return;
    }
    
    const hosts = Object.keys(agents).map(agentId => {
        const hostname = agentId.split('-')[0] || agentId;
        return { name: hostname, agent: agentId };
    });
    
    const html = `
        <div class="host-list">
            ${hosts.slice(0, 5).map(host => `
                <div class="host-item">
                    <span class="host-name">${escapeHtml(host.name)}</span>
                    <span class="host-count">●</span>
                </div>
            `).join('')}
        </div>
    `;
    
    container.innerHTML = html;
}

function renderProcessesChart(processes) {
    const canvas = document.getElementById('processesChart');
    const noData = document.getElementById('processesNoData');
    if (!canvas || !noData) return;
    
    if (!processes || Object.keys(processes).length === 0) {
        canvas.classList.add('hidden');
        noData.classList.remove('hidden');
        return;
    }
    
    canvas.classList.remove('hidden');
    noData.classList.add('hidden');
    
    const sorted = Object.entries(processes)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5);
    
    const labels = sorted.map(([name]) => name);
    const data = sorted.map(([, count]) => count);
    
    if (charts.processes) {
        charts.processes.data.labels = labels;
        charts.processes.data.datasets[0].data = data;
        charts.processes.update();
    } else {
        charts.processes = new Chart(canvas, {
            type: 'bar',
            data: {
                labels: labels,
                datasets: [{
                    data: data,
                    backgroundColor: colors.info,
                    borderWidth: 0,
                    borderRadius: 4
                }]
            },
            options: {
                indexAxis: 'y',
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    x: {
                        beginAtZero: true,
                        grid: {
                            color: '#21262d'
                        }
                    },
                    y: {
                        grid: {
                            display: false
                        },
                        ticks: {
                            font: {
                                family: "'JetBrains Mono', monospace"
                            }
                        }
                    }
                }
            }
        });
    }
}

function renderUsersList(users) {
    const container = document.getElementById('usersList');
    if (!container) return;
    
    if (!users || Object.keys(users).length === 0) {
        container.innerHTML = '<div class="no-data">Нет данных для анализа</div>';
        return;
    }
    
    const sorted = Object.entries(users)
        .sort((a, b) => b[1] - a[1])
        .slice(0, 5);
    
    const html = `
        <div class="user-list">
            ${sorted.map(([name, count]) => `
                <a href="events.html?search=${encodeURIComponent(name)}&regex=true" class="user-item clickable">
                    <span class="user-name">${escapeHtml(name)}</span>
                    <span class="user-count">${count}</span>
                </a>
            `).join('')}
        </div>
    `;
    
    container.innerHTML = html;
}

function renderTimelineChart(eventsPerHour) {
    const canvas = document.getElementById('timelineChart');
    const noData = document.getElementById('timelineNoData');
    if (!canvas || !noData) return;
    
    if (!eventsPerHour || Object.keys(eventsPerHour).length === 0) {
        canvas.classList.add('hidden');
        noData.classList.remove('hidden');
        return;
    }
    
    canvas.classList.remove('hidden');
    noData.classList.add('hidden');
    
    const labels = [];
    const data = [];
    for (let i = 0; i < 24; i++) {
        labels.push(`${i.toString().padStart(2, '0')}:00`);
        data.push(eventsPerHour[i] || 0);
    }
    
    if (charts.timeline) {
        charts.timeline.data.labels = labels;
        charts.timeline.data.datasets[0].data = data;
        charts.timeline.update();
    } else {
        charts.timeline = new Chart(canvas, {
            type: 'line',
            data: {
                labels: labels,
                datasets: [{
                    data: data,
                    borderColor: colors.info,
                    backgroundColor: 'transparent',
                    borderWidth: 2,
                    tension: 0.3,
                    pointRadius: 0,
                    pointHoverRadius: 4,
                    pointHoverBackgroundColor: colors.info
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: {
                        display: false
                    }
                },
                scales: {
                    y: {
                        beginAtZero: true,
                        grid: {
                            color: '#21262d'
                        }
                    },
                    x: {
                        grid: {
                            display: false
                        },
                        ticks: {
                            maxTicksLimit: 12
                        }
                    }
                },
                interaction: {
                    intersect: false,
                    mode: 'index'
                }
            }
        });
    }
}

function showErrorState() {
    const containers = [
        'agentsList', 'loginsBody', 'hostsList', 'usersList'
    ];
    
    containers.forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            if (el.tagName === 'TBODY') {
                el.innerHTML = '<tr><td colspan="4" class="no-data">Ошибка загрузки данных</td></tr>';
            } else {
                el.innerHTML = '<div class="no-data">Ошибка загрузки данных</div>';
            }
        }
    });
    
    ['eventsTypeNoData', 'severityNoData', 'processesNoData', 'timelineNoData'].forEach(id => {
        const el = document.getElementById(id);
        if (el) {
            el.classList.remove('hidden');
            el.textContent = 'Ошибка загрузки данных';
        }
    });
    
    ['eventsTypeChart', 'severityChart', 'processesChart', 'timelineChart'].forEach(id => {
        const el = document.getElementById(id);
        if (el) el.classList.add('hidden');
    });
}

function formatTimestamp(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    return date.toLocaleString('ru-RU', {
        day: '2-digit',
        month: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatTimeAgo(date) {
    const now = new Date();
    const diffMs = now - date;
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);
    
    if (diffMins < 1) return 'только что';
    if (diffMins < 60) return `${diffMins} мин. назад`;
    if (diffHours < 24) return `${diffHours} ч. назад`;
    return `${diffDays} дн. назад`;
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

document.addEventListener('DOMContentLoaded', initDashboard);
