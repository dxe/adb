'use client'

import { useCallback, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { API_PATH, apiClient } from '@/lib/api'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getActivistDisplayName } from '../display-name'
import { ActivistDetails } from './activist-details'
import { ActivistEngagement } from './activist-engagement'
import { ActivistHeader } from './activist-header'
import { HideActivistDialog } from './hide-activist-dialog'
import { LogInteractionDialog } from './log-interaction-dialog'
import { MergeActivistDialog } from './merge-activist-dialog'

function useActivist(activistId: number) {
  return useQuery({
    queryKey: [API_PATH.ACTIVIST_GET, activistId],
    queryFn: ({ signal }) => apiClient.getActivist(activistId, signal),
  })
}

export function Activist({ activistId }: { activistId: number }) {
  const { data: activist, isError, isLoading } = useActivist(activistId)
  const [isHideDialogOpen, setIsHideDialogOpen] = useState(false)
  const [isMergeDialogOpen, setIsMergeDialogOpen] = useState(false)
  const [isLogInteractionDialogOpen, setIsLogInteractionDialogOpen] =
    useState(false)
  const [tab, setTab] = useState('details')
  const [isDetailsDirty, setIsDetailsDirty] = useState(false)

  // Switching tabs unmounts the details form, so confirm before discarding.
  const handleTabChange = useCallback(
    (nextTab: string) => {
      if (
        isDetailsDirty &&
        !window.confirm(
          'You have unsaved changes. Discard them and leave this tab?',
        )
      ) {
        return
      }
      setIsDetailsDirty(false)
      setTab(nextTab)
    },
    [isDetailsDirty],
  )

  if (isLoading) {
    return <div className="animate-pulse">Loading activist details...</div>
  }
  if (isError || !activist) {
    return <div>Unable to load activist details.</div>
  }

  const displayName = getActivistDisplayName(activist)

  return (
    <>
      <HideActivistDialog
        open={isHideDialogOpen}
        onOpenChange={setIsHideDialogOpen}
        activistId={activistId}
        activistName={displayName.text ?? ''}
      />
      <MergeActivistDialog
        open={isMergeDialogOpen}
        onOpenChange={setIsMergeDialogOpen}
        activistId={activistId}
        activistName={displayName.text ?? ''}
      />

      <LogInteractionDialog
        open={isLogInteractionDialogOpen}
        onOpenChange={setIsLogInteractionDialogOpen}
        activistId={activistId}
        activistName={displayName.text ?? ''}
      />

      <ActivistHeader
        activist={activist}
        onLogInteraction={() => setIsLogInteractionDialogOpen(true)}
        onMerge={() => setIsMergeDialogOpen(true)}
        onHide={() => setIsHideDialogOpen(true)}
      />

      <Tabs value={tab} onValueChange={handleTabChange}>
        <TabsList className="h-auto w-full justify-start gap-4 rounded-none border-b bg-transparent p-0">
          <TabsTrigger
            value="details"
            className="rounded-none border-b-2 border-transparent px-1 pb-2 data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none"
          >
            Details
          </TabsTrigger>
          <TabsTrigger
            value="engagement"
            className="rounded-none border-b-2 border-transparent px-1 pb-2 data-[state=active]:border-primary data-[state=active]:bg-transparent data-[state=active]:shadow-none"
          >
            Engagement
          </TabsTrigger>
        </TabsList>

        <TabsContent value="details" className="mt-6">
          <ActivistDetails
            activistId={activistId}
            activist={activist}
            onDirtyChange={setIsDetailsDirty}
          />
        </TabsContent>

        <TabsContent value="engagement" className="mt-6">
          <ActivistEngagement activistId={activistId} />
        </TabsContent>
      </Tabs>
    </>
  )
}
