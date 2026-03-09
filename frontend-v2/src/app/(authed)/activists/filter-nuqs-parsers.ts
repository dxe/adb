import { createParser, parseAsBoolean, parseAsString } from 'nuqs'
import * as codecs from './filter-url-codecs'

export const parseAsSearchAcrossChapters = parseAsBoolean.withDefault(false)
export const parseAsNameSearch = parseAsString.withDefault('')
export const parseAsIncludeHidden = parseAsBoolean.withDefault(false)

export const parseAsDateRange = createParser({
  parse: (raw) => codecs.parseDateRangeParam(raw) ?? null,
  serialize: (val) => codecs.serializeDateRangeParam(val) ?? '',
})

export const parseAsIntRange = createParser({
  parse: (raw) => codecs.parseIntRangeParam(raw) ?? null,
  serialize: (val) => codecs.serializeIntRangeParam(val) ?? '',
})

export const parseAsIncludeExclude = createParser({
  parse: (raw) => codecs.parseIncludeExcludeParam(raw) ?? null,
  serialize: (val) => codecs.serializeIncludeExcludeParam(val) ?? '',
})

export const parseAsActivistLevel = createParser({
  parse: (raw) => codecs.parseActivistLevelParam(raw) ?? null,
  serialize: (val) => codecs.serializeActivistLevelParam(val) ?? '',
})

export const parseAsAssignedTo = createParser({
  parse: (raw) => codecs.parseAssignedToParam(raw) ?? null,
  serialize: (val) => codecs.serializeAssignedToParam(val) ?? '',
})

export const parseAsFollowups = createParser({
  parse: (raw) => codecs.parseFollowupsParam(raw) ?? null,
  serialize: (val) => codecs.serializeFollowupsParam(val) ?? '',
})

export const parseAsProspect = createParser({
  parse: (raw) => codecs.parseProspectParam(raw) ?? null,
  serialize: (val) => codecs.serializeProspectParam(val) ?? '',
})
