import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { ApiError, statsApi } from '../lib/api'
import { useAuth } from '../lib/auth'
import type { GenresResponse, HeatmapResponse, Period, TopArtistsResponse } from '../lib/types'
import { Card } from '../components/ui/Card'
import { SegmentedControl } from '../components/ui/SegmentedControl'
import { Waveform } from '../components/ui/Waveform'
import { RankedBarChart } from '../components/charts/RankedBarChart'
import { Heatmap } from '../components/charts/Heatmap'
import './Dashboard.css'

const PERIOD_OPTIONS: { value: Period; label: string }[] = [
  { value: 'short_term', label: 'Last 4 weeks' },
  { value: 'medium_term', label: 'Last 6 months' },
  { value: 'long_term', label: 'All time' },
]

export function Dashboard() {
  const { user } = useAuth()
  const [period, setPeriod] = useState<Period>('medium_term')
  const [topArtists, setTopArtists] = useState<TopArtistsResponse | null>(null)
  const [genres, setGenres] = useState<GenresResponse | null>(null)
  const [heatmap, setHeatmap] = useState<HeatmapResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    setTopArtists(null)
    setGenres(null)
    setError(null)

    Promise.all([statsApi.topArtists(period), statsApi.genres(period)])
      .then(([artistsRes, genresRes]) => {
        if (cancelled) return
        setTopArtists(artistsRes)
        setGenres(genresRes)
      })
      .catch((err) => {
        if (cancelled) return
        setError(err instanceof ApiError ? err.message : 'Failed to load dashboard data')
      })

    return () => {
      cancelled = true
    }
  }, [period])

  useEffect(() => {
    statsApi
      .heatmap()
      .then(setHeatmap)
      .catch(() => setHeatmap({ cells: [] }))
  }, [])

  return (
    <div className="dashboard-page">
      <div className="dashboard-header">
        <div className="dashboard-brand">
          <span className="mono">SPOTIFY INSIGHTS</span>
          <h1 style={{ fontSize: '1.5rem' }}>{user ? `${user.displayName}'s dashboard` : 'Dashboard'}</h1>
        </div>
        <div className="dashboard-header-actions">
          <Link className="story-link" to="/story">
            View your Story →
          </Link>
          <SegmentedControl options={PERIOD_OPTIONS} value={period} onChange={setPeriod} />
        </div>
      </div>

      <div className="dashboard-divider">
        <Waveform variant="divider" />
      </div>

      {error && <p style={{ color: 'var(--signal-coral)' }}>{error}</p>}

      <div className="dashboard-grid">
        <Card label="CH.01" title="Top Artists">
          {topArtists ? (
            <RankedBarChart
              items={topArtists.artists.map((a) => ({ label: a.artist_name, value: a.play_count }))}
            />
          ) : (
            <Waveform variant="loading" label="Loading top artists" />
          )}
        </Card>

        <Card label="CH.02" title="Genres">
          {genres ? (
            <RankedBarChart items={genres.genres.map((g) => ({ label: g.genre, value: g.count }))} />
          ) : (
            <Waveform variant="loading" label="Loading genres" />
          )}
        </Card>

        <Card label="CH.03" title="Time of Day">
          {heatmap ? <Heatmap cells={heatmap.cells} /> : <Waveform variant="loading" label="Loading listening times" />}
        </Card>
      </div>
    </div>
  )
}
