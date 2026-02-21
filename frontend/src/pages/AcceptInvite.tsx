import { useState, useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { AxiosError } from 'axios';
import { authService } from '../services/auth';
import type { InviteInfo } from '../types';

export default function AcceptInvite() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');

  const [info, setInfo] = useState<InviteInfo | null>(null);
  const [loadingInfo, setLoadingInfo] = useState(true);
  const [infoError, setInfoError] = useState('');

  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    if (!token) {
      setInfoError('Token de convite não encontrado.');
      setLoadingInfo(false);
      return;
    }

    authService.getInviteInfo(token)
      .then(({ data }) => setInfo(data))
      .catch((err: AxiosError<{ error: string }>) => {
        setInfoError(err.response?.data?.error || 'Convite inválido.');
      })
      .finally(() => setLoadingInfo(false));
  }, [token]);

  const handleAccept = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;
    setLoading(true);
    setError('');
    try {
      const payload: { token: string; name?: string; password?: string } = { token };
      if (!info?.user_exists) {
        payload.name = name;
        payload.password = password;
      }
      await authService.acceptInvite(payload);
      setSuccess(true);
    } catch (err: unknown) {
      const axiosErr = err as AxiosError<{ error: string }>;
      setError(axiosErr.response?.data?.error || 'Erro ao aceitar convite');
    } finally {
      setLoading(false);
    }
  };

  const roleLabels: Record<string, string> = {
    admin: 'Administrador',
    user: 'Usuário',
  };

  if (loadingInfo) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md text-center">
          <p className="text-gray-500">Carregando convite...</p>
        </div>
      </div>
    );
  }

  if (infoError) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md text-center">
          <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
            <svg className="w-8 h-8 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <h2 className="text-xl font-semibold text-gray-900 mb-2">Convite inválido</h2>
          <p className="text-gray-500 mb-4">{infoError}</p>
          <Link to="/login" className="text-blue-600 hover:text-blue-700 font-medium">
            Ir para login
          </Link>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md text-center">
          <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
            <svg className="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="text-xl font-semibold text-gray-900 mb-2">Convite aceito!</h2>
          <p className="text-gray-500 mb-4">
            Você agora faz parte do dashboard <strong>{info?.tenant_name}</strong>.
          </p>
          <Link
            to="/login"
            className="inline-block bg-blue-600 text-white rounded-lg px-6 py-2 font-medium hover:bg-blue-700 transition-colors"
          >
            Fazer login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md">
        <div className="flex items-center justify-center gap-2 mb-2">
          <img src="/assets/logo.svg" alt="DNA Fami" className="h-8 w-8" />
          <h1 className="text-2xl font-bold text-gray-900">DNA Fami</h1>
        </div>
        <p className="text-gray-500 mb-4">Convite para dashboard</p>

        <div className="bg-blue-50 rounded-lg p-4 mb-6">
          <p className="text-sm text-blue-800">
            Você foi convidado para participar do dashboard{' '}
            <strong>{info?.tenant_name}</strong> como{' '}
            <strong>{roleLabels[info?.role || ''] || info?.role}</strong>.
          </p>
        </div>

        <form onSubmit={handleAccept} className="space-y-4">
          {!info?.user_exists && (
            <>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Nome</label>
                <input
                  type="text"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  required
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Seu nome"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Senha</label>
                <input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={6}
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Mínimo 6 caracteres"
                />
              </div>
            </>
          )}

          {info?.user_exists && (
            <p className="text-sm text-gray-600">
              Já identificamos sua conta (<strong>{info.email}</strong>).
              Clique abaixo para aceitar o convite.
            </p>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white rounded-lg py-2 font-medium hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            {loading ? 'Aceitando...' : 'Aceitar convite'}
          </button>

          {error && (
            <p className="text-sm text-red-600 text-center mt-2">{error}</p>
          )}
        </form>
      </div>
    </div>
  );
}
