'use client'

import { Button } from '@/components/ui/button'

/**
 * Informational copy shown before the applicant starts filling out the
 * form. Text content is identical to frontend/FormApply.vue.
 */
export function ChapterMemberInfo({ onApply }: { onApply: () => void }) {
  return (
    <div className="flex flex-col gap-4">
      <p>
        Are you ready to change the world for animals alongside an incredible
        community of like-minded individuals? Then you&rsquo;re in the right
        place! Decades of research into historic movements &ndash; like the
        Civil Rights Movement or Gay Rights Movement &ndash; tells us that the
        most important factor in creating social and political change is
        sustained participation in nonviolent direct action. We also know that
        sustained participation is more likely when activists are embedded in a
        strong, supportive community. When you become a chapter member,
        you&rsquo;re committing to being an integral part of this activist
        community and to attending direct action events (like a protest,
        outreach event, or animal care event) consistently.
      </p>

      <h2 className="text-xl font-semibold">How to become a Chapter Member</h2>
      <ul className="list-disc space-y-1 pl-5">
        <li>
          Attend DxE&rsquo;s{' '}
          <a
            href="http://dxe.io/workshop"
            target="_blank"
            rel="noreferrer"
            className="text-primary underline"
          >
            intro workshop
          </a>{' '}
          (held monthly) to learn about our community and theory of change
        </li>
        <li>Attend a direct action event</li>
        <li>
          Watch a video and{' '}
          <a
            href="https://dxe.io/refresherquiz"
            target="_blank"
            rel="noreferrer"
            className="text-primary underline"
          >
            take a quiz
          </a>{' '}
          on consent
        </li>
      </ul>

      <h3 className="text-lg font-semibold">
        Benefits of becoming a Chapter Member
      </h3>
      <ul className="list-disc space-y-1 pl-5">
        <li>
          Voting rights (after sustained participation in the chapter) on
          chapter decisions
        </li>
        <li>
          Access to chapter member mailing lists, Signal alerts chat, and
          Facebook group
        </li>
        <li>
          Added to a small community group with other chapter members in your
          neighborhood/area
        </li>
        <li>Invites to special events and opportunities</li>
      </ul>

      <h3 className="text-lg font-semibold">Expectations of Chapter Members</h3>
      <ul className="list-disc space-y-1 pl-5">
        <li>
          Uphold our{' '}
          <a
            href="https://dxe.io/conduct"
            target="_blank"
            rel="noreferrer"
            className="text-primary underline"
          >
            code of conduct
          </a>{' '}
          and{' '}
          <a
            href="https://www.directactioneverywhere.com/core-values"
            target="_blank"
            rel="noreferrer"
            className="text-primary underline"
          >
            values
          </a>
        </li>
        <li>Attend direct action events consistently</li>
        <li>
          Failure to uphold our code of conduct, values, or attend direct action
          events consistently may result in being removed as a chapter member
        </li>
      </ul>

      <Button type="button" onClick={onApply} className="w-fit">
        Apply now
      </Button>
    </div>
  )
}
