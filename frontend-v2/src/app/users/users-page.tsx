'use client'

import { useMemo, useState } from 'react'
import Link from 'next/link'
import { useQuery } from '@tanstack/react-query'
import { API_PATH, apiClient } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Loader2 } from 'lucide-react'
import { UserTable } from './user-table'

export default function UsersPage() {
  const { data: users, isLoading: isUsersLoading } = useQuery({
    queryKey: [API_PATH.USERS],
    queryFn: apiClient.getUsers,
  })

  const { data: chapters, isLoading: isChaptersLoading } = useQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  const [showDisabled, setShowDisabled] = useState(false)

  const filteredUsers = useMemo(() => {
    if (!users) return []
    if (showDisabled) return users
    return users.filter((user) => !user.disabled)
  }, [showDisabled, users])

  const isLoading = isUsersLoading || isChaptersLoading

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold">Users</h1>
          <p className="text-muted-foreground text-sm">
            Manage accounts, roles, and disabled status.
          </p>
        </div>
        <Button asChild>
          <Link href="/users/new">New user</Link>
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <Checkbox
          id="show-disabled-users"
          checked={showDisabled}
          onCheckedChange={(value) => setShowDisabled(Boolean(value))}
        />
        <label
          htmlFor="show-disabled-users"
          className="text-sm text-muted-foreground"
        >
          Show disabled users
        </label>
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 text-muted-foreground text-sm">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading users...
        </div>
      ) : users && chapters ? (
        <UserTable users={filteredUsers} chapters={chapters} />
      ) : (
        <div className="text-sm text-destructive">
          Unable to load users right now.
        </div>
      )}
    </div>
  )
}
