/** Minimal structural member shape satisfied by both `WorkingGroup` and circle members. */
export interface MemberLike {
  name: string
  point_person: boolean
  non_member_on_mailing_list: boolean
}

/** Returns the member flagged as point person (host), if any — there should only ever be one. */
export function findPointPerson<M extends Pick<MemberLike, 'point_person'>>(
  members: M[],
): M | undefined {
  return members.find((m) => m.point_person)
}

/** Counts members on the mailing list: includes the point person, excludes `non_member_on_mailing_list` entries. */
export function countMailingListMembers(
  members: Pick<MemberLike, 'non_member_on_mailing_list'>[],
): number {
  return members.filter((m) => !m.non_member_on_mailing_list).length
}
