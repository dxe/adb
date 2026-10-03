'use client'

import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import {
  API_PATH,
  apiClient,
  CHAPTER_ADMIN_QUERY_KEY,
  ChapterAdmin,
} from '@/lib/api'
import { ConfirmDialog } from '@/components/confirm-dialog'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  chapter: ChapterAdmin
}

export function DeleteChapterDialog({ open, onOpenChange, chapter }: Props) {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: () => apiClient.deleteChapterAdmin(chapter.ChapterID),
    onSuccess: () => {
      // Prefix match also refreshes the chapter-picker and intl-organizers caches.
      queryClient.invalidateQueries({ queryKey: [API_PATH.CHAPTER_LIST] })
      queryClient.setQueryData<ChapterAdmin[]>(CHAPTER_ADMIN_QUERY_KEY, (old) =>
        old?.filter((c) => c.ChapterID !== chapter.ChapterID),
      )
      toast.success(`${chapter.Name} deleted`)
      onOpenChange(false)
    },
    onError: (err: Error) => {
      toast.error(err.message || 'Failed to delete chapter. Please try again.')
    },
  })

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Delete chapter"
      confirmLabel="Delete"
      pendingLabel="Deleting..."
      onConfirm={() => mutation.mutate()}
      isPending={mutation.isPending}
    >
      Are you sure you want to delete{' '}
      <strong>
        {chapter.Flag} {chapter.Name}
      </strong>
      ?
    </ConfirmDialog>
  )
}
