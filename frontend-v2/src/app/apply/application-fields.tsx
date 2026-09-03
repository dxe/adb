'use client'

import { useForm } from '@tanstack/react-form'
import { Loader2 } from 'lucide-react'
import toast from 'react-hot-toast'
import { z } from 'zod'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { ApplicationFormPayload } from '@/lib/api'

// Key order and message text match the legacy Vue form's validation order;
// the form is noValidate, so inputs' `required` attrs are a11y-only.
const applicationSchema = z.object({
  firstName: z.string().trim().min(1, 'First Name is required.'),
  lastName: z.string().trim().min(1, 'Last Name is required.'),
  pronouns: z.string().trim(),
  email: z.string().trim().min(1, 'Email is required.'),
  address: z.string().trim().min(1, 'Street Address is required.'),
  city: z.string().trim().min(1, 'City is required.'),
  zip: z.string().trim().min(1, 'Zip Code is required.'),
  phone: z.string().trim().min(1, 'Phone is required.'),
  birthday: z.string().trim().min(1, 'Birthday is required.'),
  conduct: z.literal(true, {
    message: 'You must agree to the code of conduct.',
  }),
  mission: z.literal(true, {
    message: 'Please must agree with our mission & values.',
  }),
  consent: z.literal(true, {
    message: 'Please must agree to watch a video & take a quiz on consent.',
  }),
  applicationType: z.enum(['organizer', 'chapter-member'], {
    message:
      'You must choose whether or not you are interested in becoming an organizer.',
  }),
  referral: z.string().trim(),
  language: z.string().trim(),
  accessibility: z.string().trim(),
})

type ApplicationType = z.infer<typeof applicationSchema>['applicationType']

interface ApplicationFormValues {
  firstName: string
  lastName: string
  pronouns: string
  email: string
  address: string
  city: string
  zip: string
  phone: string
  birthday: string
  conduct: boolean
  mission: boolean
  consent: boolean
  applicationType: ApplicationType | null
  referral: string
  language: string
  accessibility: string
}

const defaultValues: ApplicationFormValues = {
  firstName: '',
  lastName: '',
  pronouns: '',
  email: '',
  address: '',
  city: '',
  zip: '',
  phone: '',
  birthday: '',
  conduct: false,
  mission: false,
  consent: false,
  applicationType: null,
  referral: '',
  language: '',
  accessibility: '',
}

function Field({
  id,
  label,
  message,
  error,
  className,
  children,
}: {
  id: string
  label: string
  message?: string
  error?: string
  className?: string
  children: React.ReactNode
}) {
  return (
    <div className={cn('flex flex-col gap-1', className)}>
      <Label htmlFor={id}>{label}</Label>
      {children}
      {error && <p className="text-xs text-destructive">{error}</p>}
      {message && <p className="text-xs text-muted-foreground">{message}</p>}
    </div>
  )
}

function AgreementToggle({
  id,
  checked,
  onCheckedChange,
  error,
  children,
}: {
  id: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  error?: string
  children: React.ReactNode
}) {
  return (
    <div className="flex flex-col gap-2">
      <p id={`${id}-statement`} className="text-sm">
        {children}
      </p>
      <label
        htmlFor={id}
        className="flex w-fit cursor-pointer items-center gap-2 rounded-md border p-2 text-sm"
      >
        <Checkbox
          id={id}
          aria-describedby={`${id}-statement`}
          checked={checked}
          onCheckedChange={(value) => onCheckedChange(Boolean(value))}
        />
        Yes, I agree.
      </label>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}

/** The application form fields, validated against applicationSchema and submitted to POST /apply. */
export function ApplicationFields({
  onSubmit,
  isSubmitting,
}: {
  onSubmit: (payload: ApplicationFormPayload) => void
  isSubmitting: boolean
}) {
  const form = useForm({
    defaultValues,
    validators: {
      onSubmit: applicationSchema,
    },
    onSubmit: async ({ value }) => {
      const parsed = applicationSchema.safeParse(value)
      if (!parsed.success) {
        toast.error('Please fix the highlighted errors.')
        return
      }

      const values = parsed.data
      onSubmit({
        name: `${values.firstName} ${values.lastName}`,
        firstName: values.firstName,
        lastName: values.lastName,
        pronouns: values.pronouns,
        email: values.email,
        address: values.address,
        city: values.city,
        zip: values.zip,
        phone: values.phone,
        birthday: values.birthday,
        referral: values.referral,
        language: values.language,
        accessibility: values.accessibility,
        applicationType: values.applicationType,
      })
    },
  })

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
      noValidate
      className="flex flex-col gap-6"
    >
      <h2 className="text-xl font-semibold">Take direct action for animals</h2>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <form.Field name="firstName">
          {(field) => (
            <Field
              id="firstName"
              label="First Name"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="firstName"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={35}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="lastName">
          {(field) => (
            <Field
              id="lastName"
              label="Last Name"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="lastName"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={35}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="pronouns">
          {(field) => (
            <Field
              id="pronouns"
              label="Pronouns (optional)"
              className="md:col-span-2"
            >
              <Input
                id="pronouns"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={20}
                placeholder="she/her, he/him, they/them, etc."
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="mission">
          {(field) => (
            <AgreementToggle
              id="mission"
              checked={field.state.value}
              onCheckedChange={(checked) => field.handleChange(checked)}
              error={field.state.meta.errors[0]?.message}
            >
              I support DxE&rsquo;s{' '}
              <a
                href="https://www.directactioneverywhere.com/core-values"
                target="_blank"
                rel="noreferrer"
                className="text-primary underline"
              >
                mission and values
              </a>
              .
            </AgreementToggle>
          )}
        </form.Field>

        <form.Field name="conduct">
          {(field) => (
            <AgreementToggle
              id="conduct"
              checked={field.state.value}
              onCheckedChange={(checked) => field.handleChange(checked)}
              error={field.state.meta.errors[0]?.message}
            >
              I will uphold DxE&rsquo;s{' '}
              <a
                href="https://dxe.io/conduct"
                target="_blank"
                rel="noreferrer"
                className="text-primary underline"
              >
                code of conduct
              </a>
              .
            </AgreementToggle>
          )}
        </form.Field>

        <form.Field name="consent">
          {(field) => (
            <AgreementToggle
              id="consent"
              checked={field.state.value}
              onCheckedChange={(checked) => field.handleChange(checked)}
              error={field.state.meta.errors[0]?.message}
            >
              I agree to watch a video and{' '}
              <a
                href="https://dxe.io/refresherquiz"
                target="_blank"
                rel="noreferrer"
                className="text-primary underline"
              >
                take a quiz
              </a>{' '}
              on consent.
            </AgreementToggle>
          )}
        </form.Field>

        <div className="md:col-span-2">
          <h2 className="mt-3 text-xl font-semibold">Contact Info</h2>
        </div>

        <form.Field name="email">
          {(field) => (
            <Field
              id="email"
              label="Email"
              className="md:col-span-2"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="email"
                type="email"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={60}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="address">
          {(field) => (
            <Field
              id="address"
              label="Street Address"
              className="md:col-span-2"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="address"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={60}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="city">
          {(field) => (
            <Field
              id="city"
              label="City"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="city"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={90}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="zip">
          {(field) => (
            <Field
              id="zip"
              label="Zip Code"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="zip"
                type="text"
                inputMode="numeric"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={5}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="phone">
          {(field) => (
            <Field
              id="phone"
              label="Phone"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="phone"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
                maxLength={20}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="birthday">
          {(field) => (
            <Field
              id="birthday"
              label="Birthday"
              error={field.state.meta.errors[0]?.message}
            >
              <Input
                id="birthday"
                type="date"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                required
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="referral">
          {(field) => (
            <Field
              id="referral"
              label="Who encouraged you to apply? (optional)"
            >
              <Input
                id="referral"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={100}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="language">
          {(field) => (
            <Field
              id="language"
              label="Primary language (optional)"
              message="We try to create materials and events in your primary language when possible."
            >
              <Input
                id="language"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={40}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="accessibility">
          {(field) => (
            <Field
              id="accessibility"
              label="Accessibility needs (optional)"
              message="We do our best to accommodate our events to your needs."
              className="md:col-span-2"
            >
              <Input
                id="accessibility"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={300}
              />
            </Field>
          )}
        </form.Field>

        <form.Field name="applicationType">
          {(field) => (
            <div className="flex flex-col gap-2 md:col-span-2">
              <Label>
                Are you interested in further leveling up your activism by
                becoming an Organizer?
              </Label>
              <p className="text-xs text-muted-foreground">
                Organizers take ownership over achieving the chapter&rsquo;s
                objectives and make our chapter function by organizing community
                events, editing videos, leading protests, raising money, writing
                press releases, and more. By becoming an organizer, you become a
                primary driver of the chapter&rsquo;s objectives. They volunteer
                for 2-5 hours per week.
              </p>
              <div className="flex gap-3">
                <Button
                  type="button"
                  variant={
                    field.state.value === 'organizer' ? 'default' : 'outline'
                  }
                  onClick={() => field.handleChange('organizer')}
                >
                  Yes
                </Button>
                <Button
                  type="button"
                  variant={
                    field.state.value === 'chapter-member'
                      ? 'default'
                      : 'outline'
                  }
                  onClick={() => field.handleChange('chapter-member')}
                >
                  No (or not sure)
                </Button>
              </div>
              {field.state.meta.errors[0] && (
                <p className="text-xs text-destructive">
                  {field.state.meta.errors[0]?.message}
                </p>
              )}
            </div>
          )}
        </form.Field>
      </div>

      <Button type="submit" disabled={isSubmitting} className="mt-3 w-fit">
        {isSubmitting ? (
          <>
            <Loader2 className="h-4 w-4 animate-spin" />
            Submitting...
          </>
        ) : (
          'Submit'
        )}
      </Button>
    </form>
  )
}
