import { z } from 'zod'
import { ChapterAdmin } from '@/lib/api'
import { COUNTRIES } from './countries'
import { formatDateYmd, parseDateYmd } from './chapter-utils'

// Field names match the API's PascalCase (`Name`, `Email`, ...) 1:1 so the
// organizer payload can be sent without a case-mapping step.
export const organizerFormSchema = z.object({
  Name: z.string().trim().min(1, 'Organizer name is required'),
  Email: z.string().trim(),
  Phone: z.string().trim(),
  Facebook: z.string().trim(),
  Instagram: z.string().trim(),
  Twitter: z.string().trim(),
  Website: z.string().trim(),
})

export const chapterFormSchema = z.object({
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
    .pipe(z.email('Enter a valid email').or(z.literal(''))),
  lastContact: z.date().nullable(),
  lastAction: z.date().nullable(),
  organizers: z.array(organizerFormSchema),
})

// Stricter checks that only apply at submit time.
export const chapterFormSubmitSchema = chapterFormSchema.extend({
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

export type ChapterFormValues = z.infer<typeof chapterFormSchema>
export type ChapterFormSubmitValues = z.infer<typeof chapterFormSubmitSchema>

export function toInitialValues(
  chapter: ChapterAdmin | undefined,
): ChapterFormValues {
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

// Spreads the original chapter first so fields the form never displays
// round-trip unchanged instead of being wiped on save.
// TODO: the only such field is EmailToken, which is unused and should be removed.
export function buildChapterPayload(
  chapter: ChapterAdmin | undefined,
  values: ChapterFormSubmitValues,
): Partial<ChapterAdmin> {
  return {
    ...chapter,
    Name: values.name,
    Flag: COUNTRIES.find((c) => c.code === values.country)?.flag ?? '',
    Mentor: values.mentor,
    Notes: values.notes,
    Region: values.region,
    Country: values.country,
    Lat: values.lat,
    Lng: values.lng,
    FbURL: values.fbUrl,
    TwitterURL: values.twitterUrl,
    InstaURL: values.instaUrl,
    Email: values.email,
    LastContact: values.lastContact ? formatDateYmd(values.lastContact) : '',
    LastAction: values.lastAction ? formatDateYmd(values.lastAction) : '',
    Organizers: values.organizers,
  }
}
