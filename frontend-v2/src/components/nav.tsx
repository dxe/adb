/* When making changes to this file, be
   sure to implement the same changes in
   `frontend/AdbNav.vue`.
*/

'use client'

import navbarData from '$shared/nav.json'
import { CircleUser } from 'lucide-react'
import { Suspense, useState } from 'react'
import Image from 'next/image'
import logo1 from '$public/logo.png'
import Link from 'next/link'
import { useAuthedPageContext } from '@/hooks/useAuthedPageContext'
import buefyStyles from './nav.module.css'
import clsx from 'clsx'
import { useQueryClient, useSuspenseQuery } from '@tanstack/react-query'
import { API_PATH, apiClient } from '@/lib/api'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'

function userHasAccess(
  /** Auth'd user. */
  user: {
    role: string
    ChapterID: number
  },
  /** Array of roles that are permitted to access. A user can access if they
   * hold **any** of the required roles. If undefined, any role is permitted. */
  roleRequired: string[] | undefined,
): boolean {
  if (!roleRequired) return true

  if (!user.role) return false

  if (user.ChapterID !== SF_BAY_CHAPTER_ID) {
    // If non-sfbay chapter is active, we only show non-sfbay items or admin items.
    return (
      !roleRequired ||
      roleRequired.includes('non-sfbay') ||
      (roleRequired.includes('admin') && user.role === 'admin')
    )
  }

  return roleRequired.some((it) => {
    if (it === 'admin') return user.role === 'admin'
    if (it === 'organizer')
      return user.role === 'admin' || user.role === 'organizer'
    if (it === 'attendance')
      return (
        user.role === 'admin' ||
        user.role === 'organizer' ||
        user.role === 'attendance'
      )
    if (it === 'non-sfbay') return user.role === 'non-sfbay'
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

  if (!userHasAccess(user, item.roleRequired)) {
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
              userHasAccess(user, innerItem.roleRequired) &&
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

const ChapterSwitcher = () => {
  const { user } = useAuthedPageContext()
  const queryClient = useQueryClient()

  const { data } = useSuspenseQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

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

const ChapterSwitcherFallback = () => (
  <div className={buefyStyles['navbar-item']}>
    <span className="text-sm text-neutral-500">Loading chapters...</span>
  </div>
)

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
          {user.role === 'admin' && (
            <Suspense fallback={<ChapterSwitcherFallback />}>
              <ChapterSwitcher />
            </Suspense>
          )}
        </div>
      </div>
    </nav>
  )
}
