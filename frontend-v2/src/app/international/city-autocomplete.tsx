/// <reference types="google.maps" />
'use client'

import Script from 'next/script'
import {
  forwardRef,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
} from 'react'
import { Input } from '@/components/ui/input'

export interface CityValue {
  city: string
  state: string
  country: string
  lat: number
  lng: number
}

export interface CityAutocompleteHandle {
  focus: () => void
}

// Google Maps API key exposed to the browser at build time. Mirrors the
// server-side GOOGLE_PLACES_API_KEY used by the legacy Vue form (see
// server/src/config/config.go and server/templates/form_international.html),
// but must also be set as NEXT_PUBLIC_GOOGLE_PLACES_API_KEY in the
// frontend-v2 build/deploy environment for this to work.
const GOOGLE_PLACES_API_KEY = process.env.NEXT_PUBLIC_GOOGLE_PLACES_API_KEY

/**
 * Text input backed by the Google Places Autocomplete widget, restricted to
 * cities. This replaces the legacy Vue app's `vue-google-autocomplete`
 * component (frontend/external/vue-google-autocomplete). It replicates that
 * component's address-component parsing: `locality` -> city (long_name),
 * `administrative_area_level_1` -> state (short_name), `country` -> country
 * (short_name, matching DxE's mailing-list DB convention).
 */
export const CityAutocomplete = forwardRef<
  CityAutocompleteHandle,
  {
    id?: string
    placeholder?: string
    onSelect: (value: CityValue) => void
    onNoResults?: () => void
  }
>(function CityAutocomplete({ id, placeholder, onSelect, onNoResults }, ref) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [scriptLoaded, setScriptLoaded] = useState(false)

  useImperativeHandle(ref, () => ({
    focus: () => inputRef.current?.focus(),
  }))

  useEffect(() => {
    if (!scriptLoaded) return
    const inputEl = inputRef.current
    // `scriptLoaded` is only set once the Maps JS API's onLoad fires, so the
    // ambient `google` namespace (from @types/google.maps) is available here.
    const places =
      typeof google !== 'undefined' ? google.maps?.places : undefined
    if (!inputEl || !places) return

    const autocomplete = new places.Autocomplete(inputEl, {
      types: ['(cities)'],
    })
    autocomplete.setFields(['address_components', 'geometry'])

    const listener = autocomplete.addListener('place_changed', () => {
      const place = autocomplete.getPlace()
      if (!place.geometry?.location || !place.address_components) {
        onNoResults?.()
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

      onSelect({
        city,
        state,
        country,
        lat: place.geometry.location.lat(),
        lng: place.geometry.location.lng(),
      })
    })

    return () => listener.remove()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [scriptLoaded])

  return (
    <>
      {GOOGLE_PLACES_API_KEY && (
        <Script
          src={`https://maps.googleapis.com/maps/api/js?key=${GOOGLE_PLACES_API_KEY}&libraries=places`}
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
})
