'use client'

import Link from 'next/link'
import { useMemo } from 'react'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { API_PATH, apiClient, Role, User, UserWithoutId } from '@/lib/api'
import { Loader2 } from 'lucide-react'

const userFormSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  email: z.string().trim().email('Enter a valid email'),
  // Make nullable so a chapter isn't chosen arbitrarily during new user creation. The user might not notice that
  // a chapter was selected automatically, and they may save the new user under the wrong chapter.
  chapterId: z.number().int().nullable(),
  disabled: z.boolean(),
  roles: z.array(Role),
})

const userFormSubmitSchema = userFormSchema.extend({
  // Make chapter non-nullable at submit time.
  chapterId: z.number({ message: 'Select a chapter' }).int(),
})

type UserFormValues = z.infer<typeof userFormSchema>

export function UserForm({ userId }: { userId?: number }) {
  const router = useRouter()
  const queryClient = useQueryClient()

  const {
    data: chapters,
    isLoading: isChaptersLoading,
    isError: isChaptersError,
    error: chaptersError,
  } = useQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  const {
    data: user,
    isLoading: userLoading,
    isError: userError,
    error: userErrorObj,
  } = useQuery<User | undefined>({
    queryKey: [API_PATH.USERS, userId],
    queryFn: userId ? () => apiClient.getUser(userId) : async () => undefined,
    enabled: !!userId,
  })

  const mutation = useMutation({
    mutationFn: async (payload: UserWithoutId) => {
      if (userId) {
        return apiClient.updateUser({ ...payload, id: userId })
      }
      return apiClient.createUser(payload)
    },
    onSuccess: (savedUser) => {
      queryClient.invalidateQueries({ queryKey: [API_PATH.USERS] })
      queryClient.setQueryData([API_PATH.USERS, savedUser.id], savedUser)
      toast.success(`${savedUser.email} saved`)
      router.push('/users')
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Unable to save user right now'
      toast.error(message)
    },
  })

  const initialValues: UserFormValues = useMemo(
    () => ({
      name: user?.name ?? '',
      email: user?.email ?? '',
      chapterId: user?.chapter_id ?? null,
      disabled: user?.disabled ?? false,
      roles: user?.roles ?? [],
    }),
    [user],
  )

  const form = useForm<UserFormValues>({
    defaultValues: initialValues,
    validators: {
      onSubmit: userFormSubmitSchema,
    },
    onSubmit: async ({ value }) => {
      const parsed = userFormSubmitSchema.safeParse(value)
      if (!parsed.success) {
        toast.error('Please fix the highlighted errors.')
        return
      }

      const payload: UserWithoutId = {
        name: parsed.data.name,
        email: parsed.data.email,
        chapter_id: parsed.data.chapterId,
        disabled: parsed.data.disabled,
        roles: parsed.data.roles,
      }

      await mutation.mutateAsync(payload)
    },
  })

  const isSubmitting = mutation.isPending

  const renderRoleLabel = (role: Role) => {
    if (role === 'non-sfbay') {
      return 'Non SF Bay'
    }
    return role.charAt(0).toUpperCase() + role.slice(1)
  }

  const isLoading = isChaptersLoading || userLoading
  const loadError = isChaptersError || userError
  const loadErrorMessage =
    (chaptersError as Error | undefined)?.message ||
    (userErrorObj as Error | undefined)?.message ||
    'Unable to load user data.'

  if (loadError) {
    return (
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-2">
          <div className="flex flex-col gap-1">
            <p className="text-sm uppercase tracking-wide text-muted-foreground">
              {userId ? 'Edit user' : 'Create user'}
            </p>
            <h1 className="text-2xl font-semibold">Users</h1>
          </div>
          <Button variant="ghost" asChild>
            <Link href="/users">Back to list</Link>
          </Button>
        </div>
        <div className="rounded-md border p-4 text-sm text-destructive">
          {loadErrorMessage}
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-2">
        <div className="flex flex-col gap-1">
          <p className="text-sm uppercase tracking-wide text-muted-foreground">
            {userId ? 'Edit user' : 'Create user'}
          </p>
          <h1 className="text-2xl font-semibold">{user?.name ?? 'New User'}</h1>
        </div>
        <Button variant="ghost" asChild>
          <Link href="/users">Back to list</Link>
        </Button>
      </div>

      {isLoading && (
        <div className="flex items-center gap-2 text-muted-foreground text-sm">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading...
        </div>
      )}

      {!isLoading && (
        <form
          onSubmit={(e) => {
            e.preventDefault()
            form.handleSubmit()
          }}
          className="space-y-6"
        >
          <div className="flex items-center justify-end">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin mr-2" />
                  Saving...
                </>
              ) : (
                'Save user'
              )}
            </Button>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <form.Field name="name">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="user-name">Name</Label>
                  <Input
                    id="user-name"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    placeholder="Full name"
                  />
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="email">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="user-email">Email</Label>
                  <Input
                    id="user-email"
                    type="email"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    placeholder="user@example.org"
                  />
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="chapterId">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="user-chapter">Chapter</Label>
                  <Select
                    value={
                      field.state.value !== null
                        ? String(field.state.value)
                        : undefined
                    }
                    onValueChange={(value: string) => {
                      const next = parseInt(value, 10)
                      field.handleChange(Number.isNaN(next) ? null : next)
                    }}
                    disabled={isChaptersLoading}
                  >
                    <SelectTrigger id="user-chapter">
                      <SelectValue placeholder="Select a chapter" />
                    </SelectTrigger>
                    <SelectContent>
                      {chapters?.map((chapter) => (
                        <SelectItem
                          key={chapter.ChapterID}
                          value={String(chapter.ChapterID)}
                        >
                          {chapter.Name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="disabled">
              {(field) => (
                <div className="space-y-2 rounded-md border p-3">
                  <Label className="flex items-center gap-2 text-sm">
                    <Checkbox
                      checked={field.state.value}
                      onCheckedChange={(checked: boolean | 'indeterminate') =>
                        field.handleChange(Boolean(checked))
                      }
                    />
                    Disabled
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Disabled users cannot sign in. Use this to suspend access
                    without deleting their account.
                  </p>
                </div>
              )}
            </form.Field>
          </div>

          <form.Field name="roles">
            {(field) => (
              <div className="space-y-2">
                <Label>Roles</Label>
                <div className="flex max-w-xs flex-col gap-2">
                  {Role.options.map((role) => {
                    const checked = field.state.value.includes(role)
                    return (
                      <label
                        key={role}
                        className="flex w-full items-center gap-2 rounded-md border p-3 text-sm"
                      >
                        <Checkbox
                          checked={checked}
                          onCheckedChange={() => {
                            const nextRoles = checked
                              ? field.state.value.filter((r) => r !== role)
                              : [...field.state.value, role]
                            field.handleChange(nextRoles)
                          }}
                        />
                        <span>{renderRoleLabel(role)}</span>
                      </label>
                    )
                  })}
                </div>
                {field.state.meta.errors[0] && (
                  <p className="text-sm text-destructive">
                    {field.state.meta.errors[0]}
                  </p>
                )}
              </div>
            )}
          </form.Field>
        </form>
      )}
    </div>
  )
}
