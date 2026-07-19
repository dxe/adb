/// <reference types="google.maps" />
'use client'

import Script from 'next/script'
import { useEffect, useRef, useState } from 'react'
import { Input } from '@/components/ui/input'

export interface CityValue {
  city: string
  state: string
  country: string
  lat: number
  lng: number
}

/**
 * Text input backed by the Google Places Autocomplete widget, restricted to
 * cities. This replaces the legacy Vue app's `vue-google-autocomplete`
 * component (frontend/external/vue-google-autocomplete). It replicates that
 * component's address-component parsing: `locality` -> city (long_name),
 * `administrative_area_level_1` -> state (short_name), `country` -> country
 * (short_name, matching DxE's mailing-list DB convention).
 *
 * `apiKey` is fetched at runtime by the parent (from the public
 * /places_api_key endpoint). While it is undefined/empty, the input still
 * renders and accepts text, but no autocomplete suggestions appear.
 */
export function CityAutocomplete({
  id,
  placeholder,
  apiKey,
  onSelect,
  onNoResults,
}: {
  id?: string
  placeholder?: string
  apiKey: string | undefined
  onSelect: (value: CityValue) => void
  onNoResults?: () => void
}) {
  const inputRef = useRef<HTMLInputElement>(null)
  // The Maps script may already be present (e.g. after a client-side
  // re-navigation, next/script won't re-fire onLoad), so probe for it.
  const [scriptLoaded, setScriptLoaded] = useState(
    () => typeof google !== 'undefined' && !!google.maps?.places,
  )

  // Keep the latest callbacks in a ref so the effect below doesn't need them
  // in its dependency array (re-running it would tear down and rebuild the
  // Autocomplete widget whenever the parent re-renders with new closures).
  const callbacksRef = useRef({ onSelect, onNoResults })
  useEffect(() => {
    callbacksRef.current = { onSelect, onNoResults }
  })

  useEffect(() => {
    if (!scriptLoaded) return
    const inputEl = inputRef.current
    // `scriptLoaded` is only set once the Maps JS API's onLoad fires, so the
    // ambient `google` namespace (from @types/google.maps) is available here.
    const places =
      typeof google !== 'undefined' ? google.maps?.places : undefined
    if (!inputEl || !places) return

    // Track the suggestion dropdowns (.pac-container) that already exist so
    // that only the one added by this widget instance is removed on cleanup.
    // Without this, React strict mode's double-invoked effects would leave an
    // orphaned duplicate dropdown attached to the document body.
    const pacContainersBefore = new Set(
      document.querySelectorAll('.pac-container'),
    )

    const autocomplete = new places.Autocomplete(inputEl, {
      types: ['(cities)'],
    })
    autocomplete.setFields(['address_components', 'geometry'])

    autocomplete.addListener('place_changed', () => {
      const place = autocomplete.getPlace()
      if (!place.geometry?.location || !place.address_components) {
        callbacksRef.current.onNoResults?.()
        return
      }

      let city = ''
      let state = ''
      let country = ''
      for (const component of place.address_components) {
        const type = component.types[0]
        if (type === 'locality') city = component.long_name
        else if (type === 'administrative_area_level_1')
          state = component.short_name
        else if (type === 'country') country = component.short_name
      }

      callbacksRef.current.onSelect({
        city,
        state,
        country,
        lat: place.geometry.location.lat(),
        lng: place.geometry.location.lng(),
      })
    })

    return () => {
      google.maps.event.clearInstanceListeners(autocomplete)
      document.querySelectorAll('.pac-container').forEach((el) => {
        if (!pacContainersBefore.has(el)) el.remove()
      })
    }
  }, [scriptLoaded])

  return (
    <>
      {apiKey && (
        <Script
          src={`https://maps.googleapis.com/maps/api/js?key=${encodeURIComponent(apiKey)}&libraries=places`}
          strategy="afterInteractive"
          onLoad={() => setScriptLoaded(true)}
        />
      )}
      <Input
        id={id}
        ref={inputRef}
        type="text"
        placeholder={placeholder}
        autoComplete="off"
      />
    </>
  )
}
