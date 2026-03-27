import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import '@fontsource/inter/400.css';
import '@fontsource/inter/500.css';
import '@fontsource/inter/600.css';
import '@fontsource/inter/700.css';
import '@fontsource/inter/900.css';
import './index.css'
import App from './App.jsx'
import PublicShop from './pages/PublicShop.jsx'
import Dashboard from './pages/Dashboard.jsx'
import MyProjects from './pages/MyProjects.jsx'
import Shop from './pages/Shop.jsx'
import ProductDetails from './pages/ProductDetails.jsx'
import Checkout from './pages/Checkout.jsx'
import Login from './pages/Login.jsx'
import Register from './pages/Register.jsx'
import Downloads from './pages/Downloads.jsx'
import HistoryPage from './pages/History.jsx'
import { AuthProvider } from './context/AuthContext.jsx'
import { CartProvider } from './context/CartContext.jsx'
import ProtectedRoute from './components/ProtectedRoute.jsx'
import Wallet from './pages/Wallet.jsx'
import Billing from './pages/Billing.jsx'
import Settings from './pages/Settings.jsx'
import AdminRoute from './components/AdminRoute.jsx'
import AdminDashboard from './pages/admin/Dashboard.jsx'
import AdminLayout from './layouts/AdminLayout.jsx'
import Orders from './pages/admin/Orders.jsx'
import AdminDiscounts from './pages/admin/AdminDiscounts.jsx'
import SubscriptionDetail from './pages/admin/SubscriptionDetail.jsx'
import AdminCatalog from './pages/admin/AdminCatalog.jsx'
import AdminFiles from './pages/admin/AdminFiles.jsx'
import UsersPage from './pages/admin/Users.jsx'
import UserDetailPage from './pages/admin/UserDetail.jsx'
import AdminSupport from './pages/admin/AdminSupport.jsx'
import AdminRoles from './pages/admin/AdminRoles.jsx'
import AdminSystemResources from './pages/admin/AdminSystemResources.jsx'
import Support from './pages/Support.jsx'
import ResetPassword from './pages/ResetPassword.jsx'
import ErrorBoundary from './components/ErrorBoundary.jsx'
import { ToastProvider } from './context/ToastContext.jsx'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <ErrorBoundary>
      <BrowserRouter>
        <AuthProvider>
          <CartProvider>
            <ToastProvider>
              <Routes>
                <Route path="/" element={<App />} />
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                <Route path="/shop" element={<PublicShop />} />
                <Route path="/reset-password" element={<ResetPassword />} />
                <Route
                  path="/dashboard"
                  element={
                    <ProtectedRoute>
                      <Dashboard />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/projetos"
                  element={
                    <ProtectedRoute>
                      <MyProjects />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/loja"
                  element={
                    <ProtectedRoute>
                      <Shop />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/downloads"
                  element={
                    <ProtectedRoute>
                      <Downloads />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/admin"
                  element={
                    <AdminRoute>
                      <AdminLayout />
                    </AdminRoute>
                  }
                >
                  <Route index element={<AdminDashboard />} />
                  <Route path="orders" element={<Orders />} />
                  <Route path="discounts" element={<AdminDiscounts />} />
                  <Route path="catalog" element={<AdminCatalog />} />
                  <Route path="catalog/files" element={<AdminFiles />} />
                  <Route path="subscriptions/:id" element={<SubscriptionDetail />} />
                  <Route path="users" element={<UsersPage />} />
                  <Route path="users/:id" element={<UserDetailPage />} />
                  <Route path="support" element={<AdminSupport />} />
                  <Route path="roles" element={<AdminRoles />} />
                  <Route path="system" element={<AdminSystemResources />} />
                </Route>
                <Route
                  path="/suporte"
                  element={
                    <ProtectedRoute>
                      <Support />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/carteira"
                  element={
                    <ProtectedRoute>
                      <Wallet />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/history"
                  element={
                    <ProtectedRoute>
                      <HistoryPage />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/faturas"
                  element={
                    <ProtectedRoute>
                      <Billing />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/loja/produto/:id"
                  element={
                    <ProtectedRoute>
                      <ProductDetails />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/configuracoes"
                  element={
                    <ProtectedRoute>
                      <Settings />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/checkout"
                  element={
                    <ProtectedRoute>
                      <Checkout />
                    </ProtectedRoute>
                  }
                />
                {/* 404 catch-all */}
                <Route path="*" element={
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', minHeight: '100vh', background: 'var(--bg-primary)', color: 'var(--text-primary)', textAlign: 'center', padding: '2rem' }}>
                    <h1 style={{ fontSize: '6rem', fontWeight: 900, background: 'var(--gradient-primary)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', marginBottom: '1rem' }}>404</h1>
                    <p style={{ fontSize: '1.25rem', color: 'var(--text-secondary)', marginBottom: '2rem' }}>Página não encontrada</p>
                    <a href="/" style={{ padding: '0.75rem 2rem', background: 'var(--gradient-primary)', color: '#fff', borderRadius: 'var(--radius-md)', textDecoration: 'none', fontWeight: 600 }}>Voltar ao Início</a>
                  </div>
                } />
              </Routes>
            </ToastProvider>
          </CartProvider>
        </AuthProvider>
      </BrowserRouter>
    </ErrorBoundary>
  </StrictMode>
)
