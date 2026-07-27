import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { toPng } from 'html-to-image'
import { statsApi } from '../lib/api'
import type { StorySlide } from '../lib/types'
import { Waveform } from '../components/ui/Waveform'
import { Button } from '../components/ui/Button'
import './Story.css'

const SLIDE_DURATION_MS = 5000

export function Story() {
  const [slides, setSlides] = useState<StorySlide[] | null>(null)
  const [index, setIndex] = useState(0)
  const [isPlaying, setIsPlaying] = useState(true)
  const [progress, setProgress] = useState(0)
  const slideRef = useRef<HTMLDivElement>(null)
  const rafRef = useRef(0)
  const startRef = useRef(0)
  const elapsedRef = useRef(0)
  const touchStartX = useRef(0)

  useEffect(() => {
    statsApi
      .story()
      .then((res) => setSlides(res.slides.length > 0 ? res.slides : [{ type: 'empty', title: 'No story yet' }]))
      .catch(() => setSlides([{ type: 'error', title: "Couldn't load your story" }]))
  }, [])

  function goTo(next: number) {
    if (!slides) return
    const clamped = Math.max(0, Math.min(slides.length - 1, next))
    elapsedRef.current = 0
    setProgress(0)
    setIndex(clamped)
  }

  // Drives both the segmented progress fill and auto-advance, via rAF so
  // pause/resume can pick up from the exact elapsed point rather than
  // restarting or relying on a CSS transition that can't be paused mid-flight.
  useEffect(() => {
    if (!isPlaying || !slides) return
    startRef.current = performance.now() - elapsedRef.current

    function tick(now: number) {
      const elapsed = now - startRef.current
      elapsedRef.current = elapsed
      const pct = Math.min(100, (elapsed / SLIDE_DURATION_MS) * 100)
      setProgress(pct)
      if (pct >= 100) {
        if (slides && index < slides.length - 1) {
          goTo(index + 1)
        } else {
          setIsPlaying(false)
        }
        return
      }
      rafRef.current = requestAnimationFrame(tick)
    }
    rafRef.current = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(rafRef.current)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isPlaying, index, slides])

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === 'ArrowLeft') goTo(index - 1)
      else if (e.key === 'ArrowRight' || e.key === ' ') {
        e.preventDefault()
        goTo(index + 1)
      } else if (e.key === 'Enter') setIsPlaying((p) => !p)
    }
    document.addEventListener('keydown', onKey)
    return () => document.removeEventListener('keydown', onKey)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [index, slides])

  function handleClick(e: React.MouseEvent<HTMLDivElement>) {
    if ((e.target as HTMLElement).closest('.story-controls, .story-exit')) return
    const half = window.innerWidth / 2
    goTo(e.clientX > half ? index + 1 : index - 1)
  }

  function handleTouchStart(e: React.TouchEvent) {
    touchStartX.current = e.changedTouches[0].screenX
  }
  function handleTouchEnd(e: React.TouchEvent) {
    const diff = touchStartX.current - e.changedTouches[0].screenX
    if (Math.abs(diff) > 50) goTo(diff > 0 ? index + 1 : index - 1)
  }

  async function exportSlide() {
    if (!slideRef.current) return
    const dataUrl = await toPng(slideRef.current, { backgroundColor: '#141217' })
    const link = document.createElement('a')
    link.download = 'spotify-insights-story.png'
    link.href = dataUrl
    link.click()
  }

  if (!slides) {
    return (
      <div className="story-page" style={{ display: 'grid', placeItems: 'center' }}>
        <Waveform variant="loading" label="Loading your story" />
      </div>
    )
  }

  const slide = slides[index]
  const isLast = index === slides.length - 1

  return (
    <div
      className="story-page"
      onClick={handleClick}
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      <Link to="/dashboard" className="story-exit">
        Close
      </Link>

      <div className="story-progress">
        {slides.map((_, i) => (
          <div key={i} className="story-progress-segment">
            <div
              className="story-progress-fill"
              style={{
                width: `${i < index ? 100 : i === index ? progress : 0}%`,
              }}
            />
          </div>
        ))}
      </div>

      <div className="story-slide" ref={slideRef}>
        <h1>{slide.title}</h1>
        {slide.value && <div className="story-slide-value">{slide.value}</div>}
        {slide.caption && <p className="story-slide-caption">{slide.caption}</p>}
      </div>

      <div className="story-controls">
        <button className="story-control-btn" onClick={() => goTo(index - 1)} aria-label="Previous slide">
          ‹
        </button>
        <button
          className="story-control-btn"
          onClick={() => setIsPlaying((p) => !p)}
          aria-label={isPlaying ? 'Pause' : 'Play'}
        >
          {isPlaying ? '❚❚' : '▶'}
        </button>
        <button className="story-control-btn" onClick={() => goTo(index + 1)} aria-label="Next slide">
          ›
        </button>
        {isLast && (
          <Button variant="secondary" onClick={exportSlide}>
            Export PNG
          </Button>
        )}
      </div>
    </div>
  )
}
