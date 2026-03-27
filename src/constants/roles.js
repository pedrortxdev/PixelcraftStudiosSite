/**
 * Role Constants and Configuration
 * Defines role hierarchy, display information, and utility functions
 */

// Role hierarchy levels (higher = more power)
export const ROLE_HIERARCHY = {
  PARTNER: 1,
  CLIENT: 2,
  CLIENT_VIP: 3,
  SUPPORT: 4,
  ADMIN: 5,
  DEVELOPMENT: 6,
  ENGINEERING: 7,
  DIRECTION: 8,
};

// Role display configuration (labels and colors)
export const ROLE_DISPLAY_CONFIG = {
  PARTNER: {
    label: 'Parceiro',
    color: '#00bd65',
    description: 'Parceiro: +1% de lucros em vendas',
  },
  CLIENT: {
    label: 'Cliente',
    color: '#00d415',
    description: 'Cliente: prioridade 2 estrelas',
  },
  CLIENT_VIP: {
    label: 'Cliente VIP',
    color: '#00bd13',
    description: 'Cliente VIP: prioridade 3 estrelas',
  },
  SUPPORT: {
    label: 'Suporte',
    color: '#fbff00',
    description: 'Suporte: acesso restrito (Atendimento + Email próprio)',
  },
  ADMIN: {
    label: 'Administração',
    color: '#bd005b',
    description: 'Administração: visualização total, sem edição',
  },
  DEVELOPMENT: {
    label: 'Desenvolvimento',
    color: '#0047bd',
    description: 'Desenvolvimento: edita planos/produtos',
  },
  ENGINEERING: {
    label: 'Engenharia',
    color: '#6a00ff',
    description: 'Engenharia: emails, catálogo, pedidos, editar senhas/saldo',
  },
  DIRECTION: {
    label: 'Direção',
    color: '#ff3f00',
    description: 'Direção: acesso total',
  },
};

// Admin roles (roles that have admin panel access)
export const ADMIN_ROLES = [
  'SUPPORT',
  'ADMIN',
  'DEVELOPMENT',
  'ENGINEERING',
  'DIRECTION',
];

// Client roles (roles for regular users)
export const CLIENT_ROLES = [
  'PARTNER',
  'CLIENT',
  'CLIENT_VIP',
];

// All available roles
export const ALL_ROLES = [
  'PARTNER',
  'CLIENT',
  'CLIENT_VIP',
  'SUPPORT',
  'ADMIN',
  'DEVELOPMENT',
  'ENGINEERING',
  'DIRECTION',
];
