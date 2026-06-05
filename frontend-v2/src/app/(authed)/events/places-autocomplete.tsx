'use client'

import { useEffect, useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'

// The resolved place we capture on selection. No free-text is submitted — a
// place is only set via the Google autocomplete dropdown.
export type PlaceValue = {
  google_place_id: string
  location_name: string
  formatted_address: string
  lat?: number
  lng?: number
}

// Narrow typings for just the slice of the Google Maps Places API we use, so we
// avoid pulling in @types/google.maps.
type GooglePlaceResult = {
  place_id?: string
  name?: string
  formatted_address?: string
  geometry?: { location?: { lat: () => number; lng: () => number } }
}

type GoogleAutocomplete = {
  addListener: (event: string, handler: () => void) => void
  getPlace: () => GooglePlaceResult
}

type GoogleMaps = {
  maps?: {
    places?: {
      Autocomplete: new (
        input: HTMLInputElement,
        opts?: { fields?: string[]; types?: string[] },
      ) => GoogleAutocomplete
    }
  }
}

declare global {
  interface Window {
    google?: GoogleMaps
  }
}

const MAPS_SCRIPT_ID = 'google-maps-places-js'

// Builds the text shown in the input: the place/business name followed by its
// address. For a plain-address result Google returns the street address as the
// `name`, so we skip the name when the address already begins with it to avoid
// "399 4th St, 399 4th St, San Francisco…".
function displayLabel(name: string | undefined, address: string): string {
  const trimmed = (name ?? '').trim()
  if (!trimmed) return address
  if (!address) return trimmed
  if (address.toLowerCase().startsWith(trimmed.toLowerCase())) return address
  return `${trimmed}, ${address}`
}

// Loads the Maps JS `places` library once, reusing an in-flight/finished load.
function loadPlacesLibrary(apiKey: string): Promise<void> {
  if (typeof window === 'undefined') {
    return Promise.reject(new Error('not in browser'))
  }
  if (window.google?.maps?.places) return Promise.resolve()

  const existing = document.getElementById(
    MAPS_SCRIPT_ID,
  ) as HTMLScriptElement | null
  if (existing) {
    return new Promise((resolve, reject) => {
      existing.addEventListener('load', () => resolve())
      existing.addEventListener('error', () =>
        reject(new Error('Failed to load Google Maps')),
      )
    })
  }

  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.id = MAPS_SCRIPT_ID
    script.src = `https://maps.googleapis.com/maps/api/js?key=${encodeURIComponent(
      apiKey,
    )}&libraries=places`
    script.async = true
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load Google Maps'))
    document.head.appendChild(script)
  })
}

type Props = {
  // Referrer-restricted Google Places key, served from the backend config.
  apiKey: string
  // The currently selected formatted address (display text for the input).
  value: string
  // The selected place/business name, shown alongside the address.
  locationName?: string
  onSelect: (place: PlaceValue) => void
  onClear: () => void
  disabled?: boolean
  hasError?: boolean
  id?: string
}

export function PlacesAutocomplete({
  apiKey,
  value,
  locationName,
  onSelect,
  onClear,
  disabled,
  hasError,
  id,
}: Props) {
  const inputRef = useRef<HTMLInputElement | null>(null)
  const [text, setText] = useState(() => displayLabel(locationName, value))
  // No key: there's nothing to load, so start in the error (unavailable) state.
  const [status, setStatus] = useState<'loading' | 'ready' | 'error'>(
    apiKey ? 'loading' : 'error',
  )
  // Set when we update the input programmatically (a dropdown selection or a
  // parent-driven value sync) so the resulting onChange isn't mistaken for the
  // user typing — see the onChange handler below.
  const programmaticEditRef = useRef(false)

  // Keep the displayed text in sync when the parent value changes (e.g. loading
  // an existing event, or the Online checkbox clearing the field). Done during
  // render via a previous-value check rather than in an effect, so the text
  // updates in the same commit (no flash) without a cascading re-render.
  const [syncedValue, setSyncedValue] = useState({ value, locationName })
  if (
    syncedValue.value !== value ||
    syncedValue.locationName !== locationName
  ) {
    setSyncedValue({ value, locationName })
    setText(displayLabel(locationName, value))
  }

  useEffect(() => {
    // No key: status is already 'error' from initialization; nothing to load.
    if (!apiKey) return
    let cancelled = false
    // The field silently disables itself on failure, so also surface a toast —
    // otherwise the user is left wondering why location search isn't working.
    // A stable id dedupes it across re-runs (e.g. React strict-mode double mount).
    const reportLoadError = () => {
      if (cancelled) return
      setStatus('error')
      toast.error(
        'Location search failed to load. You can still save without a location.',
        {
          id: 'places-load-error',
        },
      )
    }
    loadPlacesLibrary(apiKey)
      .then(() => {
        if (cancelled) return
        const places = window.google?.maps?.places
        if (!places || !inputRef.current) {
          reportLoadError()
          return
        }
        const autocomplete = new places.Autocomplete(inputRef.current, {
          fields: ['place_id', 'name', 'formatted_address', 'geometry'],
        })
        autocomplete.addListener('place_changed', () => {
          const place = autocomplete.getPlace()
          if (!place.place_id) return
          const formatted = place.formatted_address ?? ''
          // Selecting a suggestion mutates the input value; flag it so the
          // onChange it may fire isn't treated as the user editing (which would
          // immediately clear the place we're selecting).
          programmaticEditRef.current = true
          setText(displayLabel(place.name, formatted))
          onSelect({
            google_place_id: place.place_id,
            location_name: place.name ?? '',
            formatted_address: formatted,
            lat: place.geometry?.location?.lat(),
            lng: place.geometry?.location?.lng(),
          })
          setTimeout(() => {
            programmaticEditRef.current = false
          }, 0)
        })
        setStatus('ready')
      })
      .catch(() => {
        reportLoadError()
      })
    return () => {
      cancelled = true
    }
    // onSelect is stable enough for our usage; we intentionally load once.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [apiKey])

  return (
    <div className="flex flex-col gap-1">
      <Input
        id={id}
        ref={inputRef}
        value={text}
        disabled={disabled || status === 'error'}
        placeholder={
          status === 'error'
            ? 'Location search unavailable'
            : 'Search for a place'
        }
        onChange={(e) => {
          setText(e.target.value)
          if (programmaticEditRef.current) {
            // Value set by a selection/sync, not the user — keep the place.
            programmaticEditRef.current = false
            return
          }
          // Any manual edit invalidates the committed place: a free-text value
          // is never submitted, so drop the selected place until the user picks
          // a fresh suggestion. Otherwise the visible text and the submitted
          // location could silently drift apart.
          onClear()
        }}
        className={cn(hasError && 'border-red-500')}
        autoComplete="off"
      />
      {status === 'error' && !apiKey && (
        <p className="text-xs text-muted-foreground">
          Location search is unavailable (Google Places key not configured).
        </p>
      )}
    </div>
  )
}
