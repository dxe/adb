'use client'

import {
  PlacesAutocomplete,
  PlaceValue,
} from '@/components/places-autocomplete'

export interface CityValue {
  city: string
  state: string
  country: string
  lat: number
  lng: number
}

const CITY_FIELDS = [
  'place_id',
  'address_components',
  'geometry',
  'formatted_address',
]

function parseCityFromPlace(place: PlaceValue): CityValue | null {
  if (
    !place.address_components ||
    place.lat === undefined ||
    place.lng === undefined
  ) {
    return null
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

  return { city, state, country, lat: place.lat, lng: place.lng }
}

/** City-restricted Places autocomplete emitting parsed city/state/country/lat/lng. */
export function CityAutocomplete({
  id,
  placeholder,
  apiKey,
  hasError,
  onSelect,
  onNoResults,
  onUnavailable,
}: {
  id?: string
  placeholder?: string
  apiKey: string
  hasError?: boolean
  onSelect: (value: CityValue) => void
  onNoResults?: () => void
  /** Called when the Places script fails to load, disabling the field. */
  onUnavailable?: () => void
}) {
  return (
    <PlacesAutocomplete
      id={id}
      apiKey={apiKey}
      value=""
      placeholder={placeholder}
      types={['(cities)']}
      fields={CITY_FIELDS}
      loadErrorMessage="Location search failed to load. Please try again later."
      hasError={hasError}
      onSelect={(place) => {
        const city = parseCityFromPlace(place)
        if (city) {
          onSelect(city)
        } else {
          onNoResults?.()
        }
      }}
      onClear={() => onNoResults?.()}
      onNoResult={() => onNoResults?.()}
      onUnavailable={onUnavailable}
    />
  )
}
