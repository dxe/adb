'use client'

import navbarData from '../../../frontend/nav.json'
import { User } from 'lucide-react'
import { useState } from 'react'
import Image from 'next/image'
import logo1 from '../../../frontend/static/img/logo1.png'
import { cn } from '@/lib/utils'
import Link from 'next/link'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'

function userHasAccess(
  /** Role the auth'd user has. */
  userRole: string,
  /** Array of roles that are permitted to access. A user can access if they
   * hold **any** of the required roles. If undefined, any role is permitted. */
  roleRequired: string[] | undefined,
): boolean {
  if (!roleRequired) return true

  if (!userRole) return false

  return roleRequired.some((it) => {
    if (it === 'admin') return userRole === 'admin'
    if (it === 'organizer')
      return userRole === 'admin' || userRole === 'organizer'
    if (it === 'attendance')
      return (
        userRole === 'admin' ||
        userRole === 'organizer' ||
        userRole === 'attendance'
      )
    if (it === 'non-sfbay') return userRole === 'non-sfbay'
    return false
  })
}

type TDropdownItem = (typeof navbarData.items)[number]

const DropdownItem = ({
  item,
  isExpanded,
  onClick,
}: {
  item: TDropdownItem
  isExpanded: boolean
  onClick: () => void
}) => {
  const { user, pageName } = useAuthedPageContext()

  if (!userHasAccess(user.role, item.roleRequired)) {
    return null
  }
  return (
    <div className="navbar-item has-dropdown">
      {/* Important: This must remain an `a` element and not a `button` for the buefy mobile styles to work properly. */}
      <a
        role="menuitem"
        aria-haspopup
        className="navbar-link"
        onClick={(e) => {
          e.preventDefault()
          onClick()
        }}
      >
        {item.label}
      </a>
      {isExpanded && (
        <div className="navbar-dropdown !block" onClick={onClick}>
          {item.items.map((innerItem) => {
            const classNames = cn(
              'navbar-item',
              { 'is-active': pageName === innerItem.page },
              { 'mb-2': innerItem.separatorBelow },
            )
            return (
              userHasAccess(user.role, innerItem.roleRequired) &&
              (innerItem.href.startsWith('/v2') ? (
                <Link
                  href={innerItem.href.substring(3)}
                  className={classNames}
                  key={innerItem.href}
                >
                  {innerItem.label}
                </Link>
              ) : (
                <a
                  href={innerItem.href}
                  className={classNames}
                  key={innerItem.href}
                >
                  {innerItem.label}
                </a>
              ))
            )
          })}
        </div>
      )}
    </div>
  )
}

export const Navbar = () => {
  const { user } = useAuthedPageContext()
  const [isMobileExpanded, setMobileExpanded] = useState(false)
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null)

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
          {/* These spans are what make the hamburger icon on mobile (via buefy styles). */}
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
        </button>
      </div>

      <div className={cn('navbar-menu', { 'is-active': isMobileExpanded })}>
        <div className="navbar-start">
          {navbarData.items.map((item) => (
            <DropdownItem
              key={item.label}
              item={item}
              isExpanded={activeDropdown === item.label}
              onClick={() =>
                setActiveDropdown((prev) =>
                  prev === item.label ? null : item.label,
                )
              }
            />
          ))}
        </div>

        <div className="navbar-end">
          <div className="navbar-item">
            <div className="flex gap-3 justify-between">
              <div className="has-text-grey-dark flex items-center gap-2">
                <span className="icon is-small">
                  <User />
                </span>
                {user.Name} ({user.ChapterName})
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
