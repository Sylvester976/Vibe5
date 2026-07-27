import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from './auth'
import { Waveform } from '../components/ui/Waveform'

export function ProtectedRoute() {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <div style={{ display: 'grid', placeItems: 'center', height: '100vh' }}>
        <Waveform variant="loading" label="Checking your session" />
      </div>
    )
  }
  if (!user) {
    return <Navigate to="/login" replace />
  }
  return <Outlet />
}
