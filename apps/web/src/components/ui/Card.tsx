import type { ReactNode } from 'react'
import './Card.css'

interface CardProps {
  label?: string
  title: string
  children: ReactNode
}

/** A "channel strip" card — the dashboard's mixing-console layout unit. */
export function Card({ label, title, children }: CardProps) {
  return (
    <div className="card">
      <div className="card-header">
        {label && <span className="card-label mono">{label}</span>}
        <h3 className="card-title">{title}</h3>
      </div>
      <div className="card-body">{children}</div>
    </div>
  )
}
