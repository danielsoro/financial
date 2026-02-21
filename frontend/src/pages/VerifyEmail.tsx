import { useState, useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { authService } from '../services/auth';

export default function VerifyEmail() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>(token ? 'loading' : 'error');
  const [message, setMessage] = useState(token ? '' : 'Token de verificação não encontrado.');

  useEffect(() => {
    if (!token) return;

    authService.verifyEmail(token)
      .then(() => {
        setStatus('success');
        setMessage('Email verificado com sucesso!');
      })
      .catch(() => {
        setStatus('error');
        setMessage('Token inválido ou expirado.');
      });
  }, [token]);

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-xl shadow-lg p-8 w-full max-w-md text-center">
        <div className="flex items-center justify-center gap-2 mb-6">
          <img src="/assets/logo.svg" alt="DNA Fami" className="h-8 w-8" />
          <h1 className="text-2xl font-bold text-gray-900">DNA Fami</h1>
        </div>

        {status === 'loading' && (
          <p className="text-gray-500">Verificando email...</p>
        )}

        {status === 'success' && (
          <>
            <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
              <svg className="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">{message}</h2>
            <p className="text-gray-500 mb-4">Sua conta está ativa. Faça login para começar.</p>
            <Link
              to="/login"
              className="inline-block bg-blue-600 text-white rounded-lg px-6 py-2 font-medium hover:bg-blue-700 transition-colors"
            >
              Fazer login
            </Link>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="mx-auto w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mb-4">
              <svg className="w-8 h-8 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">Erro na verificação</h2>
            <p className="text-gray-500 mb-4">{message}</p>
            <Link
              to="/login"
              className="inline-block text-blue-600 hover:text-blue-700 font-medium"
            >
              Ir para login
            </Link>
          </>
        )}
      </div>
    </div>
  );
}
