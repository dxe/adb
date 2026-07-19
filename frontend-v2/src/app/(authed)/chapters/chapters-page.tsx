'use client'

import { useCallback, useMemo, useState } from 'react'
import Link from 'next/link'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Download, Loader2, Plus } from 'lucide-react'
import { apiClient, CHAPTER_ADMIN_QUERY_KEY, ChapterAdmin } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ChapterTable } from './chapter-table'
import { isDateInLastThreeMonths } from './chapter-utils'

function useMentorOptions(chapters: ChapterAdmin[]): string[] {
  return useMemo(() => {
    const mentors = new Set<string>()
    chapters.forEach((c) => {
      if (c.Mentor) mentors.add(c.Mentor)
    })
    return ['All', ...Array.from(mentors).sort()]
  }, [chapters])
}

export default function ChaptersPage() {
  const queryClient = useQueryClient()
  const { data: chapters, isLoading } = useQuery({
    queryKey: CHAPTER_ADMIN_QUERY_KEY,
    queryFn: ({ signal }) => apiClient.getChapterAdminList(signal),
  })

  const [mentorFilter, setMentorFilter] = useState('All')
  const [filterName, setFilterName] = useState('')
  const [showFacebookColumns, setShowFacebookColumns] = useState(false)

  const mentorOptions = useMentorOptions(chapters ?? [])

  const filteredChapters = useMemo(() => {
    if (!chapters) return []
    return chapters.filter((c) => {
      const matchesMentor = mentorFilter === 'All' || c.Mentor === mentorFilter
      const matchesName = c.Name.toLowerCase().startsWith(
        filterName.toLowerCase(),
      )
      return matchesMentor && matchesName
    })
  }, [chapters, mentorFilter, filterName])

  const totalChapters = filteredChapters.filter(
    (c) => c.Region !== 'Online',
  ).length
  const activeChapters = filteredChapters.filter(
    (c) => c.Region !== 'Online' && isDateInLastThreeMonths(c.LastAction),
  ).length

  const deleteMutation = useMutation({
    mutationFn: (chapterId: number) => apiClient.deleteChapterAdmin(chapterId),
    onSuccess: (_, chapterId) => {
      queryClient.setQueryData<ChapterAdmin[]>(CHAPTER_ADMIN_QUERY_KEY, (old) =>
        old?.filter((c) => c.ChapterID !== chapterId),
      )
    },
    onError: (error: Error) => {
      toast.error(
        error.message || 'Failed to delete chapter. Please try again.',
      )
    },
  })

  const { mutate: deleteChapter } = deleteMutation
  const handleDelete = useCallback(
    (chapter: ChapterAdmin) => {
      const confirmed = window.confirm(
        `Are you sure you want to delete ${chapter.Flag} ${chapter.Name}?`,
      )
      if (confirmed) {
        deleteChapter(chapter.ChapterID, {
          onSuccess: () => toast.success(`${chapter.Name} deleted`),
        })
      }
    },
    [deleteChapter],
  )

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <h1 className="text-2xl font-semibold">Chapters</h1>
          <p className="text-muted-foreground text-sm">
            Manage DxE chapters worldwide.
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-6">
          <div className="text-center">
            <p className="text-xs uppercase tracking-wide text-muted-foreground">
              Total Chapters
            </p>
            <p className="text-xl font-semibold">{totalChapters}</p>
          </div>
          <div className="text-center">
            <p className="text-xs uppercase tracking-wide text-muted-foreground">
              Active Chapters
            </p>
            <p className="text-xl font-semibold">{activeChapters}</p>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-wrap items-end gap-3">
          <Button asChild>
            <Link href="/chapters/new">
              <Plus className="h-4 w-4" />
              New chapter
            </Link>
          </Button>

          <div className="flex flex-col gap-1">
            <label
              className="text-xs text-muted-foreground"
              htmlFor="mentor-filter"
            >
              Mentor
            </label>
            <Select value={mentorFilter} onValueChange={setMentorFilter}>
              <SelectTrigger id="mentor-filter" className="w-40">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {mentorOptions.map((mentor) => (
                  <SelectItem key={mentor} value={mentor}>
                    {mentor}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-1">
            <label
              className="text-xs text-muted-foreground"
              htmlFor="name-filter"
            >
              Filter by name
            </label>
            <Input
              id="name-filter"
              value={filterName}
              onChange={(e) => setFilterName(e.target.value)}
              className="w-48"
            />
          </div>

          <label className="flex items-center gap-2 pb-2 text-sm">
            <Checkbox
              checked={showFacebookColumns}
              onCheckedChange={(checked) =>
                setShowFacebookColumns(checked === true)
              }
            />
            Show FB columns
          </label>
        </div>

        <Button variant="secondary" asChild>
          <a href="/csv/international_organizers">
            <Download className="h-4 w-4" />
            Export CSV
          </a>
        </Button>
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 text-muted-foreground text-sm">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading chapters...
        </div>
      ) : (
        <ChapterTable
          chapters={filteredChapters}
          showFacebookColumns={showFacebookColumns}
          onDelete={handleDelete}
        />
      )}
    </div>
  )
}
