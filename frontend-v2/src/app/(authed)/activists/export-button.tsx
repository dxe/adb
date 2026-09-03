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

export function ExportButton({ queryOptions }: ExportButtonProps) {
  const [isExporting, setIsExporting] = useState(false)
  const [isCopying, setIsCopying] = useState(false)
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

  const isBusy = isExporting || isCopying

  const runCopy = async () => {
    if (isBusy) return
    const controller = abortControllerRef.current
    if (!controller) return
    const { signal } = controller
    setIsCopying(true)
    // Start the request synchronously and hand the pending text to the
    // clipboard, rather than awaiting the export first: browsers only honor a
    // clipboard write while the click's user-activation grant is still live.
    const tsvPromise = apiClient.exportActivistsTsvText(queryOptions, signal)
    try {
      const [, tsv] = await Promise.all([
        writeToClipboard(tsvPromise),
        tsvPromise,
      ])
      if (signal.aborted) return
      const rows = countDataRows(tsv)
      toast.success(
        `Copied ${rows.toLocaleString()} ${rows === 1 ? 'activist' : 'activists'} to clipboard`,
      )
    } catch (err) {
      if (signal.aborted) return
      console.error('Failed to copy activists to clipboard', err)
      toast.error(
        err instanceof Error && err.message
          ? `Failed to copy activists: ${err.message}`
          : 'Failed to copy activists. Please try again.',
      )
    } finally {
      if (!signal.aborted) setIsCopying(false)
    }
  }

  const runExport = async (
    fetchBlob: (signal: AbortSignal) => Promise<Blob>,
    filenamePrefix: string,
  ) => {
    if (isBusy) return
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
          disabled={isBusy}
        >
          <Download className="h-4 w-4" />
          {isCopying ? 'Copying…' : isExporting ? 'Exporting…' : 'Export'}
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
              runCopy()
            }}
          >
            Copy to clipboard
          </button>
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
            CSV
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
            Spoke CSV
          </button>
        </div>
      </PopoverContent>
    </Popover>
  )
}

// Counts the data rows in a TSV export.
// A header row must be included, every row must be one line (values must not contain newlines)
// and every row must be newline-terminated (including the last one).
function countDataRows(tsv: string) {
  let newlines = 0
  for (let i = tsv.indexOf('\n'); i !== -1; i = tsv.indexOf('\n', i + 1)) {
    newlines++
  }
  return Math.max(newlines - 1, 0)
}

// Writes text to the clipboard via the async Clipboard API. The text is passed
// as a promise so the write can be issued before the export has finished
// downloading, keeping the user-activation grant that Safari requires. Browsers
// whose ClipboardItem rejects a pending value fall back to writeText once the
// text has arrived.
async function writeToClipboard(text: Promise<string>) {
  if (!navigator.clipboard) {
    throw new Error('the clipboard is unavailable in this browser')
  }
  if (typeof ClipboardItem !== 'undefined') {
    try {
      await navigator.clipboard.write([
        new ClipboardItem({
          'text/plain': text.then((t) => new Blob([t], { type: 'text/plain' })),
        }),
      ])
      return
    } catch (err) {
      console.warn('ClipboardItem write failed; falling back to writeText', err)
    }
  }
  await navigator.clipboard.writeText(await text)
}
