'use client'

import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { apiClient, ApplicationFormPayload } from '@/lib/api'
import { LocalCheck } from '@/app/apply/local-check'
import { ChapterMemberInfo } from '@/app/apply/chapter-member-info'
import { ApplicationFields } from '@/app/apply/application-fields'
import { ThankYou } from '@/app/apply/thank-you'

const ERROR_MESSAGE =
  'Sorry, there was an error submitting your form. Please try again.'

type Step = 'localCheck' | 'info' | 'fields' | 'thankYou'

/** Orchestrates the multi-step "Apply" flow ported from frontend/FormApply.vue. */
export function ApplyForm() {
  const [step, setStep] = useState<Step>('localCheck')

  const mutation = useMutation({
    mutationFn: (payload: ApplicationFormPayload) =>
      apiClient.submitApplicationForm(payload),
    onSuccess: () => {
      toast.success('Submitted!')
      setStep('thankYou')
    },
    onError: () => {
      toast.error(ERROR_MESSAGE)
    },
  })

  function handleNotLocal() {
    // The international flow isn't ported yet; go to the legacy Vue-served page.
    window.location.href = '/international'
  }

  function handleApply() {
    setStep('fields')
    window.scrollTo(0, 0)
  }

  switch (step) {
    case 'localCheck':
      return (
        <LocalCheck
          onLocal={() => setStep('info')}
          onNotLocal={handleNotLocal}
        />
      )
    case 'info':
      return <ChapterMemberInfo onApply={handleApply} />
    case 'fields':
      return (
        <ApplicationFields
          onSubmit={(payload) => mutation.mutate(payload)}
          isSubmitting={mutation.isPending}
        />
      )
    case 'thankYou':
      return <ThankYou />
  }
}
