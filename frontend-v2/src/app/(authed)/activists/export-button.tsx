'use client'

import { useEffect, useMemo, useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { ChevronDown, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { apiClient, QueryActivistOptions } from '@/lib/api'

interface ExportButtonProps {
  queryOptions: QueryActivistOptions
}

/**
 * Renders a popover "Export" button that lets the user download activists CSVs for the current filters.
 *
 * Renders a menu with two export actions ("Current columns" and "Spoke columns"), manages export state and cancellation, and triggers client-side CSV downloads with date-stamped filenames.
 *
 * @param queryOptions - Filters and shape used to request CSV exports; the component uses `queryOptions` for the "Current columns" export and a variant with `shape.columns` forced to an empty array for the "Spoke columns" export.
 * @returns The export button and popover menu as JSX.
 */
export function ExportButton({ queryOptions }: ExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false)
  const [isMenuOpen, setIsMenuOpen] = useState(false)
  const abortControllerRef = useRef<AbortController | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    abortControllerRef.current = controller
    return () => {
      controller.abort()
      abortControllerRef.current = null
    }
  }, [])

  // The spoke export uses the current filters but a server-selected column
  // set, so we send an empty columns array. The server hard-codes the spoke
  // columns and rejects a non-empty list.
  const spokeQueryOptions = useMemo<QueryActivistOptions>(
    () => ({
      ...queryOptions,
      shape: { ...queryOptions.shape, columns: [] },
    }),
    [queryOptions],
  )

  const runExport = async (
    fetchBlob: (signal: AbortSignal) => Promise<Blob>,
    filenamePrefix: string,
  ) => {
    if (isExporting) return
    const controller = abortControllerRef.current
    if (!controller) return
    const { signal } = controller
    setIsExporting(true)
    let url: string | undefined
    try {
      const blob = await fetchBlob(signal)
      if (signal.aborted) return
      url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${filenamePrefix}-${new Date().toISOString().slice(0, 10)}.csv`
      document.body.appendChild(a)
      a.click()
      a.remove()
    } catch (err) {
      if (signal.aborted) return
      console.error('Failed to export activists CSV', err)
      toast.error(
        err instanceof Error && err.message
          ? `Failed to export activists: ${err.message}`
          : 'Failed to export activists. Please try again.',
      )
    } finally {
      if (url) URL.revokeObjectURL(url)
      if (!signal.aborted) setIsExporting(false)
    }
  }

  return (
    <Popover open={isMenuOpen} onOpenChange={setIsMenuOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-12 gap-1"
          disabled={isExporting}
        >
          <Download className="h-4 w-4" />
          {isExporting ? 'Exporting…' : 'Export'}
          <ChevronDown className="h-4 w-4" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-44 p-2" align="start">
        <div className="flex flex-col">
          <button
            type="button"
            className="flex w-full items-center rounded px-2 py-1.5 text-sm hover:bg-muted transition-colors text-left"
            onClick={() => {
              setIsMenuOpen(false)
              runExport(
                (signal) => apiClient.exportActivistsCsv(queryOptions, signal),
                'activists',
              )
            }}
          >
            Current columns
          </button>
          <button
            type="button"
            className="flex w-full items-center rounded px-2 py-1.5 text-sm hover:bg-muted transition-colors text-left"
            onClick={() => {
              setIsMenuOpen(false)
              runExport(
                (signal) =>
                  apiClient.exportActivistsSpokeCsv(spokeQueryOptions, signal),
                'activists-spoke',
              )
            }}
          >
            Spoke columns
          </button>
        </div>
      </PopoverContent>
    </Popover>
  )
}
