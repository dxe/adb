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

const isMapsPlacesLoaded = () =>
  typeof google !== 'undefined' && !!google.maps?.places

// Name forms (long vs short) mirror the legacy vue-google-autocomplete parsing.
function parseCityFromPlace(
  place: google.maps.places.PlaceResult,
): CityValue | null {
  if (!place.geometry?.location || !place.address_components) return null

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

  return {
    city,
    state,
    country,
    lat: place.geometry.location.lat(),
    lng: place.geometry.location.lng(),
  }
}

/** City-restricted Google Places Autocomplete input. With no `apiKey` (yet),
 *  the input still works but offers no suggestions. */
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
  // next/script won't re-fire onLoad if the Maps script is already loaded.
  const [scriptLoaded, setScriptLoaded] = useState(isMapsPlacesLoaded)

  // Ref keeps the widget-building effect from re-running on parent re-renders.
  const callbacksRef = useRef({ onSelect, onNoResults })
  useEffect(() => {
    callbacksRef.current = { onSelect, onNoResults }
  })

  useEffect(() => {
    const inputEl = inputRef.current
    if (!scriptLoaded || !inputEl || !isMapsPlacesLoaded()) return

    // Strict mode's double-invoked effect would otherwise leave an orphaned
    // duplicate suggestion dropdown (.pac-container) on document.body.
    const pacContainersBefore = new Set(
      document.querySelectorAll('.pac-container'),
    )

    const autocomplete = new google.maps.places.Autocomplete(inputEl, {
      types: ['(cities)'],
    })
    autocomplete.setFields(['address_components', 'geometry'])

    autocomplete.addListener('place_changed', () => {
      const value = parseCityFromPlace(autocomplete.getPlace())
      if (value) {
        callbacksRef.current.onSelect(value)
      } else {
        callbacksRef.current.onNoResults?.()
      }
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
