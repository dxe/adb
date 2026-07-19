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

/**
 * Orchestrates the multi-step "Apply" flow, mirroring the top-level
 * `local` / `showForm` / `submitSuccess` state machine in
 * frontend/FormApply.vue:
 *  1. Ask whether the applicant lives near the SF Bay Area chapter. If not,
 *     send them to the (legacy, still Vue-served) /international page.
 *  2. Show informational copy about becoming a chapter member, with an
 *     "Apply now" button that reveals the form.
 *  3. Show the application form. On submit, POST to /apply.
 *  4. Show a thank-you message on success.
 */
export function ApplyForm() {
  const [local, setLocal] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [submitSuccess, setSubmitSuccess] = useState(false)

  const mutation = useMutation({
    mutationFn: (payload: ApplicationFormPayload) =>
      apiClient.submitApplicationForm(payload),
    onSuccess: () => {
      toast.success('Submitted!')
      setSubmitSuccess(true)
    },
    onError: () => {
      toast.error(ERROR_MESSAGE)
    },
  })

  function handleNotLocal() {
    // The international application flow hasn't been ported yet, so this
    // intentionally sends the user to the legacy Vue-served page (outside
    // of /v2), matching frontend/FormApply.vue's notLocal().
    window.location.href = '/international'
  }

  function handleApply() {
    setShowForm(true)
    window.scrollTo(0, 0)
  }

  return (
    <>
      {!local && (
        <LocalCheck
          onLocal={() => setLocal(true)}
          onNotLocal={handleNotLocal}
        />
      )}

      {submitSuccess && <ThankYou />}

      {local && !submitSuccess && (
        <>
          {!showForm && <ChapterMemberInfo onApply={handleApply} />}

          {showForm && (
            <ApplicationFields
              onSubmit={(payload) => mutation.mutate(payload)}
              isSubmitting={mutation.isPending}
            />
          )}
        </>
      )}
    </>
  )
}
