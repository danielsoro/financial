import { useState, useEffect } from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import {
  HiHome,
  HiArrowTrendingUp,
  HiArrowTrendingDown,
  HiTag,
  HiShieldCheck,
  HiArrowRightOnRectangle,
  HiPencilSquare,
  HiUserGroup,
  HiBars3,
  HiXMark,
  HiChevronLeft,
  HiChevronRight,
} from 'react-icons/hi2';

const navItems = [
  { to: '/', label: 'Dashboard', icon: HiHome },
  { to: '/income', label: 'Receitas', icon: HiArrowTrendingUp },
  { to: '/expenses', label: 'Despesas', icon: HiArrowTrendingDown },
  { to: '/categories', label: 'Categorias', icon: HiTag },
  { to: '/expense-limits', label: 'Tetos de Gastos', icon: HiShieldCheck },
];

export default function AppLayout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const isAdmin = user?.role === 'admin';

  const [collapsed, setCollapsed] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const closeMobile = () => setMobileOpen(false);

  // Reset mobileOpen when resizing to desktop
  useEffect(() => {
    const mql = window.matchMedia('(min-width: 768px)');
    const handler = () => {
      if (mql.matches) setMobileOpen(false);
    };
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, []);

  return (
    <div className="flex min-h-screen w-full overflow-x-hidden">
      {/* Mobile backdrop */}
      {mobileOpen && (
        <div
          className="fixed inset-0 bg-black/50 z-40 md:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}

      {/* Mobile top bar */}
      <div className="fixed top-0 left-0 right-0 h-14 bg-white border-b border-gray-200 flex items-center px-4 z-30 md:hidden">
        <button
          onClick={() => setMobileOpen(true)}
          className="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded-lg transition-colors"
          aria-label="Abrir menu"
        >
          <HiBars3 className="h-6 w-6" />
        </button>
        <h1 className="ml-3 text-lg font-bold text-gray-900">Finance</h1>
      </div>

      {/* Sidebar */}
      <aside
        className={`fixed left-0 top-0 h-screen bg-gray-900 text-white flex flex-col z-50 transition-all duration-300 ease-in-out w-64 ${
          mobileOpen ? 'translate-x-0' : '-translate-x-full'
        } md:translate-x-0 ${collapsed ? 'md:w-16' : 'md:w-64'}`}
      >
        {/* Close button (mobile) */}
        <button
          onClick={() => setMobileOpen(false)}
          className="absolute top-4 right-4 p-1 text-gray-400 hover:text-white md:hidden"
          aria-label="Fechar menu"
        >
          <HiXMark className="h-6 w-6" />
        </button>

        {/* Logo */}
        <div className={`py-6 ${collapsed ? 'px-0 flex justify-center' : 'px-6'}`}>
          <h1 className={`font-bold tracking-tight transition-all duration-300 ${collapsed ? 'text-lg' : 'text-2xl'}`}>
            {collapsed ? 'F' : 'Finance'}
          </h1>
        </div>

        {/* Navigation */}
        <nav className={`flex-1 ${collapsed ? 'px-2' : 'px-3'} space-y-1`}>
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              onClick={closeMobile}
              className={({ isActive }) =>
                `flex items-center ${collapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-gray-800 text-white'
                    : 'text-gray-400 hover:text-white hover:bg-gray-800'
                }`
              }
              title={collapsed ? item.label : undefined}
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {!collapsed && <span className="whitespace-nowrap overflow-hidden">{item.label}</span>}
            </NavLink>
          ))}

          {/* Admin Section */}
          {isAdmin && (
            <>
              {collapsed ? (
                <div className="pt-4 pb-1">
                  <hr className="border-gray-700 mx-1" />
                </div>
              ) : (
                <div className="pt-4 pb-1 px-3">
                  <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Administração</p>
                </div>
              )}
              <NavLink
                to="/admin/users"
                onClick={closeMobile}
                className={({ isActive }) =>
                  `flex items-center ${collapsed ? 'justify-center' : 'gap-3'} px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-gray-800 text-white'
                      : 'text-gray-400 hover:text-white hover:bg-gray-800'
                  }`
                }
                title={collapsed ? 'Usuários' : undefined}
              >
                <HiUserGroup className="h-5 w-5 flex-shrink-0" />
                {!collapsed && <span className="whitespace-nowrap overflow-hidden">Usuários</span>}
              </NavLink>
            </>
          )}
        </nav>

        {/* User Info + Logout */}
        <div className={`py-4 border-t border-gray-800 ${collapsed ? 'px-2' : 'px-3'}`}>
          {user && (
            <button
              onClick={() => { navigate('/profile'); closeMobile(); }}
              className={`flex items-center ${collapsed ? 'justify-center' : 'justify-between'} w-full px-3 py-2 mb-2 rounded-lg hover:bg-gray-800 transition-colors group text-left`}
              title={collapsed ? `${user.name} - Perfil` : undefined}
            >
              {collapsed ? (
                <HiPencilSquare className="h-5 w-5 text-gray-400 group-hover:text-gray-300" />
              ) : (
                <>
                  <div className="min-w-0">
                    <p className="text-sm font-medium text-white truncate">{user.name}</p>
                    <p className="text-xs text-gray-400 truncate">{user.email}</p>
                  </div>
                  <HiPencilSquare className="h-4 w-4 text-gray-500 group-hover:text-gray-300 flex-shrink-0 ml-2" />
                </>
              )}
            </button>
          )}
          <button
            onClick={logout}
            className={`flex items-center ${collapsed ? 'justify-center' : 'gap-3'} w-full px-3 py-2.5 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-800 transition-colors`}
            title={collapsed ? 'Sair' : undefined}
          >
            <HiArrowRightOnRectangle className="h-5 w-5 flex-shrink-0" />
            {!collapsed && 'Sair'}
          </button>
        </div>

        {/* Collapse toggle (desktop only) */}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="hidden md:flex items-center justify-center w-full py-3 text-gray-500 hover:text-white hover:bg-gray-800 transition-colors border-t border-gray-800"
          title={collapsed ? 'Expandir menu' : 'Recolher menu'}
        >
          {collapsed ? <HiChevronRight className="h-5 w-5" /> : <HiChevronLeft className="h-5 w-5" />}
        </button>
      </aside>

      {/* Main Content */}
      <main
        className={`flex-1 bg-gray-50 min-h-screen transition-[margin] duration-300 pt-18 px-4 pb-4 md:pt-8 md:px-8 md:pb-8 ml-0 ${
          collapsed ? 'md:ml-16' : 'md:ml-64'
        }`}
      >
        <Outlet />
      </main>
    </div>
  );
}
