/**
 * SIEM Events Page Module
 * Handles events listing, filtering, and detail view
 */

// State
let allEvents = [];
let filteredEvents = [];
let currentPage = 1;
const pageSize = 50; // Match server-side pagination
let totalPages = 1;
let totalEvents = 0;
let eventTypes = new Set();
let isLoading = false;

/**
 * Initialize events page
 */
async function initEvents() {
    // Check authentication
    if (!Auth.requireAuth()) {
        return;
    }
    
    // Setup event listeners
    setupEventListeners();
    
    // Initial load
    await loadEvents();
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Search input
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        let debounceTimer;
        searchInput.addEventListener('input', () => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => {
                currentPage = 1;
                filterEvents();
            }, 300);
        });
    }
    
    // Severity filter
    const severityFilter = document.getElementById('severityFilter');
    if (severityFilter) {
        severityFilter.addEventListener('change', () => {
            currentPage = 1;
            filterEvents();
        });
    }
    
    // Type filter
    const typeFilter = document.getElementById('typeFilter');
    if (typeFilter) {
        typeFilter.addEventListener('change', () => {
            currentPage = 1;
            filterEvents();
        });
    }
    
    // Pagination buttons
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');
    
    if (prevBtn) {
        prevBtn.addEventListener('click', () => {
            if (currentPage > 1 && !isLoading) {
                currentPage--;
                loadEvents();
            }
        });
    }
    
    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            if (currentPage < totalPages && !isLoading) {
                currentPage++;
                loadEvents();
            }
        });
    }
    
    // Modal close
    const closeModal = document.getElementById('closeModal');
    const modalOverlay = document.getElementById('eventModal');
    
    if (closeModal) {
        closeModal.addEventListener('click', closeEventModal);
    }
    
    if (modalOverlay) {
        modalOverlay.addEventListener('click', (e) => {
            if (e.target === modalOverlay) {
                closeEventModal();
            }
        });
    }
    
    // Close on Escape
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeEventModal();
        }
    });
}

/**
 * Load events from API with server-side pagination
 */
async function loadEvents() {
    if (isLoading) return;
    isLoading = true;
    
    const tbody = document.getElementById('eventsBody');
    if (tbody) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading">Загрузка...</td></tr>';
    }
    
    try {
        const response = await API.getEvents(currentPage, pageSize);
        
        if (response.status === 'success' && response.data) {
            allEvents = response.data;
            filteredEvents = allEvents;
            totalPages = response.totalPages || 1;
            totalEvents = response.total || response.count;
            
            // Extract event types for filter (from current page)
            allEvents.forEach(event => {
                if (event.event_type) {
                    eventTypes.add(event.event_type);
                }
            });
            
            populateTypeFilter();
            renderEvents();
            updateEventsCount();
            updatePagination();
        } else {
            showNoData();
        }
    } catch (error) {
        console.error('Failed to load events:', error);
        showError();
    } finally {
        isLoading = false;
    }
}

/**
 * Populate event type filter dropdown
 */
function populateTypeFilter() {
    const typeFilter = document.getElementById('typeFilter');
    if (!typeFilter) return;
    
    // Keep first option (All types)
    const firstOption = typeFilter.options[0];
    typeFilter.innerHTML = '';
    typeFilter.appendChild(firstOption);
    
    // Add event types
    Array.from(eventTypes).sort().forEach(type => {
        const option = document.createElement('option');
        option.value = type;
        option.textContent = type;
        typeFilter.appendChild(option);
    });
}

/**
 * Filter events based on search and filters (client-side filtering on current page)
 */
function filterEvents() {
    const searchInput = document.getElementById('searchInput');
    const severityFilter = document.getElementById('severityFilter');
    const typeFilter = document.getElementById('typeFilter');
    
    const searchTerm = (searchInput?.value || '').toLowerCase();
    const severityValue = severityFilter?.value || '';
    const typeValue = typeFilter?.value || '';
    
    // If any filters are active, filter client-side on current page data
    if (searchTerm || severityValue || typeValue) {
        filteredEvents = allEvents.filter(event => {
            // Search filter
            if (searchTerm) {
                const message = (event.message || '').toLowerCase();
                const rawLog = (event.raw_log || '').toLowerCase();
                if (!message.includes(searchTerm) && !rawLog.includes(searchTerm)) {
                    return false;
                }
            }
            
            // Severity filter
            if (severityValue && event.severity !== severityValue) {
                return false;
            }
            
            // Type filter
            if (typeValue && event.event_type !== typeValue) {
                return false;
            }
            
            return true;
        });
    } else {
        filteredEvents = allEvents;
    }
    
    renderEvents();
}

/**
 * Render events table
 */
function renderEvents() {
    const tbody = document.getElementById('eventsBody');
    if (!tbody) return;
    
    if (filteredEvents.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="no-data">Нет событий для отображения</td></tr>';
        return;
    }
    
    // Events are already paginated from server, render all
    const html = filteredEvents.map((event, index) => {
        const timestamp = formatTimestamp(event.timestamp);
        const agentId = event.agent_id || '-';
        const type = event.event_type || '-';
        const severity = event.severity || 'low';
        const user = event.user || '-';
        const message = event.message || event.raw_log || '-';
        
        return `
            <tr data-index="${index}" onclick="showEventDetail(${index})">
                <td class="mono">${timestamp}</td>
                <td class="mono">${escapeHtml(agentId)}</td>
                <td>${escapeHtml(type)}</td>
                <td><span class="severity-badge ${severity}">${severity.toUpperCase()}</span></td>
                <td class="mono">${escapeHtml(user)}</td>
                <td class="message-cell">${escapeHtml(message)}</td>
            </tr>
        `;
    }).join('');
    
    tbody.innerHTML = html;
}

/**
 * Update pagination controls (using server-side pagination info)
 */
function updatePagination() {
    const paginationInfo = document.getElementById('paginationInfo');
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');
    
    const startIndex = (currentPage - 1) * pageSize + 1;
    const endIndex = Math.min(currentPage * pageSize, totalEvents);
    
    if (paginationInfo) {
        if (totalEvents === 0) {
            paginationInfo.textContent = 'Показано 0-0 из 0';
        } else {
            paginationInfo.textContent = `Показано ${startIndex}-${endIndex} из ${totalEvents}`;
        }
    }
    
    if (prevBtn) {
        prevBtn.disabled = currentPage <= 1 || isLoading;
    }
    
    if (nextBtn) {
        nextBtn.disabled = currentPage >= totalPages || isLoading;
    }
}

/**
 * Update total events count
 */
function updateEventsCount() {
    const eventsCount = document.getElementById('eventsCount');
    if (eventsCount) {
        eventsCount.textContent = `Всего: ${totalEvents}`;
    }
}

/**
 * Show event detail modal
 */
function showEventDetail(index) {
    const event = filteredEvents[index];
    if (!event) return;
    
    const modal = document.getElementById('eventModal');
    const jsonContainer = document.getElementById('eventJson');
    
    if (!modal || !jsonContainer) return;
    
    // Syntax highlight JSON
    const jsonHtml = syntaxHighlightJson(JSON.stringify(event, null, 2));
    jsonContainer.innerHTML = jsonHtml;
    
    // Show modal
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
}

/**
 * Close event detail modal
 */
function closeEventModal() {
    const modal = document.getElementById('eventModal');
    if (modal) {
        modal.classList.remove('active');
        document.body.style.overflow = '';
    }
}

/**
 * Syntax highlight JSON
 */
function syntaxHighlightJson(json) {
    return json
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function(match) {
            let cls = 'json-number';
            if (/^"/.test(match)) {
                if (/:$/.test(match)) {
                    cls = 'json-key';
                } else {
                    cls = 'json-string';
                }
            } else if (/true|false/.test(match)) {
                cls = 'json-boolean';
            } else if (/null/.test(match)) {
                cls = 'json-null';
            }
            return '<span class="' + cls + '">' + match + '</span>';
        });
}

/**
 * Show no data message
 */
function showNoData() {
    const tbody = document.getElementById('eventsBody');
    if (tbody) {
        tbody.innerHTML = '<tr><td colspan="6" class="no-data">Нет событий для отображения</td></tr>';
    }
    
    const eventsCount = document.getElementById('eventsCount');
    if (eventsCount) {
        eventsCount.textContent = 'Всего: 0';
    }
}

/**
 * Show error message
 */
function showError() {
    const tbody = document.getElementById('eventsBody');
    if (tbody) {
        tbody.innerHTML = '<tr><td colspan="6" class="no-data">Ошибка загрузки данных</td></tr>';
    }
}

// Utility functions

function formatTimestamp(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    return date.toLocaleString('ru-RU', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Make showEventDetail available globally
window.showEventDetail = showEventDetail;

// Initialize on DOM ready
document.addEventListener('DOMContentLoaded', initEvents);
