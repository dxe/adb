'use client'

import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import toast from 'react-hot-toast'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Button } from '@/components/ui/button'
import { apiClient } from '@/lib/api'
import { SF_BAY_CHAPTER_ID } from '@/lib/constants'

const ERROR_MESSAGE =
  'Sorry, there was an error submitting your form. Please try again.'

const ACTIVISM_INTERESTS = [
  {
    value: 'Animal Care',
    label: 'Animal Care:',
    description: 'Work with and spend time with rescued animals',
  },
  {
    value: 'Community',
    label: 'Community Events:',
    description:
      'Make friends and create connections in the animal rights community',
  },
  {
    value: 'Outreach',
    label: 'Outreach:',
    description: 'Educate the public about animal cruelty',
  },
  {
    value: 'Protest',
    label: 'Demonstrations:',
    description:
      'Challenge corporations and other institutions to make change for animals via non-violent protests or marches',
  },
  {
    value: 'Trainings',
    label: 'Trainings:',
    description:
      'Learn how to talk to people effectively, stay legally safe as an activist, and organize protests',
  },
] as const

const REFERRAL_OUTLET_OPTIONS = [
  { value: 'Social Media', label: 'Social Media' },
  { value: 'Email', label: 'Email' },
  { value: 'Meetup', label: 'Saturday morning meetup' },
  { value: 'In-person Invite', label: 'Someone invited me in person' },
] as const

export function InterestForm() {
  const searchParams = useSearchParams()

  // Form options are derived once from the URL query params, mirroring the
  // `created()` hook in frontend/FormInterest.vue.
  const formOptions = useMemo(() => {
    const chapterIdStr = searchParams.get('chapterId')
    const parsedChapterId =
      chapterIdStr != null ? parseInt(chapterIdStr, 10) : NaN
    const chapterId = !Number.isNaN(parsedChapterId)
      ? parsedChapterId
      : SF_BAY_CHAPTER_ID
    const formName = searchParams.get('name') || 'Interest Form'
    const formTitle =
      searchParams.get('title') ||
      (chapterId === SF_BAY_CHAPTER_ID
        ? 'DxE SF Bay - Get Involved'
        : 'Direct Action Everywhere - Get Involved')
    const formDescription = searchParams.get('description') || ''
    // Note the asymmetry, matching the Vue original: showInterests defaults
    // to true (hidden only when explicitly "false"), while the other show*
    // flags default to false (shown only when explicitly "true").
    const showInterests = searchParams.get('showInterests') !== 'false'
    const showReferralFriends =
      searchParams.get('showReferralFriends') === 'true'
    const showReferralApply = searchParams.get('showReferralApply') === 'true'
    const showReferralOutlet = searchParams.get('showReferralOutlet') === 'true'
    const referralApplyParam = searchParams.get('referralApply') || ''
    const initialReferralApply =
      referralApplyParam !== 'null' ? referralApplyParam : ''

    return {
      chapterId,
      formName,
      formTitle,
      formDescription,
      showInterests,
      showReferralFriends,
      showReferralApply,
      showReferralOutlet,
      initialReferralApply,
    }
  }, [searchParams])

  useEffect(() => {
    document.title = formOptions.formTitle
  }, [formOptions.formTitle])

  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [email, setEmail] = useState('')
  const [phone, setPhone] = useState('')
  const [zip, setZip] = useState('')
  const [referralFriends, setReferralFriends] = useState('')
  const [referralApply, setReferralApply] = useState(
    formOptions.initialReferralApply,
  )
  const [referralOutlet, setReferralOutlet] = useState('')
  const [activismInterests, setActivismInterests] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [submitSuccess, setSubmitSuccess] = useState(false)

  const toggleActivismInterest = (value: string, checked: boolean) => {
    setActivismInterests((prev) =>
      checked ? [...prev, value] : prev.filter((v) => v !== value),
    )
  }

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formEl = e.currentTarget
    if (!formEl.checkValidity()) {
      formEl.reportValidity()
      return
    }

    setSubmitting(true)
    try {
      await apiClient.submitInterestForm({
        chapterId: formOptions.chapterId,
        form: `${formOptions.formName} Form`,
        name: `${firstName} ${lastName}`,
        email,
        zip,
        phone,
        referralFriends,
        referralApply,
        referralOutlet,
        // The legacy form has a "Circle Interest" mode that submits
        // `circleInterests` instead of `activismInterests`, but the Vue
        // template never rendered any UI bound to `circleInterests`, so it
        // was always an empty array. We replicate that exact (dead-code)
        // behavior here rather than porting unused functionality.
        interests:
          formOptions.formName === 'Circle Interest'
            ? ''
            : activismInterests.join(', '),
      })
      toast.success('Submitted!')
      setSubmitSuccess(true)
    } catch {
      toast.error(ERROR_MESSAGE)
    } finally {
      setSubmitting(false)
    }
  }

  if (submitSuccess) {
    return (
      <div className="flex flex-col gap-4">
        <h1 className="text-lg">{formOptions.formTitle}</h1>
        {formOptions.formName === 'Check-in' ? (
          <>
            <p>
              Thank you, {firstName} {lastName}!
            </p>
            <Button
              onClick={() => window.location.reload()}
              disabled={submitting}
            >
              Submit another form
            </Button>
          </>
        ) : (
          <p>Thank you for your submission!</p>
        )}
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-lg">{formOptions.formTitle}</h1>
        {formOptions.formDescription && <p>{formOptions.formDescription}</p>}
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div className="flex flex-col gap-2">
          <Label htmlFor="firstName">First Name</Label>
          <Input
            id="firstName"
            type="text"
            required
            maxLength={35}
            value={firstName}
            onChange={(e) => setFirstName(e.target.value.trim())}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="lastName">Last Name</Label>
          <Input
            id="lastName"
            type="text"
            required
            maxLength={35}
            value={lastName}
            onChange={(e) => setLastName(e.target.value.trim())}
          />
        </div>

        <div className="flex flex-col gap-2 sm:col-span-2">
          <Label htmlFor="email">Email</Label>
          <Input
            id="email"
            type="email"
            required
            maxLength={80}
            value={email}
            onChange={(e) => setEmail(e.target.value.trim())}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="phone">Phone Number</Label>
          <Input
            id="phone"
            type="text"
            required
            maxLength={20}
            value={phone}
            onChange={(e) => setPhone(e.target.value.trim())}
          />
        </div>

        <div className="flex flex-col gap-2">
          <Label htmlFor="zip">Zip Code</Label>
          <Input
            id="zip"
            type="text"
            required
            maxLength={5}
            value={zip}
            onChange={(e) => setZip(e.target.value.trim())}
          />
        </div>
      </div>

      {formOptions.showInterests && (
        <div className="flex flex-col gap-3">
          <p>What are your activism interests, if any?</p>
          <div className="flex flex-col gap-3">
            {ACTIVISM_INTERESTS.map((interest) => (
              <Label
                key={interest.value}
                className="flex flex-row items-start gap-2 font-normal"
              >
                <Checkbox
                  className="mt-1"
                  checked={activismInterests.includes(interest.value)}
                  onCheckedChange={(checked) =>
                    toggleActivismInterest(interest.value, Boolean(checked))
                  }
                />
                <span>
                  <strong>{interest.label}</strong>{' '}
                  <span className="text-sm">{interest.description}</span>
                </span>
              </Label>
            ))}
          </div>
        </div>
      )}

      {formOptions.showReferralFriends && (
        <div className="flex flex-col gap-2">
          <Label htmlFor="referralFriends">
            List any existing DxE activists who you are close friends with
          </Label>
          <Input
            id="referralFriends"
            type="text"
            maxLength={200}
            value={referralFriends}
            onChange={(e) => setReferralFriends(e.target.value.trim())}
          />
        </div>
      )}

      {formOptions.showReferralApply && (
        <div className="flex flex-col gap-2">
          <Label htmlFor="referralApply">Who encouraged you to sign up?</Label>
          <Input
            id="referralApply"
            type="text"
            maxLength={200}
            value={referralApply}
            onChange={(e) => setReferralApply(e.target.value.trim())}
          />
        </div>
      )}

      {formOptions.showReferralOutlet && (
        <div className="flex flex-col gap-3">
          <p>
            Where did you hear about this opportunity to get involved in DxE?
          </p>
          <RadioGroup value={referralOutlet} onValueChange={setReferralOutlet}>
            {REFERRAL_OUTLET_OPTIONS.map((option) => (
              <Label
                key={option.value}
                className="flex flex-row items-center gap-2 font-normal"
              >
                <RadioGroupItem value={option.value} />
                {option.label}
              </Label>
            ))}
          </RadioGroup>
        </div>
      )}

      <Button type="submit" disabled={submitting} className="w-fit">
        Submit
      </Button>
    </form>
  )
}
