'use client'

import Link from 'next/link'
import { useMemo } from 'react'
import { useForm } from '@tanstack/react-form'
import { z } from 'zod'
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from '@tanstack/react-query'
import { notFound, useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import { ArrowLeft, Loader2, Plus, Save, Trash2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { DatePicker } from '@/components/ui/date-picker'
import {
  API_PATH,
  apiClient,
  CHAPTER_ADMIN_QUERY_KEY,
  ChapterAdmin,
} from '@/lib/api'
import { COUNTRIES } from './countries'
import { formatDateYmd, parseDateYmd, REGIONS } from './chapter-utils'

// Field names match the API's PascalCase (`Name`, `Email`, ...) 1:1 so the
// organizer payload can be sent without a case-mapping step.
const organizerFormSchema = z.object({
  Name: z.string().trim().min(1, 'Organizer name is required'),
  Email: z.string().trim(),
  Phone: z.string().trim(),
  Facebook: z.string().trim(),
  Instagram: z.string().trim(),
  Twitter: z.string().trim(),
  Website: z.string().trim(),
})

const chapterFormSchema = z.object({
  name: z.string().trim().max(100),
  mentor: z.string().trim().max(100),
  notes: z.string().trim().max(512),
  region: z.string(),
  country: z.string(),
  lat: z.number().optional(),
  lng: z.number().optional(),
  fbUrl: z.string().trim().max(100),
  twitterUrl: z.string().trim().max(100),
  instaUrl: z.string().trim().max(100),
  email: z
    .string()
    .trim()
    .max(100)
    .refine(
      (v) => v === '' || z.string().email().safeParse(v).success,
      'Enter a valid email',
    ),
  lastContact: z.date().nullable(),
  lastAction: z.date().nullable(),
  organizers: z.array(organizerFormSchema),
})

// Stricter checks that only apply at submit time, mirroring the legacy Vue
// page's confirmEditChapterModal validation.
const chapterFormSubmitSchema = chapterFormSchema.extend({
  name: z.string().trim().min(1, 'Chapter name is required').max(100),
  region: z.string().min(1, 'Region is required'),
  country: z.string().min(1, 'Country is required'),
  lat: z
    .number({ message: 'Lat is required and must be a number' })
    .min(-90)
    .max(90),
  lng: z
    .number({ message: 'Lng is required and must be a number' })
    .min(-180)
    .max(180),
})

type ChapterFormValues = z.infer<typeof chapterFormSchema>

function toInitialValues(chapter: ChapterAdmin | undefined): ChapterFormValues {
  return {
    name: chapter?.Name ?? '',
    mentor: chapter?.Mentor ?? '',
    notes: chapter?.Notes ?? '',
    region: chapter?.Region ?? '',
    country: chapter?.Country ?? '',
    lat: chapter?.Lat,
    lng: chapter?.Lng,
    fbUrl: chapter?.FbURL ?? '',
    twitterUrl: chapter?.TwitterURL ?? '',
    instaUrl: chapter?.InstaURL ?? '',
    email: chapter?.Email ?? '',
    lastContact: chapter?.LastContact
      ? parseDateYmd(chapter.LastContact)
      : null,
    lastAction: chapter?.LastAction ? parseDateYmd(chapter.LastAction) : null,
    organizers: chapter?.Organizers ?? [],
  }
}

function ChapterFormInner({
  chapterId,
  chapter,
}: {
  chapterId?: number
  chapter: ChapterAdmin | undefined
}) {
  const router = useRouter()
  const queryClient = useQueryClient()
  const isEditing = chapterId != null

  const mutation = useMutation({
    mutationFn: (payload: Partial<ChapterAdmin>) =>
      apiClient.saveChapterAdmin(payload),
    onSuccess: (savedChapter) => {
      // Prefix match also refreshes the chapter-picker and intl-organizers caches.
      queryClient.invalidateQueries({ queryKey: [API_PATH.CHAPTER_LIST] })
      queryClient.setQueryData<ChapterAdmin[]>(
        CHAPTER_ADMIN_QUERY_KEY,
        (old) => {
          if (!old) return old
          const exists = old.some((c) => c.ChapterID === savedChapter.ChapterID)
          return exists
            ? old.map((c) =>
                c.ChapterID === savedChapter.ChapterID ? savedChapter : c,
              )
            : [savedChapter, ...old]
        },
      )
      toast.success(`${savedChapter.Name} saved`)
      if (isEditing) {
        router.push('/chapters')
      } else {
        router.push(`/chapters/${savedChapter.ChapterID}`)
      }
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Failed to save chapter'
      toast.error(message)
    },
  })

  const initialValues = useMemo(() => toInitialValues(chapter), [chapter])

  const form = useForm({
    defaultValues: initialValues,
    validators: {
      onSubmit: chapterFormSubmitSchema,
    },
    onSubmit: async ({ value }) => {
      const parsed = chapterFormSubmitSchema.safeParse(value)
      if (!parsed.success) {
        toast.error('Please fix the highlighted errors.')
        return
      }

      const flag =
        COUNTRIES.find((c) => c.code === parsed.data.country)?.flag ?? ''

      // Spread the original chapter first so fields the form never displays
      // (e.g. EmailToken) round-trip unchanged instead of being wiped on save.
      const payload: Partial<ChapterAdmin> = {
        ...chapter,
        Name: parsed.data.name,
        Flag: flag,
        Mentor: parsed.data.mentor,
        Notes: parsed.data.notes,
        Region: parsed.data.region,
        Country: parsed.data.country,
        Lat: parsed.data.lat,
        Lng: parsed.data.lng,
        FbURL: parsed.data.fbUrl,
        TwitterURL: parsed.data.twitterUrl,
        InstaURL: parsed.data.instaUrl,
        Email: parsed.data.email,
        LastContact: parsed.data.lastContact
          ? formatDateYmd(parsed.data.lastContact)
          : '',
        LastAction: parsed.data.lastAction
          ? formatDateYmd(parsed.data.lastAction)
          : '',
        Organizers: parsed.data.organizers,
      }

      await mutation.mutateAsync(payload)
    },
  })

  const isSubmitting = mutation.isPending

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-2">
        <div className="flex flex-col gap-1">
          <p className="text-sm uppercase tracking-wide text-muted-foreground">
            {isEditing ? 'Edit chapter' : 'Create chapter'}
          </p>
          <h1 className="text-2xl font-semibold">
            {chapter ? `${chapter.Flag} ${chapter.Name}` : 'New Chapter'}
          </h1>
        </div>
        <Button variant="ghost" asChild>
          <Link href="/chapters">
            <ArrowLeft className="h-4 w-4" />
            Back to list
          </Link>
        </Button>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault()
          form.handleSubmit()
        }}
        className="space-y-8"
      >
        <div className="flex items-center justify-end">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                Save chapter
              </>
            )}
          </Button>
        </div>

        <section className="space-y-4">
          <h2 className="text-sm font-semibold text-primary">Basic Info</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <form.Field name="name">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-name">Name</Label>
                  <Input
                    id="chapter-name"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    disabled={isEditing}
                    maxLength={100}
                  />
                  {isEditing && (
                    <p className="text-xs text-muted-foreground">
                      Chapter name can&apos;t be changed after creation.
                    </p>
                  )}
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="mentor">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-mentor">Mentor</Label>
                  <Input
                    id="chapter-mentor"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    maxLength={100}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="notes">
              {(field) => (
                <div className="space-y-1 md:col-span-2">
                  <Label htmlFor="chapter-notes">Notes</Label>
                  <Textarea
                    id="chapter-notes"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    maxLength={512}
                  />
                </div>
              )}
            </form.Field>
          </div>
        </section>

        <section className="space-y-4">
          <h2 className="text-sm font-semibold text-primary">Location</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <form.Field name="region">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-region">Region</Label>
                  <Select
                    value={field.state.value}
                    onValueChange={field.handleChange}
                  >
                    <SelectTrigger id="chapter-region">
                      <SelectValue placeholder="Select a region" />
                    </SelectTrigger>
                    <SelectContent>
                      {REGIONS.map((region) => (
                        <SelectItem key={region} value={region}>
                          {region}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="country">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-country">Country</Label>
                  <Select
                    value={field.state.value}
                    onValueChange={field.handleChange}
                  >
                    <SelectTrigger id="chapter-country">
                      <SelectValue placeholder="Select a country" />
                    </SelectTrigger>
                    <SelectContent>
                      {COUNTRIES.map((country) => (
                        <SelectItem key={country.code} value={country.code}>
                          {country.flag} {country.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="lat">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-lat">Lat</Label>
                  <Input
                    id="chapter-lat"
                    type="number"
                    inputMode="decimal"
                    step="any"
                    min={-90}
                    max={90}
                    placeholder="00.000000"
                    value={field.state.value ?? ''}
                    onChange={(e) => {
                      const n = e.target.valueAsNumber
                      field.handleChange(Number.isNaN(n) ? undefined : n)
                    }}
                    onBlur={field.handleBlur}
                  />
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>

            <form.Field name="lng">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-lng">Lng</Label>
                  <Input
                    id="chapter-lng"
                    type="number"
                    inputMode="decimal"
                    step="any"
                    min={-180}
                    max={180}
                    placeholder="000.000000"
                    value={field.state.value ?? ''}
                    onChange={(e) => {
                      const n = e.target.valueAsNumber
                      field.handleChange(Number.isNaN(n) ? undefined : n)
                    }}
                    onBlur={field.handleBlur}
                  />
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>
          </div>
        </section>

        <section className="space-y-4">
          <h2 className="text-sm font-semibold text-primary">Social Links</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <form.Field name="fbUrl">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-fb">Facebook</Label>
                  <Input
                    id="chapter-fb"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    maxLength={100}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="twitterUrl">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-twitter">Twitter</Label>
                  <Input
                    id="chapter-twitter"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    maxLength={100}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="instaUrl">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-insta">Instagram</Label>
                  <Input
                    id="chapter-insta"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    maxLength={100}
                  />
                </div>
              )}
            </form.Field>

            <form.Field name="email">
              {(field) => (
                <div className="space-y-1">
                  <Label htmlFor="chapter-email">Email (Public)</Label>
                  <Input
                    id="chapter-email"
                    type="email"
                    value={field.state.value}
                    onChange={(e) => field.handleChange(e.target.value)}
                    onBlur={field.handleBlur}
                    maxLength={100}
                  />
                  {field.state.meta.errors[0] && (
                    <p className="text-sm text-destructive">
                      {field.state.meta.errors[0]?.message}
                    </p>
                  )}
                </div>
              )}
            </form.Field>
          </div>
        </section>

        <section className="space-y-4">
          <h2 className="text-sm font-semibold text-primary">Dates</h2>
          <div className="grid gap-4 md:grid-cols-2">
            <form.Field name="lastContact">
              {(field) => (
                <div className="space-y-1">
                  <Label>Last Contact</Label>
                  <div className="flex items-center gap-2">
                    <DatePicker
                      value={field.state.value ?? undefined}
                      onValueChange={(d) => field.handleChange(d ?? null)}
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => field.handleChange(new Date())}
                    >
                      Today
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => field.handleChange(null)}
                    >
                      Clear
                    </Button>
                  </div>
                </div>
              )}
            </form.Field>

            <form.Field name="lastAction">
              {(field) => (
                <div className="space-y-1">
                  <Label>Last Action</Label>
                  <div className="flex items-center gap-2">
                    <DatePicker
                      value={field.state.value ?? undefined}
                      onValueChange={(d) => field.handleChange(d ?? null)}
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => field.handleChange(new Date())}
                    >
                      Today
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => field.handleChange(null)}
                    >
                      Clear
                    </Button>
                  </div>
                </div>
              )}
            </form.Field>
          </div>
        </section>

        <section className="space-y-4">
          <h2 className="text-sm font-semibold text-primary">Organizers</h2>
          {!isEditing ? (
            <p className="text-sm text-muted-foreground">
              Please save the new chapter before adding organizers.
            </p>
          ) : (
            <form.Field name="organizers" mode="array">
              {(arrayField) => (
                <div className="space-y-4">
                  {arrayField.state.value.length === 0 && (
                    <p className="text-sm text-muted-foreground">
                      No organizers found. Add one below.
                    </p>
                  )}

                  {arrayField.state.value.map((_, index) => (
                    <div
                      key={index}
                      className="space-y-3 rounded-md border p-4"
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="grid flex-1 gap-3 md:grid-cols-3">
                          <form.Field name={`organizers[${index}].Name`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Name</Label>
                                <Input
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Name"
                                />
                                {field.state.meta.errors[0] && (
                                  <p className="text-sm text-destructive">
                                    {field.state.meta.errors[0]?.message}
                                  </p>
                                )}
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`organizers[${index}].Email`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Email</Label>
                                <Input
                                  type="email"
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Email"
                                />
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`organizers[${index}].Phone`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Phone</Label>
                                <Input
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Phone"
                                />
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`organizers[${index}].Facebook`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Facebook</Label>
                                <Input
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Facebook"
                                  maxLength={100}
                                />
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`organizers[${index}].Instagram`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Instagram</Label>
                                <Input
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Instagram"
                                  maxLength={100}
                                />
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`organizers[${index}].Twitter`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Twitter</Label>
                                <Input
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Twitter"
                                  maxLength={100}
                                />
                              </div>
                            )}
                          </form.Field>
                          <form.Field name={`organizers[${index}].Website`}>
                            {(field) => (
                              <div className="space-y-1">
                                <Label>Website</Label>
                                <Input
                                  value={field.state.value}
                                  onChange={(e) =>
                                    field.handleChange(e.target.value)
                                  }
                                  onBlur={field.handleBlur}
                                  placeholder="Website"
                                  maxLength={100}
                                />
                              </div>
                            )}
                          </form.Field>
                        </div>
                        <Button
                          type="button"
                          variant="outline"
                          size="icon"
                          aria-label="Delete organizer"
                          onClick={() => arrayField.removeValue(index)}
                        >
                          <Trash2 className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  ))}

                  <Button
                    type="button"
                    variant="outline"
                    onClick={() =>
                      arrayField.pushValue({
                        Name: '',
                        Email: '',
                        Phone: '',
                        Facebook: '',
                        Instagram: '',
                        Twitter: '',
                        Website: '',
                      })
                    }
                  >
                    <Plus className="h-4 w-4" />
                    Add new organizer
                  </Button>
                </div>
              )}
            </form.Field>
          )}
        </section>
      </form>
    </div>
  )
}

function EditChapterForm({ chapterId }: { chapterId: number }) {
  const { data: chapters } = useSuspenseQuery({
    queryKey: CHAPTER_ADMIN_QUERY_KEY,
    queryFn: ({ signal }) => apiClient.getChapterAdminList(signal),
  })
  const chapter = chapters.find((c) => c.ChapterID === chapterId)
  if (!chapter) {
    notFound()
  }
  return <ChapterFormInner chapterId={chapterId} chapter={chapter} />
}

export function ChapterForm({ chapterId }: { chapterId?: number }) {
  if (chapterId != null) {
    return <EditChapterForm chapterId={chapterId} />
  }
  return <ChapterFormInner chapter={undefined} />
}
