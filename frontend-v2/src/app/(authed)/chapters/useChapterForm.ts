'use client'

import { useMemo } from 'react'
import { useForm } from '@tanstack/react-form'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRouter } from 'next/navigation'
import toast from 'react-hot-toast'
import {
  API_PATH,
  apiClient,
  CHAPTER_ADMIN_QUERY_KEY,
  ChapterAdmin,
} from '@/lib/api'
import {
  buildChapterPayload,
  chapterFormSubmitSchema,
  toInitialValues,
} from './chapter-form-schema'

export type ChapterFormApi = ReturnType<typeof useChapterForm>['form']

export function useChapterForm({
  chapter,
  isEditing,
}: {
  chapter: ChapterAdmin | undefined
  isEditing: boolean
}) {
  const router = useRouter()
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (payload: Partial<ChapterAdmin>) =>
      apiClient.saveChapterAdmin(payload),
    onSuccess: (savedChapter) => {
      // Prefix match also refreshes the chapter-picker and intl-organizers caches.
      queryClient.invalidateQueries({ queryKey: [API_PATH.CHAPTER_LIST] })
      queryClient.setQueryData<ChapterAdmin[]>(
        CHAPTER_ADMIN_QUERY_KEY,
        (old) => {
          if (!old) return old
          const exists = old.some((c) => c.ChapterID === savedChapter.ChapterID)
          return exists
            ? old.map((c) =>
                c.ChapterID === savedChapter.ChapterID ? savedChapter : c,
              )
            : [savedChapter, ...old]
        },
      )
      toast.success(`${savedChapter.Name} saved`)
      if (isEditing) {
        router.push('/chapters')
      } else {
        router.push(`/chapters/${savedChapter.ChapterID}`)
      }
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Failed to save chapter'
      toast.error(message)
    },
  })

  const initialValues = useMemo(() => toInitialValues(chapter), [chapter])

  const form = useForm({
    defaultValues: initialValues,
    validators: {
      onSubmit: chapterFormSubmitSchema,
    },
    onSubmit: async ({ value }) => {
      // validators.onSubmit already blocked invalid values; this parse only
      // narrows the type to ChapterFormSubmitValues.
      const parsed = chapterFormSubmitSchema.parse(value)
      await mutation.mutateAsync(buildChapterPayload(chapter, parsed))
    },
  })

  return { form, isSubmitting: mutation.isPending }
}
