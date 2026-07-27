import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuth } from './lib/auth'
import { ProtectedRoute } from './lib/ProtectedRoute'
import { Waveform } from './components/ui/Waveform'
import { Login } from './pages/Login'
import { Dashboard } from './pages/Dashboard'
import { Story } from './pages/Story'

function RootRedirect() {
  const { user, loading } = useAuth()
  if (loading) {
    return (
      <div style={{ display: 'grid', placeItems: 'center', height: '100vh' }}>
        <Waveform variant="loading" />
      </div>
    )
  }
  return <Navigate to={user ? '/dashboard' : '/login'} replace />
}

export function App() {
  return (
    <Routes>
      <Route path="/" element={<RootRedirect />} />
      <Route path="/login" element={<Login />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/story" element={<Story />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
