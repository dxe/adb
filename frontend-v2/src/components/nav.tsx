'use client'

import navbarData from '../../../frontend/nav.json'
import { fetchSession } from '@/app/session'
import { User } from 'lucide-react'
import { useCallback, useState } from 'react'
import Image from 'next/image'
import logo1 from '../../../frontend/static/img/logo1.png'
import { cn } from '@/lib/utils'
import Link from 'next/link'

export const Navbar = ({
  pageName,
  session,
}: {
  /** The name of the active page, corresponding to the name in Vue. */
  pageName: string
  /** The user sesion returned from the `fetchSession` function. */
  session: Awaited<ReturnType<typeof fetchSession>>
}) => {
  const [isMobileExpanded, setMobileExpanded] = useState(false)
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null)

  const hasAccess = useCallback(
    (roleRequired: string[] | undefined) => {
      if (!roleRequired) return true

      const role = session?.user?.role
      if (!role) return false

      return roleRequired.some((it) => {
        if (it === 'admin') return role === 'admin'
        if (it === 'organizer') return role === 'admin' || role === 'organizer'
        if (it === 'attendance')
          return (
            role === 'admin' || role === 'organizer' || role === 'attendance'
          )
        if (it === 'non-sfbay') return role === 'non-sfbay'
        return false
      })
    },
    [session?.user?.role],
  )

  // Note that this navbar currently uses the Vue stylesheet. Once we
  // are no longer using Vue, we should update this using tailwind.
  return (
    <nav
      role="navigation"
      aria-label="main navigation"
      className="navbar is-fixed-top has-shadow"
      id="mainNav"
    >
      <div className="navbar-brand">
        <div className="navbar-item">
          <Image src={logo1} alt="DxE" className="w-auto h-auto" priority />
        </div>
        <button
          aria-label="menu"
          className={cn('navbar-burger burger', {
            'is-active': isMobileExpanded,
          })}
          onClick={() => setMobileExpanded((prev) => !prev)}
        >
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
        </button>
      </div>

      <div className={cn('navbar-menu', { 'is-active': isMobileExpanded })}>
        <div className="navbar-start">
          {navbarData.items.map(
            (dropdown) =>
              hasAccess(dropdown.roleRequired) && (
                <div className="navbar-item has-dropdown" key={dropdown.label}>
                  {/* Important: This must remain an `a` element and not a `button` for the buefy mobile styles to work properly. */}
                  <a
                    role="menuitem"
                    aria-haspopup
                    className="navbar-link"
                    onClick={(e) => {
                      e.preventDefault()
                      setActiveDropdown((prev) =>
                        prev === dropdown.label ? null : dropdown.label,
                      )
                    }}
                  >
                    {dropdown.label}
                  </a>
                  {activeDropdown === dropdown.label && (
                    <div
                      className={cn('navbar-dropdown !block')}
                      onClick={() => setActiveDropdown(dropdown.label)}
                    >
                      {dropdown.items.map((item) => {
                        const classNames = cn(
                          'navbar-item',
                          { 'is-active': pageName === item.page },
                          { 'mb-2': item.separatorBelow },
                        )
                        return (
                          hasAccess(item.roleRequired) &&
                          (item.href.startsWith('/v2') ? (
                            <Link
                              href={item.href.substring(3)}
                              className={classNames}
                              key={item.href}
                            >
                              {item.label}
                            </Link>
                          ) : (
                            <a
                              href={item.href}
                              className={classNames}
                              key={item.href}
                            >
                              {item.label}
                            </a>
                          ))
                        )
                      })}
                    </div>
                  )}
                </div>
              ),
          )}
        </div>

        <div className="navbar-end">
          <div className="navbar-item">
            <div className="flex gap-3 justify-between">
              <div className="has-text-grey-dark flex items-center gap-2">
                <span className="icon is-small">
                  <User />
                </span>
                {session.user?.Name} ({session.user?.ChapterName})
              </div>
              <a href="/logout" style={{ color: 'linktext' }}>
                Log out
              </a>
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}
