import { createParser, type ParserMap } from 'nuqs/server'
import { getFilterSchemaEntry } from './filter-schema'
import type { FilterState } from './query-state'

function createFilterNuqsParser<K extends keyof FilterState>(key: K) {
  const entry = getFilterSchemaEntry(key)
  const parser = createParser<FilterState[K]>({
    parse: (raw) => entry.parseParam(raw) ?? null,
    serialize: (value) => entry.serializeParam(value) ?? '',
  })

  if (entry.defaultValue === undefined) {
    return parser
  }

  return parser.withDefault(entry.defaultValue as NonNullable<FilterState[K]>)
}

export const FILTER_NUQS_PARSERS = {
  searchAcrossChapters: createFilterNuqsParser('searchAcrossChapters'),
  nameSearch: createFilterNuqsParser('nameSearch'),
  includeHidden: createFilterNuqsParser('includeHidden'),
  lastEvent: createFilterNuqsParser('lastEvent'),
  interestDate: createFilterNuqsParser('interestDate'),
  firstEvent: createFilterNuqsParser('firstEvent'),
  totalEvents: createFilterNuqsParser('totalEvents'),
  totalInteractions: createFilterNuqsParser('totalInteractions'),
  activistLevel: createFilterNuqsParser('activistLevel'),
  source: createFilterNuqsParser('source'),
  training: createFilterNuqsParser('training'),
  assignedTo: createFilterNuqsParser('assignedTo'),
  followups: createFilterNuqsParser('followups'),
  prospect: createFilterNuqsParser('prospect'),
} satisfies ParserMap
