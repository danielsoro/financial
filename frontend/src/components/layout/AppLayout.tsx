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

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="fixed left-0 top-0 h-screen w-64 bg-gray-900 text-white flex flex-col">
        {/* Logo / Title */}
        <div className="px-6 py-6">
          <h1 className="text-2xl font-bold tracking-tight">Finance</h1>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-3 space-y-1">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-gray-800 text-white'
                    : 'text-gray-400 hover:text-white hover:bg-gray-800'
                }`
              }
            >
              <item.icon className="h-5 w-5" />
              {item.label}
            </NavLink>
          ))}

          {/* Admin Section */}
          {isAdmin && (
            <>
              <div className="pt-4 pb-1 px-3">
                <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider">Administração</p>
              </div>
              <NavLink
                to="/admin/users"
                className={({ isActive }) =>
                  `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                    isActive
                      ? 'bg-gray-800 text-white'
                      : 'text-gray-400 hover:text-white hover:bg-gray-800'
                  }`
                }
              >
                <HiUserGroup className="h-5 w-5" />
                Usuários
              </NavLink>
            </>
          )}
        </nav>

        {/* User Info + Logout */}
        <div className="px-3 py-4 border-t border-gray-800">
          {user && (
            <button
              onClick={() => navigate('/profile')}
              className="flex items-center justify-between w-full px-3 py-2 mb-2 rounded-lg hover:bg-gray-800 transition-colors group text-left"
            >
              <div className="min-w-0">
                <p className="text-sm font-medium text-white truncate">{user.name}</p>
                <p className="text-xs text-gray-400 truncate">{user.email}</p>
              </div>
              <HiPencilSquare className="h-4 w-4 text-gray-500 group-hover:text-gray-300 flex-shrink-0 ml-2" />
            </button>
          )}
          <button
            onClick={logout}
            className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-gray-800 transition-colors"
          >
            <HiArrowRightOnRectangle className="h-5 w-5" />
            Sair
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="ml-64 flex-1 bg-gray-50 min-h-screen p-8">
        <Outlet />
      </main>
    </div>
  );
}
