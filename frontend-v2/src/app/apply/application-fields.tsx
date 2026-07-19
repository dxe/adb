'use client'

import { FormEvent, useState } from 'react'
import { Loader2 } from 'lucide-react'
import toast from 'react-hot-toast'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { ApplicationFormPayload } from '@/lib/api'

type ApplicationType = 'organizer' | 'chapter-member' | ''

function Field({
  id,
  label,
  message,
  className,
  children,
}: {
  id: string
  label: string
  message?: string
  className?: string
  children: React.ReactNode
}) {
  return (
    <div className={cn('flex flex-col gap-1', className)}>
      <Label htmlFor={id}>{label}</Label>
      {children}
      {message && <p className="text-xs text-muted-foreground">{message}</p>}
    </div>
  )
}

function AgreementToggle({
  id,
  checked,
  onCheckedChange,
  children,
}: {
  id: string
  checked: boolean
  onCheckedChange: (checked: boolean) => void
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
    </div>
  )
}

/**
 * The actual application fields, submitted to POST /apply. Mirrors the
 * `showForm && !submitSuccess` section of frontend/FormApply.vue, including
 * the field-required validation (native HTML5 `required` here, matching the
 * legacy checkHtml5Validity() checks) and the ordered business-rule checks
 * (conduct -> mission -> consent -> applicationType) surfaced as toasts,
 * which fire after native validation passes.
 */
export function ApplicationFields({
  onSubmit,
  isSubmitting,
}: {
  onSubmit: (payload: ApplicationFormPayload) => void
  isSubmitting: boolean
}) {
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [pronouns, setPronouns] = useState('')
  const [mission, setMission] = useState(false)
  const [conduct, setConduct] = useState(false)
  const [consent, setConsent] = useState(false)
  const [email, setEmail] = useState('')
  const [address, setAddress] = useState('')
  const [city, setCity] = useState('')
  const [zip, setZip] = useState('')
  const [phone, setPhone] = useState('')
  const [birthday, setBirthday] = useState('')
  const [referral, setReferral] = useState('')
  const [language, setLanguage] = useState('')
  const [accessibility, setAccessibility] = useState('')
  const [applicationType, setApplicationType] = useState<ApplicationType>('')

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()

    // Trim at submit time rather than on every keystroke (unlike the legacy
    // Vue form's `v-model.trim`, which trims per-keystroke and can clip the
    // space between words while typing a multi-word address/city).
    const trimmed = {
      firstName: firstName.trim(),
      lastName: lastName.trim(),
      pronouns: pronouns.trim(),
      email: email.trim(),
      address: address.trim(),
      city: city.trim(),
      zip: zip.trim(),
      phone: phone.trim(),
      birthday: birthday.trim(),
      referral: referral.trim(),
      language: language.trim(),
      accessibility: accessibility.trim(),
    }

    // Native HTML5 `required` blocks empty fields before this handler fires,
    // but a whitespace-only value passes `required` and would submit as an
    // empty string after trimming. Re-check the trimmed values here, in field
    // order, before the business-rule checks (mirroring the legacy flow where
    // field validation ran first).
    const requiredFields: Array<[keyof typeof trimmed, string]> = [
      ['firstName', 'First Name'],
      ['lastName', 'Last Name'],
      ['email', 'Email'],
      ['address', 'Street Address'],
      ['city', 'City'],
      ['zip', 'Zip Code'],
      ['phone', 'Phone'],
      ['birthday', 'Birthday'],
    ]
    for (const [field, label] of requiredFields) {
      if (!trimmed[field]) {
        toast.error(`${label} is required.`)
        return
      }
    }

    // Business-rule checks, same order as the legacy Vue submitForm().
    if (!conduct) {
      toast.error('You must agree to the code of conduct.')
      return
    }
    if (!mission) {
      toast.error('Please must agree with our mission & values.')
      return
    }
    if (!consent) {
      toast.error(
        'Please must agree to watch a video & take a quiz on consent.',
      )
      return
    }
    if (!applicationType) {
      toast.error(
        'You must choose whether or not you are interested in becoming an organizer.',
      )
      return
    }

    onSubmit({
      name: `${trimmed.firstName} ${trimmed.lastName}`,
      ...trimmed,
      applicationType,
    })
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <h2 className="text-xl font-semibold">Take direct action for animals</h2>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <Field id="firstName" label="First Name">
          <Input
            id="firstName"
            type="text"
            value={firstName}
            onChange={(e) => setFirstName(e.target.value)}
            required
            maxLength={35}
          />
        </Field>

        <Field id="lastName" label="Last Name">
          <Input
            id="lastName"
            type="text"
            value={lastName}
            onChange={(e) => setLastName(e.target.value)}
            required
            maxLength={35}
          />
        </Field>

        <Field
          id="pronouns"
          label="Pronouns (optional)"
          className="md:col-span-2"
        >
          <Input
            id="pronouns"
            type="text"
            value={pronouns}
            onChange={(e) => setPronouns(e.target.value)}
            maxLength={20}
            placeholder="she/her, he/him, they/them, etc."
          />
        </Field>

        <AgreementToggle
          id="mission"
          checked={mission}
          onCheckedChange={setMission}
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

        <AgreementToggle
          id="conduct"
          checked={conduct}
          onCheckedChange={setConduct}
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

        <AgreementToggle
          id="consent"
          checked={consent}
          onCheckedChange={setConsent}
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

        <div className="md:col-span-2">
          <h2 className="mt-3 text-xl font-semibold">Contact Info</h2>
        </div>

        <Field id="email" label="Email" className="md:col-span-2">
          <Input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            maxLength={60}
          />
        </Field>

        <Field id="address" label="Street Address" className="md:col-span-2">
          <Input
            id="address"
            type="text"
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            required
            maxLength={60}
          />
        </Field>

        <Field id="city" label="City">
          <Input
            id="city"
            type="text"
            value={city}
            onChange={(e) => setCity(e.target.value)}
            required
            maxLength={90}
          />
        </Field>

        <Field id="zip" label="Zip Code">
          <Input
            id="zip"
            type="text"
            inputMode="numeric"
            value={zip}
            onChange={(e) => setZip(e.target.value)}
            required
            maxLength={5}
          />
        </Field>

        <Field id="phone" label="Phone">
          <Input
            id="phone"
            type="text"
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            required
            maxLength={20}
          />
        </Field>

        <Field id="birthday" label="Birthday">
          <Input
            id="birthday"
            type="date"
            value={birthday}
            onChange={(e) => setBirthday(e.target.value)}
            required
          />
        </Field>

        <Field id="referral" label="Who encouraged you to apply? (optional)">
          <Input
            id="referral"
            type="text"
            value={referral}
            onChange={(e) => setReferral(e.target.value)}
            maxLength={100}
          />
        </Field>

        <Field
          id="language"
          label="Primary language (optional)"
          message="We try to create materials and events in your primary language when possible."
        >
          <Input
            id="language"
            type="text"
            value={language}
            onChange={(e) => setLanguage(e.target.value)}
            maxLength={40}
          />
        </Field>

        <Field
          id="accessibility"
          label="Accessibility needs (optional)"
          message="We do our best to accommodate our events to your needs."
          className="md:col-span-2"
        >
          <Input
            id="accessibility"
            type="text"
            value={accessibility}
            onChange={(e) => setAccessibility(e.target.value)}
            maxLength={300}
          />
        </Field>

        <div className="flex flex-col gap-2 md:col-span-2">
          <Label>
            Are you interested in further leveling up your activism by becoming
            an Organizer?
          </Label>
          <p className="text-xs text-muted-foreground">
            Organizers take ownership over achieving the chapter&rsquo;s
            objectives and make our chapter function by organizing community
            events, editing videos, leading protests, raising money, writing
            press releases, and more. By becoming an organizer, you become a
            primary driver of the chapter&rsquo;s objectives. They volunteer for
            2-5 hours per week.
          </p>
          <div className="flex gap-3">
            <Button
              type="button"
              variant={applicationType === 'organizer' ? 'default' : 'outline'}
              onClick={() => setApplicationType('organizer')}
            >
              Yes
            </Button>
            <Button
              type="button"
              variant={
                applicationType === 'chapter-member' ? 'default' : 'outline'
              }
              onClick={() => setApplicationType('chapter-member')}
            >
              No (or not sure)
            </Button>
          </div>
        </div>
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
