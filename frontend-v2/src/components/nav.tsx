/* When making changes to this file, be
   sure to implement the same changes in
   `frontend/AdbNav.vue`.
*/

'use client'

import navbarData from '$shared/nav.json'
import { CircleUser } from 'lucide-react'
import { useState, useMemo } from 'react'
import Image from 'next/image'
import logo1 from '$public/logo.png'
import Link from 'next/link'
import { usePathname, useSearchParams } from 'next/navigation'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import buefyStyles from './nav.module.css'
import clsx from 'clsx'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { API_PATH, apiClient } from '@/lib/api'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'

type NavAccessItem = {
  roleRequired?: string[]
  visibleForNonSFBay?: boolean
}

function toNavAccessItem(item: object | undefined): NavAccessItem | undefined {
  if (!item) return undefined

  return {
    roleRequired:
      'roleRequired' in item
        ? (item.roleRequired as string[] | undefined)
        : undefined,
    visibleForNonSFBay:
      'visibleForNonSFBay' in item
        ? (item.visibleForNonSFBay as boolean | undefined)
        : undefined,
  }
}

function userHasAccess(
  /** Auth'd user. */
  user: {
    role: string
    ChapterID: number
  },
  item: NavAccessItem | undefined,
): boolean {
  if (!item?.roleRequired) return true

  if (!user.role) return false

  const hasRequiredRole = item.roleRequired.some((it) => {
    if (it === 'admin') return user.role === 'admin'
    if (it === 'organizer')
      return user.role === 'admin' || user.role === 'organizer'
    if (it === 'attendance')
      return (
        user.role === 'admin' ||
        user.role === 'organizer' ||
        user.role === 'attendance'
      )
    return false
  })

  if (user.ChapterID !== SF_BAY_CHAPTER_ID) {
    // Outside SF Bay, keep the limited non-SF Bay view while still honoring the user's stored roles.
    return item.roleRequired.includes('admin')
      ? user.role === 'admin'
      : Boolean(item.visibleForNonSFBay) && hasRequiredRole
  }

  return hasRequiredRole
}

/** For ActivistListV2 pages, check if the nav item's query params match the current URL exactly. */
function isExactParamsMatch(
  navHref: string,
  pathname: string,
  currentSearchParams: URLSearchParams,
): boolean {
  const qIndex = navHref.indexOf('?')
  const navPath = qIndex >= 0 ? navHref.substring(0, qIndex) : navHref
  if (pathname !== navPath) return false
  const navParams = new URLSearchParams(
    qIndex >= 0 ? navHref.substring(qIndex + 1) : '',
  )
  const sortEntries = (params: URLSearchParams) =>
    Array.from(params.entries())
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([k, v]) => `${k}=${v}`)
      .join('&')
  return sortEntries(navParams) === sortEntries(currentSearchParams)
}

type TDropdownItem = (typeof navbarData.items)[number]

const DropdownItem = ({
  item,
  isExpanded,
  onClick,
  onNavigate,
}: {
  item: TDropdownItem
  isExpanded: boolean
  onClick: () => void
  onNavigate: () => void
}) => {
  const { user } = useAuthedPageContext()
  const pathname = usePathname()
  const searchParams = useSearchParams()

  const accessibleItems = useMemo(
    () =>
      isExpanded
        ? item.items.filter((innerItem) =>
            userHasAccess(user, toNavAccessItem(innerItem)),
          )
        : null,
    [isExpanded, item.items, user],
  )

  // Suppress prefix-matching if any sibling exactly matches the current path,
  // so e.g. "All Events" doesn't also highlight when on "New Event".
  const hasExactPathMatch = useMemo(
    () =>
      accessibleItems?.some(({ href }) => {
        const navPath = (
          href.startsWith('/v2') ? href.substring(3) : href
        ).split('?')[0]
        return pathname === navPath
      }) ?? false,
    [accessibleItems, pathname],
  )

  const childrenItems = useMemo(
    () =>
      accessibleItems?.map((innerItem) => {
        const navHref = innerItem.href.startsWith('/v2')
          ? innerItem.href.substring(3)
          : innerItem.href
        const navPath = navHref.split('?')[0]
        // Items with query params (e.g. activist presets) require exact path + params.
        // Plain-path items use exact or prefix match (prefix suppressed if a sibling is more specific).
        const isActive = navHref.includes('?')
          ? isExactParamsMatch(navHref, pathname, searchParams)
          : pathname === navPath ||
            (!hasExactPathMatch && pathname.startsWith(navPath + '/'))
        return { innerItem, isActive }
      }) ?? null,
    [accessibleItems, hasExactPathMatch, pathname, searchParams],
  )

  if (!userHasAccess(user, toNavAccessItem(item))) {
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
      {childrenItems && (
        <div
          className={clsx(buefyStyles['navbar-dropdown'], '!block')}
          onClick={onNavigate}
        >
          {childrenItems.map(({ innerItem, isActive }) => {
            const classNames = clsx(
              buefyStyles['navbar-item'],
              { [buefyStyles['is-active']]: isActive },
              { 'mb-2': innerItem.separatorBelow },
            )
            return innerItem.href.startsWith('/v2') ? (
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
            )
          })}
        </div>
      )}
    </div>
  )
}

const ChapterSwitcher = () => {
  const { user } = useAuthedPageContext()
  const queryClient = useQueryClient()

  const { data, isLoading, isError } = useQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: ({ signal }) => apiClient.getChapterList(signal),
  })

  if (isLoading) {
    return (
      <div
        className={buefyStyles['navbar-item']}
        role="status"
        aria-live="polite"
      >
        <span className="text-sm text-neutral-500">Loading chapters...</span>
      </div>
    )
  }

  if (isError) {
    return (
      <div className={buefyStyles['navbar-item']} role="alert">
        <span className="text-sm text-neutral-500">Chapters unavailable</span>
      </div>
    )
  }

  const switchChapter = (e: React.ChangeEvent<HTMLSelectElement>) => {
    queryClient.invalidateQueries() // invalidate existing cache for previous chapter
    window.location.href = `/auth/switch_chapter?chapter_id=${e.target.value}`
  }

  return (
    <div className={buefyStyles['navbar-item']}>
      {/* TODO(jh): use a better styled select component here eventually. */}
      <select
        onChange={switchChapter}
        value={user.ChapterID}
        className="cursor-pointer rounded-lg border border-input pl-3 pr-8 py-1.5 text-sm bg-white hover:border-gray-400 focus:border-primary focus:outline-none focus:ring-1 focus:ring-ring transition-colors"
        disabled={!data?.length}
      >
        {data?.map((chapter) => (
          <option key={chapter.ChapterID} value={chapter.ChapterID}>
            {chapter.Name}
          </option>
        ))}
      </select>
    </div>
  )
}

export const Navbar = () => {
  const { user } = useAuthedPageContext()
  const [isMobileExpanded, setMobileExpanded] = useState(false)
  const [activeDropdown, setActiveDropdown] = useState<string | null>(null)

  const closeNav = () => {
    setMobileExpanded(false)
    setActiveDropdown(null)
  }

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
              onNavigate={closeNav}
            />
          ))}
        </div>

        <div className={buefyStyles['navbar-end']}>
          <div className={buefyStyles['navbar-item']}>
            <div className="flex gap-4 items-center">
              <div className="flex items-center gap-2">
                <CircleUser className="text-neutral-600" size={20} />
                <span>
                  <span className="font-medium">{user.Name}</span>
                  {/* Admins see the ChapterSwitcher instead of the active chapter after the user name. */}
                  {user.role !== 'admin' && (
                    <>
                      {' '}
                      <span className="text-neutral-500 text-sm">
                        ({user.ChapterName})
                      </span>
                    </>
                  )}
                </span>
              </div>
              <div className="h-5 w-px bg-neutral-300" />
              <a
                href="/logout"
                className="text-sm text-neutral-700 hover:text-neutral-900 hover:underline transition-colors"
              >
                Log out
              </a>
            </div>
          </div>
          {user.role === 'admin' && <ChapterSwitcher />}
        </div>
      </div>
    </nav>
  )
}
