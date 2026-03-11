import type { QueryActivistOptions } from '@/lib/api'
import type { FilterApiContext } from './filter-api-transform'
import * as apiTransform from './filter-api-transform'
import * as urlCodecs from './filter-url-codecs'
import type { FilterState } from './query-state'

type ApiFilters = QueryActivistOptions['filters']

type FilterSchemaEntry<K extends keyof FilterState> = {
  paramKey: string
  defaultValue: FilterState[K]
  isDefaultValue?: (value: FilterState[K]) => boolean
  parseParam: (raw: string | undefined) => FilterState[K]
  serializeParam: (value: FilterState[K]) => string | undefined
  toApi: (
    value: FilterState[K],
    context: FilterApiContext,
  ) => Partial<ApiFilters>
}

type AnyFilterSchemaEntry = FilterSchemaEntry<keyof FilterState>

export const FILTER_SCHEMA = {
  searchAcrossChapters: {
    paramKey: 'searchAcrossChapters',
    defaultValue: false,
    parseParam: urlCodecs.parseSearchAcrossChaptersParam,
    serializeParam: urlCodecs.serializeSearchAcrossChaptersParam,
    toApi: apiTransform.toApiSearchAcrossChapters,
  },
  nameSearch: {
    paramKey: 'nameSearch',
    defaultValue: '',
    parseParam: urlCodecs.parseNameSearchParam,
    serializeParam: urlCodecs.serializeNameSearchParam,
    toApi: apiTransform.toApiNameSearch,
  },
  includeHidden: {
    paramKey: 'includeHidden',
    defaultValue: false,
    parseParam: urlCodecs.parseIncludeHiddenParam,
    serializeParam: urlCodecs.serializeIncludeHiddenParam,
    toApi: apiTransform.toApiIncludeHidden,
  },
  lastEvent: {
    paramKey: 'lastEvent',
    defaultValue: undefined,
    parseParam: urlCodecs.parseDateRangeParam,
    serializeParam: urlCodecs.serializeDateRangeParam,
    toApi: apiTransform.toApiLastEvent,
  },
  interestDate: {
    paramKey: 'interestDate',
    defaultValue: undefined,
    parseParam: urlCodecs.parseDateRangeParam,
    serializeParam: urlCodecs.serializeDateRangeParam,
    toApi: apiTransform.toApiInterestDate,
  },
  firstEvent: {
    paramKey: 'firstEvent',
    defaultValue: undefined,
    parseParam: urlCodecs.parseDateRangeParam,
    serializeParam: urlCodecs.serializeDateRangeParam,
    toApi: apiTransform.toApiFirstEvent,
  },
  totalEvents: {
    paramKey: 'totalEvents',
    defaultValue: undefined,
    parseParam: urlCodecs.parseIntRangeParam,
    serializeParam: urlCodecs.serializeIntRangeParam,
    toApi: apiTransform.toApiTotalEvents,
  },
  totalInteractions: {
    paramKey: 'totalInteractions',
    defaultValue: undefined,
    parseParam: urlCodecs.parseIntRangeParam,
    serializeParam: urlCodecs.serializeIntRangeParam,
    toApi: apiTransform.toApiTotalInteractions,
  },
  activistLevel: {
    paramKey: 'level',
    defaultValue: undefined,
    parseParam: urlCodecs.parseActivistLevelParam,
    serializeParam: urlCodecs.serializeActivistLevelParam,
    toApi: apiTransform.toApiActivistLevel,
  },
  source: {
    paramKey: 'source',
    defaultValue: undefined,
    parseParam: urlCodecs.parseIncludeExcludeParam,
    serializeParam: urlCodecs.serializeIncludeExcludeParam,
    toApi: apiTransform.toApiSource,
  },
  training: {
    paramKey: 'training',
    defaultValue: undefined,
    parseParam: urlCodecs.parseIncludeExcludeParam,
    serializeParam: urlCodecs.serializeIncludeExcludeParam,
    toApi: apiTransform.toApiTraining,
  },
  assignedTo: {
    paramKey: 'assignedTo',
    defaultValue: undefined,
    parseParam: urlCodecs.parseAssignedToParam,
    serializeParam: urlCodecs.serializeAssignedToParam,
    toApi: apiTransform.toApiAssignedToFilter,
  },
  followups: {
    paramKey: 'followups',
    defaultValue: undefined,
    parseParam: urlCodecs.parseFollowupsParam,
    serializeParam: urlCodecs.serializeFollowupsParam,
    toApi: apiTransform.toApiFollowups,
  },
  prospect: {
    paramKey: 'prospect',
    defaultValue: undefined,
    parseParam: urlCodecs.parseProspectParam,
    serializeParam: urlCodecs.serializeProspectParam,
    toApi: apiTransform.toApiProspect,
  },
} satisfies { [K in keyof FilterState]: FilterSchemaEntry<K> }

export const FILTER_KEYS = Object.keys(FILTER_SCHEMA) as Array<
  keyof FilterState
>

const FILTER_SCHEMA_ANY = FILTER_SCHEMA as Record<
  keyof FilterState,
  AnyFilterSchemaEntry
>

export const getFilterSchemaEntry = <K extends keyof FilterState>(
  key: K,
): FilterSchemaEntry<K> =>
  FILTER_SCHEMA_ANY[key] as unknown as FilterSchemaEntry<K>

// URL building / parsing

export type FilterParamGetter = (key: string) => string | undefined

export const getFilterParamKey = (key: keyof FilterState): string =>
  FILTER_SCHEMA_ANY[key].paramKey

export const parseFilterParam = (
  key: keyof FilterState,
  raw: string | undefined,
): FilterState[keyof FilterState] => FILTER_SCHEMA_ANY[key].parseParam(raw)

export const serializeFilterParam = (
  key: keyof FilterState,
  value: FilterState[keyof FilterState],
): string | undefined => FILTER_SCHEMA_ANY[key].serializeParam(value)

export const FILTER_PARAM_KEYS = Object.fromEntries(
  FILTER_KEYS.map((key) => [key, getFilterParamKey(key)]),
) as Record<keyof FilterState, string>

export const parseFiltersFromParams = (
  getParam: FilterParamGetter,
): FilterState => {
  return Object.fromEntries(
    FILTER_KEYS.map((key) => [
      key,
      parseFilterParam(key, getParam(getFilterParamKey(key))),
    ]),
  ) as FilterState
}

type PartialFilterStateInput = {
  [K in keyof FilterState]?: FilterState[K] | null | undefined
}

export const normalizeFilterState = (
  filters: PartialFilterStateInput,
): FilterState =>
  Object.fromEntries(
    FILTER_KEYS.map((key) => [
      key,
      filters[key] ?? FILTER_SCHEMA_ANY[key].defaultValue,
    ]),
  ) as FilterState

export const isFilterValueDirty = <K extends keyof FilterState>(
  key: K,
  value: FilterState[K],
): boolean => {
  const entry = FILTER_SCHEMA_ANY[key]
  const isDefaultValue =
    entry.isDefaultValue ??
    ((candidate) => Object.is(candidate, entry.defaultValue))
  return !isDefaultValue(value)
}

export const isFilterStateDirty = (filters: FilterState): boolean =>
  FILTER_KEYS.some((key) => isFilterValueDirty(key, filters[key]))

export const buildFilterParamEntries = (
  filters: FilterState,
): [string, string | undefined][] =>
  FILTER_KEYS.map((key) => [
    getFilterParamKey(key),
    serializeFilterParam(key, filters[key]),
  ])

// API query building

export const mapFilterToApi = (
  key: keyof FilterState,
  value: FilterState[keyof FilterState],
  context: FilterApiContext,
): Partial<ApiFilters> => FILTER_SCHEMA_ANY[key].toApi(value, context)

export const buildApiFiltersFromState = (
  filters: FilterState,
  context: FilterApiContext,
): QueryActivistOptions['filters'] => {
  const apiFilters = {} as QueryActivistOptions['filters']

  for (const key of FILTER_KEYS) {
    Object.assign(apiFilters, mapFilterToApi(key, filters[key], context))
  }

  return apiFilters
}
