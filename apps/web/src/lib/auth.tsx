import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { ApiError, meApi } from './api'

interface AuthUser {
  displayName: string
}

interface AuthState {
  user: AuthUser | null
  loading: boolean
  refresh: () => Promise<void>
}

const AuthContext = createContext<AuthState | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [loading, setLoading] = useState(true)

  const refresh = async () => {
    try {
      const me = await meApi.get()
      setUser({ displayName: me.display_name })
    } catch (err) {
      // A 401 just means "not logged in" — anything else is unexpected but
      // still resolves to "treat as logged out" for the purposes of routing.
      if (!(err instanceof ApiError) || err.status !== 401) {
        console.error('failed to load session', err)
      }
      setUser(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
  }, [])

  return <AuthContext.Provider value={{ user, loading, refresh }}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider')
  return ctx
}
