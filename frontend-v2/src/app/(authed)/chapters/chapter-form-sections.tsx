'use client'

import { Plus, Trash2 } from 'lucide-react'
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
import { COUNTRIES } from './countries'
import { REGIONS } from './chapter-utils'
import type { ChapterFormApi } from './useChapterForm'

const Section = ({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) => (
  <section className="space-y-4">
    <h2 className="text-sm font-semibold text-primary">{title}</h2>
    {children}
  </section>
)

const FieldError = ({ message }: { message: string | undefined }) => {
  if (!message) return null
  return <p className="text-sm text-destructive">{message}</p>
}

export const BasicInfoSection = ({
  form,
  isEditing,
}: {
  form: ChapterFormApi
  isEditing: boolean
}) => (
  <Section title="Basic Info">
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
            <FieldError message={field.state.meta.errors[0]?.message} />
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
  </Section>
)

const COORDINATE_FIELDS = [
  { name: 'lat', label: 'Lat', range: 90, placeholder: '00.000000' },
  { name: 'lng', label: 'Lng', range: 180, placeholder: '000.000000' },
] as const

export const LocationSection = ({ form }: { form: ChapterFormApi }) => (
  <Section title="Location">
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
            <FieldError message={field.state.meta.errors[0]?.message} />
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
            <FieldError message={field.state.meta.errors[0]?.message} />
          </div>
        )}
      </form.Field>

      {COORDINATE_FIELDS.map(({ name, label, range, placeholder }) => (
        <form.Field key={name} name={name}>
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor={`chapter-${name}`}>{label}</Label>
              <Input
                id={`chapter-${name}`}
                type="number"
                inputMode="decimal"
                step="any"
                min={-range}
                max={range}
                placeholder={placeholder}
                value={field.state.value ?? ''}
                onChange={(e) => {
                  const n = e.target.valueAsNumber
                  field.handleChange(Number.isNaN(n) ? undefined : n)
                }}
                onBlur={field.handleBlur}
              />
              <FieldError message={field.state.meta.errors[0]?.message} />
            </div>
          )}
        </form.Field>
      ))}
    </div>
  </Section>
)

const SOCIAL_FIELDS = [
  { name: 'fbUrl', label: 'Facebook' },
  { name: 'twitterUrl', label: 'Twitter' },
  { name: 'instaUrl', label: 'Instagram' },
] as const

export const SocialLinksSection = ({ form }: { form: ChapterFormApi }) => (
  <Section title="Social Links">
    <div className="grid gap-4 md:grid-cols-2">
      {SOCIAL_FIELDS.map(({ name, label }) => (
        <form.Field key={name} name={name}>
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor={`chapter-${name}`}>{label}</Label>
              <Input
                id={`chapter-${name}`}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={100}
              />
            </div>
          )}
        </form.Field>
      ))}

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
            <FieldError message={field.state.meta.errors[0]?.message} />
          </div>
        )}
      </form.Field>
    </div>
  </Section>
)

const DATE_FIELDS = [
  { name: 'lastContact', label: 'Last Contact' },
  { name: 'lastAction', label: 'Last Action' },
] as const

export const DatesSection = ({ form }: { form: ChapterFormApi }) => (
  <Section title="Dates">
    <div className="grid gap-4 md:grid-cols-2">
      {DATE_FIELDS.map(({ name, label }) => (
        <form.Field key={name} name={name}>
          {(field) => (
            <div className="space-y-1">
              <Label>{label}</Label>
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
      ))}
    </div>
  </Section>
)

const ORGANIZER_FIELDS = [
  { key: 'Name', type: 'text' },
  { key: 'Email', type: 'email' },
  { key: 'Phone', type: 'text' },
  { key: 'Facebook', type: 'text', maxLength: 100 },
  { key: 'Instagram', type: 'text', maxLength: 100 },
  { key: 'Twitter', type: 'text', maxLength: 100 },
  { key: 'Website', type: 'text', maxLength: 100 },
] as const

const EMPTY_ORGANIZER = {
  Name: '',
  Email: '',
  Phone: '',
  Facebook: '',
  Instagram: '',
  Twitter: '',
  Website: '',
}

export const OrganizersSection = ({
  form,
  isEditing,
}: {
  form: ChapterFormApi
  isEditing: boolean
}) => (
  <Section title="Organizers">
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
              <div key={index} className="space-y-3 rounded-md border p-4">
                <div className="flex items-start justify-between gap-2">
                  <div className="grid flex-1 gap-3 md:grid-cols-3">
                    {ORGANIZER_FIELDS.map(({ key, type, ...rest }) => (
                      <form.Field
                        key={key}
                        name={`organizers[${index}].${key}`}
                      >
                        {(field) => (
                          <div className="space-y-1">
                            <Label>{key}</Label>
                            <Input
                              type={type}
                              value={field.state.value}
                              onChange={(e) =>
                                field.handleChange(e.target.value)
                              }
                              onBlur={field.handleBlur}
                              placeholder={key}
                              {...rest}
                            />
                            <FieldError
                              message={field.state.meta.errors[0]?.message}
                            />
                          </div>
                        )}
                      </form.Field>
                    ))}
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
              onClick={() => arrayField.pushValue({ ...EMPTY_ORGANIZER })}
            >
              <Plus className="h-4 w-4" />
              Add new organizer
            </Button>
          </div>
        )}
      </form.Field>
    )}
  </Section>
)
