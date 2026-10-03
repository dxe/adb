'use client'

import Link from 'next/link'
import { useSuspenseQuery } from '@tanstack/react-query'
import { notFound } from 'next/navigation'
import { ArrowLeft, Loader2, Save } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { apiClient, CHAPTER_ADMIN_QUERY_KEY, ChapterAdmin } from '@/lib/api'
import { useChapterForm } from './useChapterForm'
import {
  BasicInfoSection,
  DatesSection,
  LocationSection,
  OrganizersSection,
  SocialLinksSection,
} from './chapter-form-sections'

function ChapterFormInner({
  chapterId,
  chapter,
}: {
  chapterId?: number
  chapter: ChapterAdmin | undefined
}) {
  const isEditing = chapterId != null
  const { form, isSubmitting } = useChapterForm({ chapter, isEditing })

  return (
    <div className="flex flex-col gap-6">
      <div className="flex items-start justify-between gap-2">
        <div className="flex flex-col gap-1">
          <p className="text-sm uppercase tracking-wide text-muted-foreground">
            {isEditing ? 'Edit chapter' : 'Create chapter'}
          </p>
          <h1 className="text-2xl font-semibold">
            {chapter ? `${chapter.Flag} ${chapter.Name}` : 'New Chapter'}
          </h1>
        </div>
        <Button variant="ghost" asChild>
          <Link href="/chapters">
            <ArrowLeft className="h-4 w-4" />
            Back to list
          </Link>
        </Button>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault()
          form.handleSubmit()
        }}
        className="space-y-8"
      >
        <div className="flex items-center justify-end">
          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="h-4 w-4" />
                Save chapter
              </>
            )}
          </Button>
        </div>

        <BasicInfoSection form={form} isEditing={isEditing} />
        <LocationSection form={form} />
        <SocialLinksSection form={form} />
        <DatesSection form={form} />
        <OrganizersSection form={form} isEditing={isEditing} />
      </form>
    </div>
  )
}

function EditChapterForm({ chapterId }: { chapterId: number }) {
  const { data: chapters } = useSuspenseQuery({
    queryKey: CHAPTER_ADMIN_QUERY_KEY,
    queryFn: ({ signal }) => apiClient.getChapterAdminList(signal),
  })
  const chapter = chapters.find((c) => c.ChapterID === chapterId)
  if (!chapter) {
    notFound()
  }
  return <ChapterFormInner chapterId={chapterId} chapter={chapter} />
}

export function ChapterForm({ chapterId }: { chapterId?: number }) {
  if (chapterId != null) {
    return <EditChapterForm chapterId={chapterId} />
  }
  return <ChapterFormInner chapter={undefined} />
}
