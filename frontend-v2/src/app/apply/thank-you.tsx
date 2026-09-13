export function ThankYou() {
  return (
    <div className="flex flex-col gap-2">
      <h2 className="text-xl font-semibold">Thank you!</h2>
      <p>
        An organizer from our Development Team will be reaching out to you about
        next steps and how you can get more involved. If you&rsquo;d like to see
        our upcoming events, please go to{' '}
        <a
          href="https://dxe.io/events"
          className="text-primary underline"
          target="_blank"
          rel="noreferrer"
        >
          dxe.io/events
        </a>
        .
      </p>
    </div>
  )
}
