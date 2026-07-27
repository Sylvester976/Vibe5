import './Waveform.css'

interface WaveformProps {
  variant?: 'hero' | 'loading' | 'divider'
  bars?: number
  color?: 'coral' | 'cyan'
  height?: number
  label?: string
}

/**
 * The one signature visual element (docs/ARCHITECTURE.md §6): hero on the
 * login screen, loading state anywhere data is fetched, and a divider
 * between dashboard sections. Respects prefers-reduced-motion via CSS
 * (falls back to a static bar shape rather than disabling the component).
 */
export function Waveform({
  variant = 'loading',
  bars = variant === 'divider' ? 16 : 24,
  color = 'coral',
  height = variant === 'hero' ? 96 : variant === 'divider' ? 24 : 32,
  label = 'Loading',
}: WaveformProps) {
  return (
    <div
      className={`waveform waveform--${variant} waveform--${color}`}
      style={{ height }}
      role={variant === 'loading' ? 'status' : 'presentation'}
      aria-label={variant === 'loading' ? label : undefined}
    >
      {Array.from({ length: bars }, (_, i) => (
        <span
          key={i}
          className="waveform-bar"
          style={{
            height: '100%',
            animationDelay: `${(i / bars) * 1.1}s`,
          }}
        />
      ))}
    </div>
  )
}
