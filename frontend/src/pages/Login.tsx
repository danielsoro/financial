import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { AxiosError } from 'axios';
import { useAuth } from '../contexts/AuthContext';
import { authService } from '../services/auth';
import type { TenantInfo } from '../types';

export default function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Two-phase state
  const [selectorToken, setSelectorToken] = useState<string | null>(null);
  const [tenants, setTenants] = useState<TenantInfo[]>([]);

  const { login } = useAuth();
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const { data } = await authService.login(email, password);

      // Single tenant: auto-selected
      if (data.token && data.user) {
        login(data.token, data.user);
        navigate('/');
        return;
      }

      // Multi-tenant: show selector
      if (data.selector_token && data.tenants) {
        setSelectorToken(data.selector_token);
        setTenants(data.tenants);
      }
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{ error: string }>;
      setError(axiosErr.response?.data?.error || 'Erro ao fazer login');
    } finally {
      setLoading(false);
    }
  };

  const handleSelectTenant = async (tenantId: string) => {
    if (!selectorToken) return;
    setLoading(true);
    setError('');
    try {
      const { data } = await authService.selectTenant(selectorToken, tenantId);
      login(data.token, data.user);
      navigate('/');
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{ error: string }>;
      setError(axiosErr.response?.data?.error || 'Erro ao selecionar dashboard');
      // If selector expired, reset to login
      if (axiosErr.response?.status === 401) {
        setSelectorToken(null);
        setTenants([]);
      }
    } finally {
      setLoading(false);
    }
  };

  const roleLabels: Record<string, string> = {
    owner: 'Proprietário',
    admin: 'Admin',
    user: 'Usuário',
  };

  // Tenant selector view
  if (selectorToken && tenants.length > 0) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
          <div className="flex items-center justify-center gap-2 mb-2">
            <img src="/assets/logo.svg" alt="DNA Fami" className="h-8 w-8" />
            <h1 className="text-2xl font-bold text-gray-900">DNA Fami</h1>
          </div>
          <p className="text-gray-500 mb-6">Selecione um dashboard</p>

          <div className="space-y-3">
            {tenants.map((t) => (
              <button
                key={t.tenant_id}
                onClick={() => handleSelectTenant(t.tenant_id)}
                disabled={loading}
                className="w-full flex items-center justify-between p-4 rounded-lg border border-gray-200 hover:border-blue-500 hover:bg-blue-50 transition-colors disabled:opacity-50 text-left"
              >
                <div>
                  <p className="font-medium text-gray-900">{t.tenant_name}</p>
                  <p className="text-sm text-gray-500">{roleLabels[t.role] || t.role}</p>
                </div>
                <svg className="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </button>
            ))}
          </div>

          <button
            onClick={() => { setSelectorToken(null); setTenants([]); setError(''); }}
            className="mt-4 w-full text-sm text-gray-500 hover:text-gray-700"
          >
            Voltar ao login
          </button>

          {error && (
            <p className="text-sm text-red-600 text-center mt-2">{error}</p>
          )}
        </div>
      </div>
    );
  }

  // Login form
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
        <div className="flex items-center justify-center gap-2 mb-2">
          <img src="/assets/logo.svg" alt="DNA Fami" className="h-8 w-8" />
          <h1 className="text-2xl font-bold text-gray-900">DNA Fami</h1>
        </div>
        <p className="text-gray-500 mb-6">Faça login para continuar</p>
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Senha</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white rounded-lg py-2 font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {loading ? 'Entrando...' : 'Entrar'}
          </button>
          {error && (
            <p className="text-sm text-red-600 text-center mt-2">{error}</p>
          )}
        </form>
        <p className="text-sm text-gray-500 text-center mt-6">
          Não tem conta?{' '}
          <Link to="/register" className="text-blue-600 hover:text-blue-700 font-medium">
            Criar conta
          </Link>
        </p>
      </div>
    </div>
  );
}
