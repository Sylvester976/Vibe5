import type { AnchorHTMLAttributes, ButtonHTMLAttributes, ReactNode } from 'react'
import './Button.css'

type Variant = 'primary' | 'secondary'

type ButtonAsButton = ButtonHTMLAttributes<HTMLButtonElement> & {
  as?: 'button'
  variant?: Variant
  children: ReactNode
}

type ButtonAsAnchor = AnchorHTMLAttributes<HTMLAnchorElement> & {
  as: 'a'
  variant?: Variant
  children: ReactNode
}

type ButtonProps = ButtonAsButton | ButtonAsAnchor

export function Button({ as = 'button', variant = 'primary', className, children, ...rest }: ButtonProps) {
  const classes = `btn btn-${variant}${className ? ` ${className}` : ''}`
  if (as === 'a') {
    return (
      <a className={classes} {...(rest as AnchorHTMLAttributes<HTMLAnchorElement>)}>
        {children}
      </a>
    )
  }
  return (
    <button className={classes} {...(rest as ButtonHTMLAttributes<HTMLButtonElement>)}>
      {children}
    </button>
  )
}
