import {
  sortFn_alphanumeric,
  sortFn_basic,
  sortFn_datetime,
  sortFn_text,
} from '@tanstack/react-table'

export const AUTO_SORT_FNS = {
  alphanumeric: sortFn_alphanumeric,
  basic: sortFn_basic,
  datetime: sortFn_datetime,
  text: sortFn_text,
}
