import { Navigate } from 'react-router-dom'
import { useAuth } from '../lib/auth'
import { Waveform } from '../components/ui/Waveform'
import { Button } from '../components/ui/Button'
import { Card } from '../components/ui/Card'
import './Login.css'

// A visitor's only job here is the Connect Spotify click — no feature list,
// no pricing, no scroll required to find the CTA (docs/ARCHITECTURE.md §6).
export function Login() {
  const { user, loading } = useAuth()

  if (!loading && user) {
    return <Navigate to="/dashboard" replace />
  }

  return (
    <div className="login-page">
      <div className="login-hero">
        <span className="login-brand mono">SPOTIFY INSIGHTS</span>
        <Waveform variant="hero" />
        <h1>
          Your Music Story,
          <br />
          Reimagined
        </h1>
        <p className="login-tagline">
          <strong>Connect Spotify.</strong> We start listening to your listening.
        </p>
        <Button as="a" href="/auth/login">
          Connect Spotify
        </Button>
      </div>

      <div className="proof-row">
        <Card label="CH.01" title="Your Story">
          <p style={{ color: 'var(--ink-400)', fontSize: '0.85rem' }}>
            A swipeable, Wrapped-style recap of your top tracks, artists, and genres.
          </p>
        </Card>
        <Card label="CH.02" title="Now Playing Widget">
          <div className="proof-widget">
            <span className="proof-widget-accent" />
            <span>Embed a live "now playing" badge anywhere</span>
          </div>
        </Card>
        <Card label="CH.03" title="Dashboard">
          <div className="proof-bars">
            {[40, 70, 55, 90, 65, 30].map((h, i) => (
              <div key={i} className="proof-bar" style={{ height: `${h}%` }} />
            ))}
          </div>
        </Card>
      </div>
    </div>
  )
}
