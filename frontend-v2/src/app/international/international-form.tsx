'use client'

import { useState } from 'react'
import { useForm } from '@tanstack/react-form'
import { useMutation, useQuery } from '@tanstack/react-query'
import { z } from 'zod'
import toast from 'react-hot-toast'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { API_PATH, apiClient } from '@/lib/api'
import { CityAutocomplete, CityValue } from './city-autocomplete'

const citySchema = z.object({
  city: z.string(),
  state: z.string(),
  country: z.string(),
  lat: z.number(),
  lng: z.number(),
})

const formSchema = z.object({
  firstName: z.string().trim().min(1, 'First name is required').max(35),
  lastName: z.string().trim().min(1, 'Last name is required').max(35),
  email: z
    .string()
    .trim()
    .min(1, 'Email is required')
    .max(60)
    .email('Enter a valid email'),
  phone: z.string().trim().min(1, 'Phone is required').max(35),
  location: citySchema.nullable(),
  interest: z.enum(['participate', 'organize']).nullable(),
  termsAgreed: z.boolean(),
  involvement: z.string().max(500),
})

// Requires the location/interest/terms choices at submit time.
const formSubmitSchema = formSchema.extend({
  location: z.custom<CityValue>((v) => v != null, {
    message: 'Please choose your city from the dropdown list.',
  }),
  interest: z.enum(['participate', 'organize'], {
    message: "Please choose whether you'd like to participate or organize.",
  }),
  termsAgreed: z.literal(true, { message: 'You must agree to the terms.' }),
})

type FormValues = z.infer<typeof formSchema>
type SubmitValues = z.infer<typeof formSubmitSchema>

const defaultValues: FormValues = {
  firstName: '',
  lastName: '',
  email: '',
  phone: '',
  location: null,
  interest: null,
  termsAgreed: false,
  involvement: '',
}

export function InternationalForm() {
  const [submitSuccess, setSubmitSuccess] = useState(false)

  // On failure the form still works; the city field just has no suggestions.
  const { data: placesApiKey } = useQuery({
    queryKey: [API_PATH.PLACES_API_KEY],
    queryFn: ({ signal }) => apiClient.getPlacesApiKey(signal),
    staleTime: Infinity,
    retry: 1,
  })

  const mutation = useMutation({
    mutationFn: (values: SubmitValues) =>
      apiClient.submitInternationalForm({
        firstName: values.firstName,
        lastName: values.lastName,
        email: values.email,
        phone: values.phone,
        interest: values.interest,
        involvement: values.involvement,
        city: values.location.city,
        state: values.location.state,
        country: values.location.country,
        lat: values.location.lat,
        lng: values.location.lng,
      }),
    onSuccess: () => {
      toast.success('Submitted!')
      setSubmitSuccess(true)
    },
    onError: () => {
      // TODO: include error detail from backend
      toast.error(
        'Sorry, there was an error submitting your form. Please try again.',
      )
    },
  })

  const form = useForm({
    defaultValues,
    validators: {
      onSubmit: formSubmitSchema,
    },
    onSubmit: async ({ value }) => {
      const parsed = formSubmitSchema.safeParse(value)
      if (!parsed.success) return
      await mutation.mutateAsync(parsed.data)
    },
  })

  const isSubmitting = mutation.isPending

  if (submitSuccess) {
    return (
      <div className="flex flex-col gap-2">
        <h2 className="text-xl font-semibold">Thank you!</h2>
        <p>An organizer will reach out to you shortly.</p>
      </div>
    )
  }

  return (
    <form
      noValidate
      onSubmit={(e) => {
        e.preventDefault()
        form.handleSubmit()
      }}
      className="flex flex-col gap-6"
    >
      <p>
        Interested in getting involved with Direct Action Everywhere? Fill out
        this form and we&apos;ll contact you with opportunities!
      </p>

      <div className="grid gap-4 md:grid-cols-2">
        <form.Field name="firstName">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="firstName">First Name</Label>
              <Input
                id="firstName"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={35}
              />
              {field.state.meta.errors[0] && (
                <p className="text-sm text-destructive">
                  {field.state.meta.errors[0]?.message}
                </p>
              )}
            </div>
          )}
        </form.Field>

        <form.Field name="lastName">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="lastName">Last Name</Label>
              <Input
                id="lastName"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={35}
              />
              {field.state.meta.errors[0] && (
                <p className="text-sm text-destructive">
                  {field.state.meta.errors[0]?.message}
                </p>
              )}
            </div>
          )}
        </form.Field>

        <form.Field name="email">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={60}
              />
              {field.state.meta.errors[0] && (
                <p className="text-sm text-destructive">
                  {field.state.meta.errors[0]?.message}
                </p>
              )}
            </div>
          )}
        </form.Field>

        <form.Field name="phone">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="phone">Phone</Label>
              <Input
                id="phone"
                type="text"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                maxLength={35}
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

      <form.Field name="location">
        {(field) => (
          <div className="space-y-1">
            <Label htmlFor="citySearch">City</Label>
            <CityAutocomplete
              id="citySearch"
              placeholder="Enter your city & country"
              apiKey={placesApiKey}
              hasError={!!field.state.meta.errors[0]}
              onSelect={(value) => field.handleChange(value)}
              onNoResults={() => field.handleChange(null)}
            />
            {!field.state.value && (
              <p className="text-sm italic text-muted-foreground">
                Please make a selection in the dropdown list that appears here
                as you type a city name.
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

      <form.Field name="interest">
        {(field) => (
          <fieldset className="space-y-2">
            <legend className="text-sm font-medium">
              What are you most interested in?
            </legend>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                name="interest"
                value="participate"
                checked={field.state.value === 'participate'}
                onChange={() => field.handleChange('participate')}
                className="text-primary focus:ring-ring"
              />
              I&apos;d like to participate
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                name="interest"
                value="organize"
                checked={field.state.value === 'organize'}
                onChange={() => field.handleChange('organize')}
                className="text-primary focus:ring-ring"
              />
              I&apos;d like to organize or start a chapter
            </label>
            {field.state.meta.errors[0] && (
              <p className="text-sm text-destructive">
                {field.state.meta.errors[0]?.message}
              </p>
            )}
          </fieldset>
        )}
      </form.Field>

      <div className="space-y-3">
        <p>
          I am not law enforcement and my motive for expressing interest is a
          desire to help end animal exploitation. From this point forward, I
          commit to upholding DxE&apos;s{' '}
          <a
            href="https://dxe.io/values"
            target="_blank"
            rel="noreferrer"
            className="underline"
          >
            values
          </a>{' '}
          and{' '}
          <a
            href="https://dxe.io/conduct"
            target="_blank"
            rel="noreferrer"
            className="underline"
          >
            code of conduct
          </a>{' '}
          and understand that I may be removed if I fail to do so.
        </p>

        <form.Field name="termsAgreed">
          {(field) => (
            <div className="space-y-1">
              <button
                type="button"
                onClick={() => field.handleChange(!field.state.value)}
                aria-pressed={field.state.value}
                className={`inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium transition-colors ${
                  field.state.value
                    ? 'border-primary bg-primary text-primary-foreground'
                    : 'border-input bg-background hover:bg-accent'
                }`}
              >
                Yes, I agree with the above statement.
              </button>
              {field.state.meta.errors[0] && (
                <p className="text-sm text-destructive">
                  {field.state.meta.errors[0]?.message}
                </p>
              )}
            </div>
          )}
        </form.Field>
      </div>

      <form.Field name="involvement">
        {(field) => (
          <div className="space-y-1">
            <Label htmlFor="involvement">
              Describe your interest in and/or experience with animal rights
              activism.
            </Label>
            <Textarea
              id="involvement"
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={field.handleBlur}
              maxLength={500}
              placeholder="Let us know where we can get you plugged in by telling us about your skills and interests."
            />
          </div>
        )}
      </form.Field>

      <div>
        <Button type="submit" disabled={isSubmitting}>
          Submit
        </Button>
      </div>
    </form>
  )
}
