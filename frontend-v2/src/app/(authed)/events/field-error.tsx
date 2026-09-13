// Shared field-error message. Centralizes the error styling so it lives in one
// place. Some callers render it with an extra `mt-1`, so callers can opt into
// that spacing via `className`.
export const FieldError = ({
  message,
  className,
}: {
  message?: string
  className?: string
}) => {
  if (!message) return null
  return (
    <p className={`text-sm text-red-500${className ? ` ${className}` : ''}`}>
      {message}
    </p>
  )
}
