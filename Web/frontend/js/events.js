let allEvents = [];
let filteredEvents = [];
let currentPage = 1;
const pageSize = 50;
let totalPages = 1;
let totalEvents = 0;
let eventTypes = new Set();
let isLoading = false;
let infiniteScrollMode = true; // Always enabled
let scrollObserver = null;
let selectedSeverities = [];
let selectedTypes = [];

async function initEvents() {
    if (!Auth.requireAuth()) {
        return;
    }
    
    setupEventListeners();
    setupMultiSelectFilters();
    
    // Setup infinite scroll immediately
    const paginationContainer = document.getElementById('paginationContainer');
    if (paginationContainer) paginationContainer.style.display = 'none';
    setupInfiniteScroll();

    // Handle URL parameters for search
    const urlParams = new URLSearchParams(window.location.search);
    const searchParam = urlParams.get('search');
    const regexParam = urlParams.get('regex');

    if (searchParam) {
        const searchInput = document.getElementById('searchInput');
        if (searchInput) {
            searchInput.value = searchParam;
        }
    }

    if (regexParam === 'true') {
        const regexToggle = document.getElementById('regexToggle');
        if (regexToggle) {
            regexToggle.checked = true;
        }
    }

    await loadEventsScroll();
}

function setupEventListeners() {
    // Поиск с debounce
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
    
    // Regex toggle
    const regexToggle = document.getElementById('regexToggle');
    if (regexToggle) {
        regexToggle.addEventListener('change', () => {
            currentPage = 1;
            allEvents = []; // Clear for new search
            const tbody = document.getElementById('eventsBody');
            if (tbody) tbody.innerHTML = '';
            loadEventsScroll();
        });
    }

    // Export buttons
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
    
    // Export buttons
    const exportJSON = document.getElementById('exportJSON');
    const exportCSV = document.getElementById('exportCSV');
    
    if (exportJSON) {
        exportJSON.addEventListener('click', () => exportEvents('json'));
    }
    
    if (exportCSV) {
        exportCSV.addEventListener('click', () => exportEvents('csv'));
    }
    
    // Modal
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
    
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            closeEventModal();
        }
    });
}

function setupMultiSelectFilters() {
    // Severity filter
    const severityBtn = document.getElementById('severityBtn');
    const severityDropdown = document.getElementById('severityDropdown');
    const severityCheckboxes = document.querySelectorAll('.severity-checkbox');
    
    if (severityBtn && severityDropdown) {
        severityBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            severityDropdown.classList.toggle('active');
            severityBtn.classList.toggle('active');
            
            // Close type dropdown
            const typeDropdown = document.getElementById('typeDropdown');
            const typeBtn = document.getElementById('typeBtn');
            if (typeDropdown) typeDropdown.classList.remove('active');
            if (typeBtn) typeBtn.classList.remove('active');
        });
        
        severityCheckboxes.forEach(checkbox => {
            checkbox.addEventListener('change', () => {
                updateSelectedFilters();
                currentPage = 1;
                filterEvents();
            });
        });
    }
    
    // Type filter
    const typeBtn = document.getElementById('typeBtn');
    const typeDropdown = document.getElementById('typeDropdown');
    
    if (typeBtn && typeDropdown) {
        typeBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            typeDropdown.classList.toggle('active');
            typeBtn.classList.toggle('active');
            
            // Close severity dropdown
            const severityDropdown = document.getElementById('severityDropdown');
            const severityBtn = document.getElementById('severityBtn');
            if (severityDropdown) severityDropdown.classList.remove('active');
            if (severityBtn) severityBtn.classList.remove('active');
        });
    }
    
    // Close dropdowns when clicking outside
    document.addEventListener('click', (e) => {
        if (!e.target.closest('.multi-select-container')) {
            document.querySelectorAll('.multi-select-dropdown').forEach(dropdown => {
                dropdown.classList.remove('active');
            });
            document.querySelectorAll('.multi-select-btn').forEach(btn => {
                btn.classList.remove('active');
            });
        }
    });
}

function updateSelectedFilters() {
    // Update severity filters
    selectedSeverities = Array.from(document.querySelectorAll('.severity-checkbox:checked'))
        .map(cb => cb.value);
    
    // Update type filters
    selectedTypes = Array.from(document.querySelectorAll('.type-checkbox:checked'))
        .map(cb => cb.value);
    
    // Update button text
    const severityBtn = document.getElementById('severityBtn');
    if (severityBtn) {
        const span = severityBtn.querySelector('span');
        if (selectedSeverities.length > 0) {
            span.textContent = `Severity (${selectedSeverities.length})`;
        } else {
            span.textContent = 'Severity';
        }
    }
    
    const typeBtn = document.getElementById('typeBtn');
    if (typeBtn) {
        const span = typeBtn.querySelector('span');
        if (selectedTypes.length > 0) {
            span.textContent = `Тип (${selectedTypes.length})`;
        } else {
            span.textContent = 'Тип события';
        }
    }
}

function setupInfiniteScroll() {
    const sentinel = document.getElementById('scrollSentinel');
    
    if (!sentinel) return;
    
    // Disconnect existing observer if any
    if (scrollObserver) {
        scrollObserver.disconnect();
    }
    
    scrollObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
            if (entry.isIntersecting && !isLoading && currentPage < totalPages) {
                currentPage++;
                loadEventsScroll();
            }
        });
    }, {
        root: null,
        rootMargin: '200px', // Use reasonable margin to preload
        threshold: 0.1
    });
    
    scrollObserver.observe(sentinel);
}

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
            
            allEvents.forEach(event => {
                if (event.event_type) {
                    eventTypes.add(event.event_type);
                }
            });
            
            populateTypeFilter();
            filterEvents();
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

async function loadEventsScroll() {
    if (isLoading) return;
    isLoading = true;
    
    const scrollLoader = document.getElementById('scrollLoader');
    if (scrollLoader) {
        scrollLoader.style.display = 'flex'; // Show loader only when actively loading
    }
    
    try {
        const response = await API.getEvents(currentPage, pageSize);
        
        if (response.status === 'success' && response.data) {
            allEvents = allEvents.concat(response.data);
            totalPages = response.totalPages || 1;
            totalEvents = response.total || response.count;
            
            response.data.forEach(event => {
                if (event.event_type) {
                    eventTypes.add(event.event_type);
                }
            });
            
            populateTypeFilter();
            filterEvents();
            updateEventsCount();
        }
    } catch (error) {
        console.error('Failed to load events:', error);
        showError();
    } finally {
        isLoading = false;
        if (scrollLoader) {
            scrollLoader.style.display = 'none'; // Hide loader immediately after request finishes
        }
        
        // Hide sentinel if we reached the end
        const sentinel = document.getElementById('scrollSentinel');
        if (sentinel) {
            sentinel.style.display = currentPage >= totalPages ? 'none' : 'block';
        }
    }
}

function populateTypeFilter() {
    const typeDropdown = document.getElementById('typeDropdown');
    if (!typeDropdown) return;
    
    // Keep existing checkboxes state
    const checkedValues = Array.from(document.querySelectorAll('.type-checkbox:checked'))
        .map(cb => cb.value);
    
    typeDropdown.innerHTML = '';
    
    Array.from(eventTypes).sort().forEach(type => {
        const label = document.createElement('label');
        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'type-checkbox';
        checkbox.value = type;
        checkbox.checked = checkedValues.includes(type);
        
        checkbox.addEventListener('change', () => {
            updateSelectedFilters();
            currentPage = 1;
            filterEvents();
        });
        
        label.appendChild(checkbox);
        label.appendChild(document.createTextNode(type));
        typeDropdown.appendChild(label);
    });
    
    updateSelectedFilters();
}

function filterEvents() {
    const searchInput = document.getElementById('searchInput');
    const regexToggle = document.getElementById('regexToggle');
    
    const searchTerm = searchInput?.value || '';
    const useRegex = regexToggle?.checked || false;
    
    let searchRegex = null;
    if (searchTerm && useRegex) {
        try {
            searchRegex = new RegExp(searchTerm, 'i');
        } catch (e) {
            console.error('Invalid regex:', e);
            searchRegex = null;
        }
    }
    
    filteredEvents = allEvents.filter(event => {
        // Search filter
        if (searchTerm) {
            const message = (event.message || '').toLowerCase();
            const rawLog = (event.raw_log || '').toLowerCase();
            const user = (event.user || '').toLowerCase();
            
            if (searchRegex) {
                if (!searchRegex.test(message) && !searchRegex.test(rawLog) && !searchRegex.test(user)) {
                    return false;
                }
            } else {
                const searchLower = searchTerm.toLowerCase();
                if (!message.includes(searchLower) && !rawLog.includes(searchLower) && !user.includes(searchLower)) {
                    return false;
                }
            }
        }
        
        // Severity filter
        if (selectedSeverities.length > 0) {
            if (!selectedSeverities.includes(event.severity)) {
                return false;
            }
        }
        
        // Type filter
        if (selectedTypes.length > 0) {
            if (!selectedTypes.includes(event.event_type)) {
                return false;
            }
        }
        
        return true;
    });
    
    renderEvents();
}

function renderEvents() {
    const tbody = document.getElementById('eventsBody');
    if (!tbody) return;
    
    if (filteredEvents.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="no-data">Нет событий для отображения</td></tr>';
        return;
    }
    
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

function updateEventsCount() {
    const eventsCount = document.getElementById('eventsCount');
    if (eventsCount) {
        eventsCount.textContent = `Всего: ${totalEvents}`;
    }
}

async function exportEvents(format) {
    try {
        const credentials = Auth.getCredentials();
        if (!credentials) {
            window.location.href = 'login.html';
            return;
        }

        const authHeader = 'Basic ' + btoa(credentials.username + ':' + credentials.password);
        
        const response = await fetch(`${API.baseUrl}/events/export?format=${format}`, {
            method: 'GET',
            headers: {
                'Authorization': authHeader
            }
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Server error');
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        
        const contentDisposition = response.headers.get('Content-Disposition');
        let filename = `events_export.${format}`;
        if (contentDisposition) {
            const matches = /filename=([^;]+)/.exec(contentDisposition);
            if (matches && matches[1]) {
                filename = matches[1];
            }
        }
        
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);
    } catch (error) {
        console.error('Export failed:', error);
        alert('Ошибка экспорта данных: ' + error.message);
    }
}

function showEventDetail(index) {
    const event = filteredEvents[index];
    if (!event) return;
    
    const modal = document.getElementById('eventModal');
    const jsonContainer = document.getElementById('eventJson');
    
    if (!modal || !jsonContainer) return;
    
    const jsonHtml = syntaxHighlightJson(JSON.stringify(event, null, 2));
    jsonContainer.innerHTML = jsonHtml;
    
    modal.classList.add('active');
    document.body.style.overflow = 'hidden';
}

function closeEventModal() {
    const modal = document.getElementById('eventModal');
    if (modal) {
        modal.classList.remove('active');
        document.body.style.overflow = '';
    }
}

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

function showError() {
    const tbody = document.getElementById('eventsBody');
    if (tbody) {
        tbody.innerHTML = '<tr><td colspan="6" class="no-data">Ошибка загрузки данных</td></tr>';
    }
}

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

window.showEventDetail = showEventDetail;

document.addEventListener('DOMContentLoaded', initEvents);
