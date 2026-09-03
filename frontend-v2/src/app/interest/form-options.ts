import { SF_BAY_CHAPTER_ID } from '@/lib/constants'

export interface InterestFormOptions {
  chapterId: number
  formName: string
  formTitle: string
  formDescription: string
  showInterests: boolean
  showReferralFriends: boolean
  showReferralApply: boolean
  showReferralOutlet: boolean
  initialReferralApply: string
}

export function parseInterestFormOptions(
  searchParams: URLSearchParams,
): InterestFormOptions {
  const chapterIdStr = searchParams.get('chapterId')
  const parsedChapterId =
    chapterIdStr != null ? parseInt(chapterIdStr, 10) : NaN
  const chapterId = !Number.isNaN(parsedChapterId)
    ? parsedChapterId
    : SF_BAY_CHAPTER_ID
  const formName = searchParams.get('name') || 'Interest Form'
  const formTitle =
    searchParams.get('title') ||
    (parsedChapterId === SF_BAY_CHAPTER_ID
      ? 'DxE SF Bay - Get Involved'
      : 'Direct Action Everywhere - Get Involved')
  const formDescription = searchParams.get('description') || ''
  // showInterests defaults true to preserve behavior of links generated
  // before visibility of this section was configurable
  const showInterests = searchParams.get('showInterests') !== 'false'
  const showReferralFriends = searchParams.get('showReferralFriends') === 'true'
  const showReferralApply = searchParams.get('showReferralApply') === 'true'
  const showReferralOutlet = searchParams.get('showReferralOutlet') === 'true'
  const referralApplyParam = searchParams.get('referralApply') || ''
  const initialReferralApply =
    referralApplyParam !== 'null' ? referralApplyParam : ''

  return {
    chapterId,
    formName,
    formTitle,
    formDescription,
    showInterests,
    showReferralFriends,
    showReferralApply,
    showReferralOutlet,
    initialReferralApply,
  }
}
