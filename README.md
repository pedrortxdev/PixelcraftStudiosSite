# Pixelcraft - A Arquitetura do Seu Mundo

Bem-vindo ao repositório frontend do **Pixelcraft**, uma plataforma premium de hospedagem de servidores de Minecraft, construída com a precisão de um arquiteto e a paixão de um criador.

Este projeto é uma Single Page Application (SPA) moderna, desenvolvida para oferecer uma experiência de usuário fluida, responsiva e visualmente impactante.

## 🚀 Tecnologias Utilizadas

O projeto foi construído utilizando as seguintes tecnologias e bibliotecas:

-   **[React 19](https://react.dev/)**: Biblioteca JavaScript para construção de interfaces de usuário.
-   **[Vite](https://vitejs.dev/)**: Build tool rápida e leve para desenvolvimento frontend moderno.
-   **[React Router Dom](https://reactrouter.com/)**: Gerenciamento de rotas e navegação.
-   **[Framer Motion](https://www.framer.com/motion/)**: Biblioteca para animações complexas e gestos.
-   **[Lucide React](https://lucide.dev/)**: Coleção de ícones SVG limpos e consistentes.
-   **Vanilla CSS**: Sistema de design personalizado com variáveis CSS, gradientes e utilitários modernos (sem frameworks CSS pesados).
-   **Font Source (Inter)**: Tipografia moderna e legível.

## 🛠️ Pré-requisitos

Antes de começar, certifique-se de ter o seguinte instalado em sua máquina:

-   **[Node.js](https://nodejs.org/)** (versão 18 ou superior recomendada)
-   **npm** (geralmente vem com o Node.js)

## 📦 Instalação

1.  Clone o repositório (se ainda não o fez):
    ```bash
    git clone https://github.com/seu-usuario/pixelcraft.git
    cd pixelcraft
    ```

2.  Instale as dependências do projeto:
    ```bash
    npm install
    ```

## ▶️ Executando o Projeto

Para iniciar o servidor de desenvolvimento local:

```bash
npm run dev
```

O aplicativo estará disponível em `http://localhost:5173` (ou outra porta indicada no terminal).

## 🏗️ Estrutura do Projeto

A estrutura de pastas do código fonte (`src`) é organizada da seguinte forma:

```
src/
├── assets/         # Imagens, fontes e arquivos estáticos
├── components/     # Componentes Reutilizáveis de UI
├── context/        # Contextos do React (Gerenciamento de Estado Global)
├── pages/          # Páginas da aplicação (Roteamento)
├── services/       # Serviços de integração com API (Backend)
├── App.jsx         # Componente raiz da aplicação
├── main.jsx        # Ponto de entrada da aplicação
└── index.css       # Estilos globais e Design System
```

## 🎨 Design System

O projeto utiliza um sistema de design próprio definido em `index.css`, focado em um tema escuro (Dark Mode) com acentos vibrantes.

-   **Cores Principais**: Tons de azul escuro e preto para o fundo.
-   **Cores de Acento**: Vermelho (`#E01A4F`), Laranja (`#FF6B35`) e Dourado (`#FFD700`).
-   **Tipografia**: Fonte 'Inter' para uma aparência limpa e técnica.

## 🔗 Backend

Este frontend consome uma API Backend desenvolvida em **Go**. Certifique-se de que o backend esteja em execução para que todas as funcionalidades (login, cadastro, dados do servidor) funcionem corretamente.

O backend geralmente roda na porta `8080` ou conforme configurado nas variáveis de ambiente.

## 📜 Scripts Disponíveis

-   `npm run dev`: Inicia o servidor de desenvolvimento.
-   `npm run build`: Compila o projeto para produção.
-   `npm run lint`: Executa o linter para verificar problemas no código.
-   `npm run preview`: Visualiza a versão de produção localmente.

---

Desenvolvido com ❤️ pela equipe Pixelcraft.
