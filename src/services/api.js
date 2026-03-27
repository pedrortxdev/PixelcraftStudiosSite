/**
 * API Service - Comunicação com o Backend Pixelcraft
 */

const API_BASE_URL = import.meta.env.VITE_API_URL?.trim() || 'https://api.pixelcraft-studio.store/api/v1';

async function apiRequest(endpoint, options = {}) {
  const token = localStorage.getItem('pixelcraft_token');

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };

  if (token && !options.skipAuth) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const config = { ...options, headers };

  try {
    const url = `${API_BASE_URL}${endpoint}`;
    const response = await fetch(url, config);

    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      const text = await response.text();
      throw new Error(`Non-JSON response: ${text || 'Empty response'}`);
    }

    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || data.message || `HTTP Error ${response.status}`);
    }

    return data;
  } catch (error) {
    console.error('API Request Error:', error);

    if (error.message.includes('401') || error.message.includes('403')) {
      localStorage.removeItem('pixelcraft_token');
      localStorage.removeItem('pixelcraft_user');
      localStorage.removeItem('pixelcraft_token_expiry');

      if (!window.location.pathname.includes('/login')) {
        window.location.href = '/login';
      }
    }

    throw error;
  }
}

/* ---------------- APIs ---------------- */

export const productsAPI = {
  async getAll(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/products?${queryString}` : '/products';
    return apiRequest(endpoint, { method: 'GET', skipAuth: true });
  },
  async getById(id) { return apiRequest(`/products/${id}`, { method: 'GET', skipAuth: true }); },
  async create(productData) {
    // Create product can handle either download_url or file_id
    return apiRequest('/products', { method: 'POST', body: JSON.stringify(productData) });
  },
  async update(id, productData) {
    // Update product can handle either download_url or file_id
    return apiRequest(`/products/${id}`, { method: 'PUT', body: JSON.stringify(productData) });
  },
  async delete(id) { return apiRequest(`/products/${id}`, { method: 'DELETE' }); },
};

export const checkoutAPI = {
  async process(checkoutData) {
    return apiRequest('/checkout', { method: 'POST', body: JSON.stringify(checkoutData) });
  },
};

export const libraryAPI = {
  async getMyLibrary() { return apiRequest('/library', { method: 'GET' }); },
  async getDownloadUrl(productId) {
    // This function will now get the download info, but the actual download
    // happens by navigating to the download URL
    return apiRequest(`/library/${productId}/download`, { method: 'GET' });
  },
  async downloadFile(productId) {
    // Create a temporary download link for file-based products
    const token = localStorage.getItem('pixelcraft_token');
    const headers = { Authorization: `Bearer ${token}` };

    const response = await fetch(`${API_BASE_URL}/library/${productId}/download`, {
      method: 'GET',
      headers,
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || error.message || `HTTP Error ${response.status}`);
    }

    return response;
  },
};

export const gamesAPI = {
  async getAll() { return apiRequest('/games', { method: 'GET', skipAuth: true }); },
  async getWithCategories() { return apiRequest('/games/with-categories', { method: 'GET', skipAuth: true }); },
  async getCategories(gameId) { return apiRequest(`/games/${gameId}/categories`, { method: 'GET', skipAuth: true }); },
};

export const plansAPI = {
  async getAll() { return apiRequest('/plans', { method: 'GET' }); },
};

export const subscriptionsAPI = {
  async getMySubscriptions() { return apiRequest('/subscriptions', { method: 'GET' }); },
  async getPlans() { return apiRequest('/plans', { method: 'GET' }); },
  // Note: create and cancel are handled via checkout flow
  async getChatHistory(id) { return apiRequest(`/subscriptions/${id}/chat`, { method: 'GET' }); },
  async sendMessage(id, content) { return apiRequest(`/subscriptions/${id}/chat`, { method: 'POST', body: JSON.stringify({ content }) }); },
};

export const discountsAPI = {
  async validate(code, amount, cartItems = []) {
    return apiRequest('/discounts/validate', { 
      method: 'POST', 
      body: JSON.stringify({ 
        code, 
        amount, 
        cart_items: cartItems 
      }) 
    });
  },
};

export const historyAPI = {
  async getMyHistory() { return apiRequest('/history', { method: 'GET' }); },
};

export const dashboardAPI = {
  async getStats() { return apiRequest('/dashboard/stats', { method: 'GET' }); },
};

export const authAPI = {
  async login(credentials) { return apiRequest('/auth/login', { method: 'POST', body: JSON.stringify(credentials), skipAuth: true }); },
  async register(userData) { return apiRequest('/auth/register', { method: 'POST', body: JSON.stringify(userData), skipAuth: true }); },
  async forgotPassword(email) { return apiRequest('/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email }), skipAuth: true }); },
};

// paymentsAPI removed - use historyAPI.getMyHistory() instead

export const supportAPI = {
  async getMyTickets(page = 1, limit = 20) {
    return apiRequest(`/support/tickets?page=${page}&limit=${limit}`, { method: 'GET' });
  },
  async getTicket(ticketId) {
    return apiRequest(`/support/tickets/${ticketId}`, { method: 'GET' });
  },
  async createTicket(data) {
    return apiRequest('/support/tickets', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },
  async sendMessage(ticketId, content) {
    return apiRequest(`/support/tickets/${ticketId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ content }),
    });
  },
  async closeTicket(ticketId) {
    return apiRequest(`/support/tickets/${ticketId}/close`, { method: 'POST' });
  },
  getWSUrl(ticketId) {
    const base = API_BASE_URL;
    if (base.startsWith('http')) {
      return base.replace(/^http/, 'ws') + '/ws?ticket_id=' + ticketId;
    }
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}${base}/ws?ticket_id=${ticketId}`;
  },
};

export const billingAPI = {
  async getInvoices() { return apiRequest('/history/invoices', { method: 'GET' }); },
};

export const walletAPI = {
  // Balance is fetched via usersAPI.getMe() - user.balance field
  // Transactions use the /transactions endpoint
  async getTransactions() { return apiRequest('/transactions', { method: 'GET' }); },
  async checkTransactionStatus(tid) { 
    const res = await apiRequest(`/transactions/${tid}`, { method: 'GET' });
    return res.status;
  },
};

export const depositAPI = {
  async create(data) {
    // Isso vai bater no seu backend Go: POST /api/deposit
    return apiRequest('/deposit', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  },
};

export const usersAPI = {
  async getMe() { return apiRequest('/users/me', { method: 'GET' }); },
  async updateMe(updates) {
    return apiRequest('/users/me', {
      method: 'PUT',
      body: JSON.stringify(updates),
    });
  },
  async uploadAvatar(file) {
    const formData = new FormData();
    formData.append('avatar', file);
    // Use raw fetch for FormData (apiRequest adds Content-Type which breaks multipart)
    const token = localStorage.getItem('pixelcraft_token');
    const res = await fetch(`${API_BASE_URL}/users/me/avatar`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData,
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.error || 'Erro ao enviar avatar');
    }
    return res.json();
  },
};



export const adminAPI = {
  async getStats() { return apiRequest('/admin/stats', { method: 'GET' }); },
  async refreshStats() { return apiRequest('/admin/stats/refresh', { method: 'POST' }); },
  async getRecentOrders() { return apiRequest('/admin/orders/recent', { method: 'GET' }); },
  async getTopProducts() { return apiRequest('/admin/products/top', { method: 'GET' }); },
  async getActiveSubscriptions() { return apiRequest('/admin/subscriptions/active', { method: 'GET' }); },
  async getSubscriptionDetails(id) { return apiRequest(`/admin/subscriptions/${id}`, { method: 'GET' }); },
  async updateSubscription(id, data) { return apiRequest(`/admin/subscriptions/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async addSubscriptionLog(id, message) { return apiRequest(`/admin/subscriptions/${id}/logs`, { method: 'POST', body: JSON.stringify({ message }) }); },
  async getSubscriptionChat(id) { return apiRequest(`/admin/subscriptions/${id}/chat`, { method: 'GET' }); },
  async sendSubscriptionMessage(id, content) { return apiRequest(`/admin/subscriptions/${id}/chat`, { method: 'POST', body: JSON.stringify({ content }) }); },
  // Catalog Management
  async createProduct(data) { return apiRequest('/products', { method: 'POST', body: JSON.stringify(data) }); },
  async updateProduct(id, data) { return apiRequest(`/products/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async deleteProduct(id) { return apiRequest(`/products/${id}`, { method: 'DELETE' }); },
  async getPlans() { return apiRequest('/plans', { method: 'GET' }); },
  async createPlan(data) { return apiRequest('/admin/plans', { method: 'POST', body: JSON.stringify(data) }); },
  async updatePlan(id, data) { return apiRequest(`/admin/plans/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async deletePlan(id) { return apiRequest(`/admin/plans/${id}`, { method: 'DELETE' }); },
  // Game Management
  async createGame(data) { return apiRequest('/admin/games', { method: 'POST', body: JSON.stringify(data) }); },
  async updateGame(id, data) { return apiRequest(`/admin/games/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async deleteGame(id) { return apiRequest(`/admin/games/${id}`, { method: 'DELETE' }); },
  // Category Management
  async createCategory(data) { return apiRequest('/admin/categories', { method: 'POST', body: JSON.stringify(data) }); },
  async updateCategory(id, data) { return apiRequest(`/admin/categories/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async deleteCategory(id) { return apiRequest(`/admin/categories/${id}`, { method: 'DELETE' }); },

  // User Management
  async getUsers(page = 1, limit = 20, search = '') {
    return apiRequest(`/admin/users?page=${page}&limit=${limit}&search=${encodeURIComponent(search)}`, { method: 'GET' });
  },
  async getUserDetails(id) { return apiRequest(`/admin/users/${id}`, { method: 'GET' }); },
  async updateUser(id, data) { return apiRequest(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async updateUserPassword(id, password) { return apiRequest(`/admin/users/${id}/password`, { method: 'PUT', body: JSON.stringify({ password }) }); },

  // Role Management
  async addUserRole(userId, role) { return apiRequest(`/admin/users/${userId}/roles`, { method: 'POST', body: JSON.stringify({ role }) }); },
  async removeUserRole(userId, role) { return apiRequest(`/admin/users/${userId}/roles/${role}`, { method: 'DELETE' }); },

  // Support Management
  async getSupportStats() { return apiRequest('/admin/support/stats', { method: 'GET' }); },
  async getSupportTickets(params) { return apiRequest(`/admin/support/tickets?${params}`, { method: 'GET' }); },
  async getSupportTicket(id) { return apiRequest(`/admin/support/tickets/${id}`, { method: 'GET' }); },
  async sendSupportMessage(id, content) { return apiRequest(`/admin/support/tickets/${id}/messages`, { method: 'POST', body: JSON.stringify({ content }) }); },
  async updateSupportStatus(id, status) { return apiRequest(`/admin/support/tickets/${id}/status`, { method: 'PUT', body: JSON.stringify({ status }) }); },
  async assignSupportTicket(id, assigned_to) { return apiRequest(`/admin/support/tickets/${id}/assign`, { method: 'PUT', body: JSON.stringify({ assigned_to }) }); },
  async releaseSupportTicket(id) { return apiRequest(`/admin/support/tickets/${id}/release`, { method: 'PUT' }); },

  // Finance Management
  async listTransactions(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    return apiRequest(`/admin/transactions?${queryString}`, { method: 'GET' });
  },

  // Discount Management
  async getDiscounts() { return apiRequest('/admin/discounts', { method: 'GET' }); },
  async getDiscount(id) { return apiRequest(`/admin/discounts/${id}`, { method: 'GET' }); },
  async createDiscount(data) { return apiRequest('/admin/discounts', { method: 'POST', body: JSON.stringify(data) }); },
  async updateDiscount(id, data) { return apiRequest(`/admin/discounts/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async deleteDiscount(id) { return apiRequest(`/admin/discounts/${id}`, { method: 'DELETE' }); },

  // System Metrics
  async getSystemMetrics() { return apiRequest('/admin/system/metrics', { method: 'GET' }); },
};

export const filesAPI = {
  async upload(file, name) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('name', name);

    const token = localStorage.getItem('pixelcraft_token');
    const headers = { Authorization: `Bearer ${token}` };

    const response = await fetch(`${API_BASE_URL}/files`, {
      method: 'POST',
      headers,
      body: formData,
    });

    const contentType = response.headers.get('content-type');
    if (!contentType || !contentType.includes('application/json')) {
      throw new Error('Non-JSON response');
    }

    const data = await response.json();
    if (!response.ok) {
      throw new Error(data.error || data.message || `HTTP Error ${response.status}`);
    }

    return data;
  },
  async list(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/files?${queryString}` : '/files';
    return apiRequest(endpoint, { method: 'GET' });
  },
  async listForSelection(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/files/selection?${queryString}` : '/files/selection';
    return apiRequest(endpoint, { method: 'GET' });
  },
  async listAllAdmin(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/admin/files?${queryString}` : '/admin/files';
    return apiRequest(endpoint, { method: 'GET' });
  },
  async getById(id) { return apiRequest(`/files/${id}`, { method: 'GET' }); },
  async update(id, data) { return apiRequest(`/files/${id}`, { method: 'PUT', body: JSON.stringify(data) }); },
  async delete(id) { return apiRequest(`/files/${id}`, { method: 'DELETE' }); },
  async download(productId) { return apiRequest(`/files/${productId}/download`, { method: 'GET' }); },
  
  // File Permissions
  async getPermissions(id) { return apiRequest(`/files/${id}/permissions`, { method: 'GET' }); },
  async updatePermissions(id, data) { 
    return apiRequest(`/files/${id}/permissions`, { 
      method: 'PUT', 
      body: JSON.stringify(data) 
    }); 
  },
  async addRolePermission(id, role) {
    return apiRequest(`/files/${id}/permissions/roles`, {
      method: 'POST',
      body: JSON.stringify({ role })
    });
  },
  async removeRolePermission(id, role) {
    return apiRequest(`/files/${id}/permissions/roles/${role}`, { method: 'DELETE' });
  },
  async addProductPermission(id, productId) {
    return apiRequest(`/files/${id}/permissions/products`, {
      method: 'POST',
      body: JSON.stringify({ product_id: productId })
    });
  },
  async removeProductPermission(id, productId) {
    return apiRequest(`/files/${id}/permissions/products/${productId}`, { method: 'DELETE' });
  },
  async regeneratePublicLink(id) {
    return apiRequest(`/files/${id}/regenerate-public-link`, { method: 'POST' });
  },
  async generateOneTimeLink(id, options = {}) {
    return apiRequest(`/files/${id}/generate-one-time-link`, {
      method: 'POST',
      body: JSON.stringify(options)
    });
  },
  async getAccessLogs(id, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const endpoint = queryString ? `/files/${id}/access-logs?${queryString}` : `/files/${id}/access-logs`;
    return apiRequest(endpoint, { method: 'GET' });
  },
};

export const aiAPI = {
  async formatText(text) {
    return apiRequest('/ai/format', { method: 'POST', body: JSON.stringify({ text }) });
  },
  async generateAvatar(prompt, userId = null) {
    return apiRequest('/ai/generate-avatar', {
      method: 'POST',
      body: JSON.stringify({ prompt, user_id: userId })
    });
  },
};

export const rolesAPI = {
  // Get available roles
  async getAvailableRoles() {
    return apiRequest('/admin/permissions/available-roles', { method: 'GET' });
  },

  // Get permissions for a specific role
  async getRolePermissions(role) {
    return apiRequest(`/admin/permissions/roles/${role}`, { method: 'GET' });
  },

  // Add permission to a role
  async addRolePermission(role, resource, action) {
    return apiRequest(`/admin/permissions/roles/${role}`, {
      method: 'POST',
      body: JSON.stringify({ resource, action })
    });
  },

  // Remove permission from a role
  async removeRolePermission(role, resource, action) {
    return apiRequest(`/admin/permissions/roles/${role}`, {
      method: 'DELETE',
      body: JSON.stringify({ resource, action })
    });
  },

  // Get available resources
  async getAvailableResources() {
    return apiRequest('/admin/permissions/resources', { method: 'GET' });
  },

  // Get available actions
  async getAvailableActions() {
    return apiRequest('/admin/permissions/actions', { method: 'GET' });
  },

  // Get current user's permissions
  async getMyPermissions() {
    return apiRequest('/permissions/me', { method: 'GET' });
  },

  // Get all role permissions (for admin overview)
  async getAllRolePermissions() {
    return apiRequest('/admin/permissions/roles', { method: 'GET' });
  },

  // Advanced Permission Features
  async getAuditLog(page = 1, limit = 50, role = '') {
    const params = new URLSearchParams({ page, limit });
    if (role) params.append('role', role);
    return apiRequest(`/admin/permissions/audit-log?${params}`, { method: 'GET' });
  },

  async inheritPermissions(targetRole, sourceRole) {
    return apiRequest(`/admin/permissions/roles/${targetRole}/inherit`, {
      method: 'POST',
      body: JSON.stringify({ source_role: sourceRole })
    });
  },

  async removeInheritedPermissions(role) {
    return apiRequest(`/admin/permissions/roles/${role}/inherited`, { method: 'DELETE' });
  },

  async getCustomRoles() {
    return apiRequest('/admin/permissions/custom-roles', { method: 'GET' });
  },

  async createCustomRole(data) {
    return apiRequest('/admin/permissions/custom-roles', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  },

  async deleteCustomRole(id) {
    return apiRequest(`/admin/permissions/custom-roles/${id}`, { method: 'DELETE' });
  },

  async exportPermissions(roles = []) {
    const params = roles.length ? `?${roles.map(r => `roles=${r}`).join('&')}` : '';
    return apiRequest(`/admin/permissions/export${params}`, { method: 'GET' });
  },

  async importPermissions(templateData, overwrite = false) {
    return apiRequest('/admin/permissions/import', {
      method: 'POST',
      body: JSON.stringify({ template_data: templateData, overwrite })
    });
  },

  async getTemplates() {
    return apiRequest('/admin/permissions/templates', { method: 'GET' });
  },

  async saveTemplate(data) {
    return apiRequest('/admin/permissions/templates', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  },

  async getDashboard() {
    return apiRequest('/admin/permissions/dashboard', { method: 'GET' });
  },

  async getNotifications() {
    return apiRequest('/permissions/notifications', { method: 'GET' });
  },

  async markNotificationAsRead(id) {
    return apiRequest(`/permissions/notifications/${id}/read`, { method: 'PUT' });
  }
};

/* DEFAULT EXPORT */
export default {
  products: productsAPI,
  checkout: checkoutAPI,
  library: libraryAPI,
  subscriptions: subscriptionsAPI,
  plans: plansAPI,
  games: gamesAPI,
  discounts: discountsAPI,
  dashboard: dashboardAPI,
  auth: authAPI,
  history: historyAPI,
  users: usersAPI,
  billing: billingAPI,
  wallet: walletAPI,
  deposit: depositAPI,
  admin: adminAPI,
  files: filesAPI,
  ai: aiAPI,
  roles: rolesAPI,
};
