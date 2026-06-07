'use client'

import { useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import toast from 'react-hot-toast'
import { Loader2, Send } from 'lucide-react'
import { apiClient } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export default function ConfigPage() {
  const [email, setEmail] = useState('')

  const mutation = useMutation({
    mutationFn: (to: string) => apiClient.sendTestEmail(to),
    onSuccess: (_data, to) => {
      toast.success(`Test email sent to ${to}`)
    },
    onError: (err: unknown) => {
      const message =
        err instanceof Error ? err.message : 'Failed to send test email'
      toast.error(message)
    },
  })

  const trimmedEmail = email.trim()

  const handleSend = () => {
    if (!trimmedEmail) {
      toast.error('Enter an email address')
      return
    }
    mutation.mutate(trimmedEmail)
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-2xl font-semibold">Configuration</h1>
        <p className="text-muted-foreground text-sm">
          Manage system settings and run diagnostics.
        </p>
      </div>

      <section className="flex flex-col gap-4 rounded-lg border p-6">
        <div className="flex flex-col gap-1">
          <h2 className="text-lg font-semibold">SMTP</h2>
          <p className="text-muted-foreground text-sm">
            Send a test email to verify that outbound email is working.
          </p>
        </div>

        <form
          className="flex flex-col gap-3 sm:flex-row sm:items-end"
          onSubmit={(e) => {
            e.preventDefault()
            handleSend()
          }}
        >
          <div className="flex flex-1 flex-col gap-1">
            <Label htmlFor="test-email">Recipient email address</Label>
            <Input
              id="test-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              disabled={mutation.isPending}
            />
          </div>
          <Button type="submit" disabled={mutation.isPending || !trimmedEmail}>
            {mutation.isPending ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Send className="h-4 w-4" />
            )}
            Send test email
          </Button>
        </form>
      </section>
    </div>
  )
}
