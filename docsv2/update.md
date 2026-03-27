# Tarefas Rigorosas de Frontend (docsv2/update.md)

Instruções granulares e exaustivas para a modificação da base de código frontend focando no ecossistema React, roteamento, chamadas de API e manipulação de estado. Todos os desenvolvedores de UI devem seguir este escopo para garantir uma arquitetura livre de débitos.

---

## 1. Módulo: Autenticação Segura (Esqueci a Senha)

### A. Refatoração do Modal de Solicitação (`src/pages/Login.jsx`)
- **Remover lógicas antigas:** Atualmente o `handleForgotPassword` assume sucesso prematuramente chamando a API antiga que reseta automaticamente a senha.
- **Estado Visual:** Submeter apenas o `email`. O servidor confirmará a ação. Indicar na UI: *“Enviamos instruções detalhadas e um Código de Redefinição para o e-mail associado.”*
- **Bloquear Retentativas:** Prevenir spam inserindo debounce de proteção de 60 segundos após envio de solicitação de e-mail com sucesso no state local `forgotLoading / cooldown`.

### B. Criação da Rota de Resgate (`src/pages/ResetPassword.jsx`)
- **Implementação Estrutural:** Criar o componente `ResetPassword.jsx`.
- **Roteamento:** Adicionar rota `/reset-password/:token` em `main.jsx`.
- **Interface e Elementos:**
  - Extração do Token: O componente deve ler `const { token } = useParams()` do React Router.
  - O Token não deve ser editável. Deve estar invisível no DOM ou como `input hidden`.
  - Formulário visível exigirá três inputs:
    1. **Código de Verificação** (`input type="text"`, uppercase force array, max-length 8). Indicar instrução "Código recebido no e-mail".
    2. **Nova Senha** (`input type="password"`).
    3. **Confirmar Nova Senha** (`input type="password"`).
- **Validação Local de UX:** Checar força de senha e equiparação entre a nova senha e a confirmação antes de disparar o FETCH (diminuir carga do backend).
- **Integração na API:** Criar em `src/services/api.js` dentro do bloco estático de Auth:
  ```javascript
  resetPassword: async (token, code, newPassword) => {
    return apiRequest('/auth/reset-password', {
      method: 'POST', body: JSON.stringify({ token, code, new_password: newPassword })
    });
  }
  ```

---

## 2. Módulo: UI de Permissões Granulares (CPF e Dados Sensíveis)

### A. Ajuste de Hook Global (`src/hooks/usePermissions.js`)
- Caso necessário, certifique-se de que o contexto em `usePermissions` seja flexível o suficiente para checar combinações de array `['view', 'view_cpf', 'edit']`. Ele deverá ser utilizado da seguinte forma nos componentes de Dashboard e Área Administrativa:
  `const canViewCpf = permissions.some(p => p.resource === 'users' && p.action === 'view_cpf');`

### B. Implementação no Profile / Settings (`src/pages/Settings.jsx` e painel de Client)
- A área do Cliente precisa mostrar o CPF travado/mascarado. O cliente **sempre** tem direito a ver o próprio CPF (ele não confia no RBAC geral da plataforma, ele confia na AuthContext `user.id === requested_id`).
- Lógica Visual (Settings do Client): Exibir formatado. `***.***.123-**` dependendo da implementação, mas teremos como requisito disponibilizar o campo `cpf` no form profile.

### C. Implementação Administrativa (`src/pages/admin/Users.jsx` & `UserDetail.jsx`)
- Validar via Hook se o token admin em execução suporta CPF view.
- No `Users.jsx` (Listagem DataGrid): Adicionar coluna opcional.
  ```jsx
  {canViewCpf && <div style={{width: 150}}>CPF</div>}
  // no Row map:
  {canViewCpf && <div>{user.cpf ? formatCPF(user.cpf) : 'Não consta'}</div>}
  ```
- No `UserDetail.jsx`: Somente popular o field `CPF Input Value` do formulário se o Admin possuir hierarquia condizente. Ocultar inteiramente o input se a flag faltar.

---

## 3. Módulo: Arquivos Mídias e Extensões (`Downloads.jsx` & `AdminFiles.jsx`)

### A. Substituição da Lógica de Headers
- O problema da Regex engessada deve ser erradicado. Substituir a linha de Content-Disposition match por este algoritmo testado mundialmente (a ser adicionado numa utils `src/utils/fileExtract.js`):
  ```javascript
  export const extractFilename = (contentDisposition, defaultName = "download.zip") => {
    if (!contentDisposition) return defaultName;
    
    // Tenta formato RFC 5987 (filename*=UTF-8'')
    let match = contentDisposition.match(/filename\*=utf-8''([^;]*)/i);
    if (match && match[1]) return decodeURIComponent(match[1]);
    
    // Tenta formato Quoted (filename="name.ext")
    match = contentDisposition.match(/filename="([^";]+)"/i);
    if (match && match[1]) return match[1];

    // Tenta formato Loose sem aspas (filename=name.ext)
    match = contentDisposition.match(/filename=([^;\r\n"']+)/i);
    if (match && match[1]) return match[1].trim();

    return defaultName;
  };
  ```
- Integrar este utilitário à função `handleDownload` de ambas as páginas (Usuário final e Admin), removendo duplicação de lógicas. Fazer importação.

---

## 4. Módulo: Financeiro (Pix Polling & API Cleanups)

### A. Retirada do Polling Amador no Wallet (`src/pages/Wallet.jsx`)
- O polling via `setInterval` comparando saldo final vs. saldo atual `(userResp.balance > initialBalance) ` deve ser ELIMINADO total.
- **A Nova Metodologia:** 
  O retorno estipulado da `depositAPI.create()` passará a exigir o identificador único da transação gerada (`res.transaction_id`).
- Alterar o `useRef` para injetar o pool em status exato por API especializada (Criar em `services/api.js > walletAPI.checkTransactionStatus(tid)`).
  ```javascript
  const status = await walletAPI.checkTransactionStatus(activeTxId);
  if (status === 'COMPLETED') { setDepositStep('success'); ... }
  ```
- Assim, separa-se UX limpa de inferências duvidosas baseadas em balances paralelos.

### B. Remoção de Hardcodes na Integração
- **AdminFiles.jsx (Download):** A string chumbada de host deverá rodar sob o Wrapper do `api.js` e do base URL env param (`import.meta.env.VITE_API_URL`). Usar API instances que já herdam token no Header, ou então extrair `VITE_API_URL` formal.
- **Utilitários Globais (Avatares):** Exportar uma função compartilhada `getAvatarUrl(path)` global dentro das utils do Vite, ao invés da poluição presente no header de múltiplas telas admin. Mover instâncias espalhadas para apontar pro src/utils genérico.
