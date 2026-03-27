/**
 * Permission Constants
 * Defines resources, actions, and their Portuguese translations
 */

// Resource types
export const RESOURCES = {
  USERS: 'USERS',
  ROLES: 'ROLES',
  PRODUCTS: 'PRODUCTS',
  ORDERS: 'ORDERS',
  TRANSACTIONS: 'TRANSACTIONS',
  SUPPORT: 'SUPPORT',
  EMAILS: 'EMAILS',
  FILES: 'FILES',
  GAMES: 'GAMES',
  CATEGORIES: 'CATEGORIES',
  PLANS: 'PLANS',
  DASHBOARD: 'DASHBOARD',
  SETTINGS: 'SETTINGS',
  SYSTEM: 'SYSTEM',
  DISCOUNTS: 'DISCOUNTS',
};

// Action types
export const ACTIONS = {
  VIEW: 'VIEW',
  CREATE: 'CREATE',
  EDIT: 'EDIT',
  DELETE: 'DELETE',
  MANAGE: 'MANAGE', // All actions
  VIEW_CPF: 'VIEW_CPF', // Granular action for sensitive data
};

// Resource labels (Portuguese)
export const RESOURCE_LABELS = {
  USERS: 'Usuários',
  ROLES: 'Cargos',
  PRODUCTS: 'Produtos',
  ORDERS: 'Pedidos',
  TRANSACTIONS: 'Transações',
  SUPPORT: 'Suporte',
  EMAILS: 'Emails',
  FILES: 'Arquivos',
  GAMES: 'Jogos',
  CATEGORIES: 'Categorias',
  PLANS: 'Planos',
  DASHBOARD: 'Dashboard',
  SETTINGS: 'Configurações',
  SYSTEM: 'Sistema',
  DISCOUNTS: 'Descontos / Cupons',
};

// Action labels (Portuguese)
export const ACTION_LABELS = {
  VIEW: 'Visualizar',
  CREATE: 'Criar',
  EDIT: 'Editar',
  DELETE: 'Deletar',
  MANAGE: 'Gerenciar',
  VIEW_CPF: 'Ver CPF',
};

// Action descriptions
export const ACTION_DESCRIPTIONS = {
  VIEW: 'Permite visualizar informações',
  CREATE: 'Permite criar novos registros',
  EDIT: 'Permite editar registros existentes',
  DELETE: 'Permite deletar registros',
  MANAGE: 'Permite todas as ações (visualizar, criar, editar, deletar)',
  VIEW_CPF: 'Permite visualizar o CPF descriptografado dos usuários',
};

// Resource categories
export const RESOURCE_CATEGORIES = {
  Administração: ['USERS', 'ROLES', 'SETTINGS'],
  Conteúdo: ['PRODUCTS', 'DISCOUNTS', 'GAMES', 'CATEGORIES', 'PLANS', 'FILES'],
  Financeiro: ['ORDERS', 'TRANSACTIONS'],
  Suporte: ['SUPPORT', 'EMAILS'],
  Sistema: ['DASHBOARD', 'SYSTEM'],
};

// Get category for a resource
export function getResourceCategory(resource) {
  for (const [category, resources] of Object.entries(RESOURCE_CATEGORIES)) {
    if (resources.includes(resource)) {
      return category;
    }
  }
  return 'Outros';
}

// Get all resources grouped by category
export function getResourcesByCategory() {
  const result = {};
  for (const [category, resources] of Object.entries(RESOURCE_CATEGORIES)) {
    result[category] = resources.map(resource => ({
      type: resource,
      label: RESOURCE_LABELS[resource],
    }));
  }
  return result;
}
