'use client'

import { useEffect, useMemo, useState } from 'react'
import { useSearchParams } from 'next/navigation'
import { useForm } from '@tanstack/react-form'
import { useMutation } from '@tanstack/react-query'
import { z } from 'zod'
import toast from 'react-hot-toast'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Checkbox } from '@/components/ui/checkbox'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Button } from '@/components/ui/button'
import { FieldError } from '@/app/(authed)/events/field-error'
import { apiClient, InterestFormPayload } from '@/lib/api'
import { parseInterestFormOptions } from './form-options'

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

// Unlike the legacy form (native browser validation bubbles), validation is
// via zod with the form set to noValidate; `required` attrs are a11y-only.
const formSchema = z.object({
  firstName: z.string().trim().min(1, 'First Name is required.').max(35),
  lastName: z.string().trim().min(1, 'Last Name is required.').max(35),
  email: z
    .string()
    .trim()
    .min(1, 'Email is required.')
    .max(80)
    .email('Enter a valid email.'),
  phone: z.string().trim().min(1, 'Phone Number is required.').max(20),
  zip: z.string().trim().min(1, 'Zip Code is required.').max(5),
  referralFriends: z.string().trim().max(200),
  referralApply: z.string().trim().max(200),
  referralOutlet: z.string(),
  activismInterests: z.array(z.string()),
})

type FormValues = z.infer<typeof formSchema>

export function InterestForm() {
  const searchParams = useSearchParams()
  const formOptions = useMemo(
    () => parseInterestFormOptions(new URLSearchParams(searchParams)),
    [searchParams],
  )

  useEffect(() => {
    document.title = formOptions.formTitle
  }, [formOptions.formTitle])

  // Set to the submitted name on success; doubles as the success-screen toggle.
  const [submittedName, setSubmittedName] = useState<string | null>(null)

  const mutation = useMutation({
    mutationFn: (payload: InterestFormPayload) =>
      apiClient.submitInterestForm(payload),
    onSuccess: (_data, payload) => {
      toast.success('Submitted!')
      setSubmittedName(payload.name)
    },
    onError: () => {
      toast.error(
        'Sorry, there was an error submitting your form. Please try again.',
      )
    },
  })

  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      phone: '',
      zip: '',
      referralFriends: '',
      referralApply: formOptions.initialReferralApply,
      referralOutlet: '',
      activismInterests: [],
    } as FormValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmitInvalid: ({ formApi }) => {
      const firstError = Object.values(formApi.state.fieldMeta)
        .flatMap((meta) => meta?.errors ?? [])
        .at(0)
      if (firstError) toast.error(firstError.message)
    },
    onSubmit: ({ value }) => {
      // safeParse applies the schema's trim transforms, which the form's
      // onSubmit validator does not.
      const parsed = formSchema.safeParse(value)
      if (!parsed.success) return
      const values = parsed.data
      mutation.mutate({
        chapterId: formOptions.chapterId,
        form: `${formOptions.formName} Form`,
        name: `${values.firstName} ${values.lastName}`,
        email: values.email,
        zip: values.zip,
        phone: values.phone,
        referralFriends: values.referralFriends,
        referralApply: values.referralApply,
        referralOutlet: values.referralOutlet,
        // Vue parity: "Circle Interest" mode always submitted empty
        // interests (its circleInterests UI was never rendered).
        interests:
          formOptions.formName === 'Circle Interest'
            ? ''
            : values.activismInterests.join(', '),
      })
    },
  })

  const header = (
    <div className="flex flex-col gap-1">
      <h1 className="text-lg">{formOptions.formTitle}</h1>
      {formOptions.formDescription && <p>{formOptions.formDescription}</p>}
    </div>
  )

  if (submittedName !== null) {
    return (
      <div className="flex flex-col gap-4">
        {header}
        {formOptions.formName === 'Check-in' ? (
          <>
            <p>Thank you, {submittedName}!</p>
            <Button
              onClick={() => window.location.reload()}
              disabled={mutation.isPending}
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
    <form
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
      noValidate
      className="flex flex-col gap-6"
    >
      {header}

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <form.Field name="firstName">
          {(field) => (
            <div className="flex flex-col gap-2">
              <Label htmlFor="firstName">First Name</Label>
              <Input
                id="firstName"
                type="text"
                required
                maxLength={35}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
              <FieldError message={field.state.meta.errors[0]?.message} />
            </div>
          )}
        </form.Field>

        <form.Field name="lastName">
          {(field) => (
            <div className="flex flex-col gap-2">
              <Label htmlFor="lastName">Last Name</Label>
              <Input
                id="lastName"
                type="text"
                required
                maxLength={35}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
              <FieldError message={field.state.meta.errors[0]?.message} />
            </div>
          )}
        </form.Field>

        <form.Field name="email">
          {(field) => (
            <div className="flex flex-col gap-2 sm:col-span-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                required
                maxLength={80}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
              <FieldError message={field.state.meta.errors[0]?.message} />
            </div>
          )}
        </form.Field>

        <form.Field name="phone">
          {(field) => (
            <div className="flex flex-col gap-2">
              <Label htmlFor="phone">Phone Number</Label>
              <Input
                id="phone"
                type="text"
                required
                maxLength={20}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
              <FieldError message={field.state.meta.errors[0]?.message} />
            </div>
          )}
        </form.Field>

        <form.Field name="zip">
          {(field) => (
            <div className="flex flex-col gap-2">
              <Label htmlFor="zip">Zip Code</Label>
              <Input
                id="zip"
                type="text"
                required
                maxLength={5}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
              <FieldError message={field.state.meta.errors[0]?.message} />
            </div>
          )}
        </form.Field>
      </div>

      {formOptions.showInterests && (
        <form.Field name="activismInterests">
          {(field) => (
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
                      checked={field.state.value.includes(interest.value)}
                      onCheckedChange={(checked) =>
                        field.handleChange(
                          checked
                            ? [...field.state.value, interest.value]
                            : field.state.value.filter(
                                (v) => v !== interest.value,
                              ),
                        )
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
        </form.Field>
      )}

      {formOptions.showReferralFriends && (
        <form.Field name="referralFriends">
          {(field) => (
            <div className="flex flex-col gap-2">
              <Label htmlFor="referralFriends">
                List any existing DxE activists who you are close friends with
              </Label>
              <Input
                id="referralFriends"
                type="text"
                maxLength={200}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
            </div>
          )}
        </form.Field>
      )}

      {formOptions.showReferralApply && (
        <form.Field name="referralApply">
          {(field) => (
            <div className="flex flex-col gap-2">
              <Label htmlFor="referralApply">
                Who encouraged you to sign up?
              </Label>
              <Input
                id="referralApply"
                type="text"
                maxLength={200}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
            </div>
          )}
        </form.Field>
      )}

      {formOptions.showReferralOutlet && (
        <form.Field name="referralOutlet">
          {(field) => (
            <div className="flex flex-col gap-3">
              <p>
                Where did you hear about this opportunity to get involved in
                DxE?
              </p>
              <RadioGroup
                value={field.state.value}
                onValueChange={field.handleChange}
              >
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
        </form.Field>
      )}

      <Button type="submit" disabled={mutation.isPending} className="w-fit">
        Submit
      </Button>
    </form>
  )
}
