'use client'

import navbarData from '$shared/nav.json'
import { CircleUser } from 'lucide-react'
import { useState } from 'react'
import Image from 'next/image'
import logo1 from '$public/logo.png'
import Link from 'next/link'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import buefyStyles from './nav.module.css'
import clsx from 'clsx'

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
    <div
      className={clsx(buefyStyles['navbar-item'], buefyStyles['has-dropdown'])}
    >
      {/* Important: This must remain an `a` element and not a `button` for the buefy mobile styles to work properly. */}
      <a
        role="menuitem"
        aria-haspopup
        className={buefyStyles['navbar-link']}
        onClick={(e) => {
          e.preventDefault()
          onClick()
        }}
      >
        {item.label}
        <span className="border-[#7957d5] mt-[-0.375rem] right-[1.125rem] border-[3px] border-solid rounded-[2px] border-r-0 border-t-0 block h-[.625rem] absolute pointer-events-none top-[50%] -rotate-45 origin-center w-[.625rem]" />
      </a>
      {isExpanded && (
        <div
          className={clsx(buefyStyles['navbar-dropdown'], '!block')}
          onClick={onClick}
        >
          {item.items.map((innerItem) => {
            const classNames = clsx(
              buefyStyles['navbar-item'],
              { [buefyStyles['is-active']]: pageName === innerItem.page },
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
      className={clsx(
        buefyStyles.navbar,
        buefyStyles['is-fixed-top'],
        buefyStyles['has-shadow'],
      )}
      id="mainNav"
    >
      <div className={buefyStyles['navbar-brand']}>
        <div className={buefyStyles['navbar-item']}>
          <Image src={logo1} alt="DxE" className="w-[30.5px] h-auto" priority />
        </div>
        <button
          aria-label="menu"
          className={clsx(buefyStyles['navbar-burger'], buefyStyles['burger'], {
            [buefyStyles['is-active']]: isMobileExpanded,
          })}
          onClick={() => setMobileExpanded((prev) => !prev)}
        >
          {/* These spans are what make the hamburger icon on mobile (via buefy styles). */}
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
          <span aria-hidden="true"></span>
        </button>
      </div>

      <div
        className={clsx(buefyStyles['navbar-menu'], {
          [buefyStyles['is-active']]: isMobileExpanded,
        })}
      >
        <div className={buefyStyles['navbar-start']}>
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

        <div className={buefyStyles['navbar-end']}>
          <div className={buefyStyles['navbar-item']}>
            <div className="flex gap-3 justify-between">
              <div className="flex items-center gap-2">
                <CircleUser className="text-neutral-600" size={20} />
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
