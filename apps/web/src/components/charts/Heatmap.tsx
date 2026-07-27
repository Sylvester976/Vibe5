import { Fragment, useMemo, useState } from 'react'
import type { HeatmapCell } from '../../lib/types'
import './Heatmap.css'

const DAY_LABELS = ['S', 'M', 'T', 'W', 'T', 'F', 'S']
const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday']

interface HeatmapProps {
  cells: HeatmapCell[]
}

/** Listening-by-day/hour grid — magnitude only, so one hue (coral) at
 * increasing intensity, never a multi-hue categorical palette. */
export function Heatmap({ cells }: HeatmapProps) {
  const [hovered, setHovered] = useState<{ day: number; hour: number; count: number } | null>(null)

  const { matrix, max } = useMemo(() => {
    const m: number[][] = Array.from({ length: 7 }, () => Array(24).fill(0))
    let maxCount = 0
    for (const cell of cells) {
      if (cell.day_of_week >= 0 && cell.day_of_week < 7 && cell.hour >= 0 && cell.hour < 24) {
        m[cell.day_of_week][cell.hour] = cell.count
        maxCount = Math.max(maxCount, cell.count)
      }
    }
    return { matrix: m, max: maxCount }
  }, [cells])

  if (max === 0) {
    return <p style={{ color: 'var(--ink-400)', fontSize: '0.85rem' }}>No listening history yet</p>
  }

  return (
    <div className="heatmap">
      <div className="heatmap-readout">
        {hovered ? (
          <>
            {DAY_NAMES[hovered.day]} {String(hovered.hour).padStart(2, '0')}:00 —{' '}
            <span className="value">{hovered.count} plays</span>
          </>
        ) : (
          <span style={{ color: 'var(--ink-400)' }}>Hover a cell for detail</span>
        )}
      </div>

      <div className="heatmap-grid">
        {matrix.map((row, day) => (
          <Fragment key={day}>
            <span className="heatmap-day-label">{DAY_LABELS[day]}</span>
            {row.map((count, hour) => {
              const intensity = count / max
              return (
                <div
                  key={`${day}-${hour}`}
                  className="heatmap-cell"
                  tabIndex={0}
                  style={{ background: cellColor(intensity) }}
                  onMouseEnter={() => setHovered({ day, hour, count })}
                  onFocus={() => setHovered({ day, hour, count })}
                  onMouseLeave={() => setHovered(null)}
                  onBlur={() => setHovered(null)}
                  aria-label={`${DAY_NAMES[day]} ${hour}:00, ${count} plays`}
                />
              )
            })}
          </Fragment>
        ))}
      </div>

      <div className="heatmap-hours">
        <span>00:00</span>
        <span>06:00</span>
        <span>12:00</span>
        <span>18:00</span>
      </div>
    </div>
  )
}

// Sequential ramp: one hue (coral), monotonically increasing alpha over the
// dark surface so "more" reads as brighter — the dark-mode-only equivalent
// of a light-to-dark sequential scale.
function cellColor(intensity: number): string {
  if (intensity === 0) return 'var(--bg-base)'
  const alpha = 0.15 + intensity * 0.85
  return `rgb(255 107 74 / ${alpha})`
}
