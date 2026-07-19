'use client'

import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useQueryStates } from 'nuqs'
import { Loader2, Plus } from 'lucide-react'
import { API_PATH, apiClient, CircleGroup } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { circleSearchParamParsers } from './search-params'
import { CircleTable } from './circle-table'
import { CircleFormDialog } from './circle-form-dialog'
import { DeleteCircleDialog } from './delete-circle-dialog'

export default function CirclesPage() {
  const [{ type: mode }, setSearchParams] = useQueryStates(
    circleSearchParamParsers,
  )

  const {
    data: circles,
    isLoading: isCirclesLoading,
    isError,
  } = useQuery({
    queryKey: [API_PATH.CIRCLE_LIST],
    queryFn: ({ signal }) => apiClient.getCircles(signal),
  })

  const {
    data: activistNames,
    isLoading: isActivistsLoading,
    isError: isActivistsError,
  } = useQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_CHAPTER_MEMBERS],
    queryFn: ({ signal }) => apiClient.getChapterMemberActivistNames(signal),
  })

  const [isMembersVisible, setIsMembersVisible] = useState(false)
  const [editingCircle, setEditingCircle] = useState<CircleGroup | null>(null)
  const [isFormOpen, setIsFormOpen] = useState(false)
  const [deletingCircle, setDeletingCircle] = useState<CircleGroup | null>(null)

  // /circle/list returns both circle types; filter by mode client-side like the legacy Vue page.
  const filteredCircles = useMemo(() => {
    if (!circles) return []
    const wantedType = mode === 'geo' ? 'geo-circle' : 'circle'
    return circles.filter((c) => c.type === wantedType)
  }, [circles, mode])

  const isLoading = isCirclesLoading || isActivistsLoading
  const circleLabel = mode === 'geo' ? 'Geo-Circle' : 'Circle'

  const openCreateDialog = () => {
    setEditingCircle(null)
    setIsFormOpen(true)
  }

  const openEditDialog = (circle: CircleGroup) => {
    setEditingCircle(circle)
    setIsFormOpen(true)
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">
          {mode === 'geo' ? 'Geo-Circles' : 'Circles'}
        </h1>
        <p className="text-muted-foreground text-sm">
          Manage {mode === 'geo' ? 'geo-circles' : 'interest circles'} and their
          members.
        </p>
      </div>

      <div className="flex items-center gap-1 border-b">
        <button
          type="button"
          onClick={() => setSearchParams({ type: 'interest' })}
          className={cn(
            'border-b-2 px-3 py-2 text-sm font-medium',
            mode === 'interest'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          Interest Circles
        </button>
        <button
          type="button"
          onClick={() => setSearchParams({ type: 'geo' })}
          className={cn(
            'border-b-2 px-3 py-2 text-sm font-medium',
            mode === 'geo'
              ? 'border-primary text-primary'
              : 'border-transparent text-muted-foreground hover:text-foreground',
          )}
        >
          Geo-Circles
        </button>
      </div>

      <div className="flex items-center justify-between gap-3">
        <Button type="button" onClick={openCreateDialog}>
          <Plus className="h-4 w-4" />
          New {circleLabel}
        </Button>

        {mode === 'geo' && (
          <Button
            type="button"
            variant="outline"
            onClick={() => setIsMembersVisible((v) => !v)}
          >
            {isMembersVisible ? 'Hide' : 'Show'} members
          </Button>
        )}
      </div>

      {isLoading ? (
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Loader2 className="h-4 w-4 animate-spin" />
          Loading {mode === 'geo' ? 'geo-circles' : 'circles'}...
        </div>
      ) : isError ? (
        <div className="text-sm text-destructive">
          Failed to load circles. Please try again.
        </div>
      ) : (
        <>
          {isActivistsError && (
            <div className="text-sm text-destructive">
              Failed to load activist names. The host and member autocomplete
              will be unavailable — reload the page to try again.
            </div>
          )}
          <CircleTable
            circles={filteredCircles}
            mode={mode}
            isMembersVisible={isMembersVisible}
            onEdit={openEditDialog}
            onDelete={setDeletingCircle}
          />
        </>
      )}

      {isFormOpen && (
        <CircleFormDialog
          key={editingCircle?.id ?? 'new'}
          open={isFormOpen}
          onOpenChange={setIsFormOpen}
          mode={mode}
          circle={editingCircle}
          activistNames={activistNames ?? []}
        />
      )}

      {deletingCircle && (
        <DeleteCircleDialog
          open={!!deletingCircle}
          onOpenChange={(open) => {
            if (!open) setDeletingCircle(null)
          }}
          circle={deletingCircle}
        />
      )}
    </div>
  )
}
