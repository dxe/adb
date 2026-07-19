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

const ERROR_MESSAGE =
  'Sorry, there was an error submitting your form. Please try again.'

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
  involvement: z.string().max(500),
})

type Interest = 'participate' | 'organize' | ''

interface FormValues {
  firstName: string
  lastName: string
  email: string
  phone: string
  involvement: string
}

const initialLocation: CityValue = {
  city: '',
  state: '',
  country: '',
  lat: 0,
  lng: 0,
}

export function InternationalForm() {
  const [submitSuccess, setSubmitSuccess] = useState(false)
  const [interest, setInterest] = useState<Interest>('')
  const [termsAgreed, setTermsAgreed] = useState(false)
  const [location, setLocation] = useState<CityValue>(initialLocation)
  const [locationChosen, setLocationChosen] = useState(false)

  // The Google Places API key is served by a small public endpoint at
  // runtime (it is referrer-restricted, and the legacy Go-templated
  // /international page already embedded it for anonymous visitors). If the
  // fetch fails or the key is empty, the form still works — the city field
  // just won't offer autocomplete suggestions.
  const { data: placesApiKey } = useQuery({
    queryKey: [API_PATH.PLACES_API_KEY],
    queryFn: ({ signal }) => apiClient.getPlacesApiKey(signal),
    staleTime: Infinity,
    retry: 1,
  })

  const mutation = useMutation({
    mutationFn: async (values: FormValues) => {
      return apiClient.submitInternationalForm({
        firstName: values.firstName,
        lastName: values.lastName,
        email: values.email,
        phone: values.phone,
        interest: interest as 'participate' | 'organize',
        involvement: values.involvement,
        city: location.city,
        state: location.state,
        country: location.country,
        lat: location.lat,
        lng: location.lng,
      })
    },
    onSuccess: () => {
      toast.success('Submitted!')
      setSubmitSuccess(true)
    },
    onError: () => {
      // Always show the generic message (matching the legacy Vue form) —
      // backend error text may contain raw internal details (e.g. wrapped DB
      // errors) that shouldn't be shown to anonymous visitors.
      toast.error(ERROR_MESSAGE)
    },
  })

  const form = useForm({
    defaultValues: {
      firstName: '',
      lastName: '',
      email: '',
      phone: '',
      involvement: '',
    } as FormValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      const parsed = formSchema.safeParse(value)
      if (!parsed.success) {
        return
      }
      if (!locationChosen) {
        toast.error('Please choose your city from the dropdown list.')
        return
      }
      if (!interest) {
        toast.error(
          "Please choose whether you'd like to participate or organize.",
        )
        return
      }
      if (!termsAgreed) {
        toast.error('You must agree to the terms.')
        return
      }

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
                required
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
                required
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
                required
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
                required
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

      <div className="space-y-1">
        <Label htmlFor="citySearch">City</Label>
        <CityAutocomplete
          id="citySearch"
          placeholder="Enter your city & country"
          apiKey={placesApiKey}
          onSelect={(value) => {
            setLocation(value)
            setLocationChosen(true)
          }}
          // Intentional deviation from the legacy Vue form: it never listened
          // to the widget's no-results event, so a previously selected city
          // stayed "valid" even after the user typed over it. Resetting here
          // is stricter and prevents submitting a stale location.
          onNoResults={() => setLocationChosen(false)}
        />
        {!locationChosen && (
          <p className="text-sm italic text-muted-foreground">
            Please make a selection in the dropdown list that appears here as
            you type a city name.
          </p>
        )}
      </div>

      <fieldset className="space-y-2">
        <legend className="text-sm font-medium">
          What are you most interested in?
        </legend>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="radio"
            name="interest"
            value="participate"
            checked={interest === 'participate'}
            onChange={() => setInterest('participate')}
            required
            className="text-primary focus:ring-ring"
          />
          I&apos;d like to participate
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="radio"
            name="interest"
            value="organize"
            checked={interest === 'organize'}
            onChange={() => setInterest('organize')}
            required
            className="text-primary focus:ring-ring"
          />
          I&apos;d like to organize or start a chapter
        </label>
      </fieldset>

      <div className="space-y-3">
        <p>
          I am not law enforcement and my motive for expressing interest is a
          desire to help end animal exploitation. From this point forward, I
          commit to upholding DxE&apos;s{' '}
          <a
            href="http://dxe.io/values"
            target="_blank"
            rel="noreferrer"
            className="underline"
          >
            values
          </a>{' '}
          and{' '}
          <a
            href="http://dxe.io/conduct"
            target="_blank"
            rel="noreferrer"
            className="underline"
          >
            code of conduct
          </a>{' '}
          and understand that I may be removed if I fail to do so.
        </p>

        {/* Like the legacy Buefy radio-button, agreement can be selected but
            not un-selected (a radio can't be un-checked once chosen). */}
        <button
          type="button"
          onClick={() => setTermsAgreed(true)}
          aria-pressed={termsAgreed}
          className={`inline-flex items-center gap-2 rounded-md border px-3 py-2 text-sm font-medium transition-colors ${
            termsAgreed
              ? 'border-primary bg-primary text-primary-foreground'
              : 'border-input bg-background hover:bg-accent'
          }`}
        >
          Yes, I agree with the above statement.
        </button>
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
