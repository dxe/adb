'use client'

import { useState } from 'react'
import { useForm } from '@tanstack/react-form'
import { useQuery } from '@tanstack/react-query'
import { API_PATH, apiClient } from '@/lib/api'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from '@/components/ui/select'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { z } from 'zod'

const formSchema = z.object({
  name: z.string(),
  title: z.string(),
  description: z.string(),
  chapterId: z.string().min(1, 'Chapter is required.'),
  activismInterests: z.boolean(),
  referralApplyName: z.string(),
  referralFriends: z.boolean(),
  referralApply: z.boolean(),
  referralOutlet: z.boolean(),
})

type FormValues = z.infer<typeof formSchema>

const defaultValues: FormValues = {
  name: '',
  title: '',
  description: '',
  chapterId: '',
  activismInterests: false,
  referralApplyName: '',
  referralFriends: false,
  referralApply: false,
  referralOutlet: false,
}

export default function GeneratorForm(props: { adbRootUrl?: string }) {
  const { data: chapterList } = useQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: ({ signal }) => apiClient.getChapterList(signal),
  })

  const [output, setOutput] = useState('')

  const form = useForm({
    defaultValues,
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: ({ value }) => {
      // adbRootUrl may be provided by parent for testing purposes.
      const adbRootUrl =
        props.adbRootUrl ??
        `${window.location.protocol}//${window.location.host}`
      const params = new URLSearchParams()
      params.append('name', value.name)
      params.append('title', value.title)
      params.append('description', value.description)
      params.append('chapterId', value.chapterId)
      params.append('showInterests', value.activismInterests.toString())
      params.append(
        'referralApply',
        value.referralApplyName === '' ? 'null' : value.referralApplyName,
      )
      params.append('showReferralFriends', value.referralFriends.toString())
      params.append('showReferralApply', value.referralApply.toString())
      params.append('showReferralOutlet', value.referralOutlet.toString())

      const url = `${adbRootUrl}/interest?${params.toString()}`
      setOutput(url)
    },
  })

  return (
    <div className="w-full max-w-2xl mx-auto p-8 bg-white rounded shadow">
      <form
        onSubmit={(e) => {
          e.preventDefault()
          form.handleSubmit()
        }}
        className="space-y-6"
      >
        <form.Field name="name">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="gen-name">Form name (visible in ADB)</Label>
              <Input
                id="gen-name"
                type="text"
                placeholder="Name"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
            </div>
          )}
        </form.Field>

        <form.Field name="title">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="gen-title">Form title</Label>
              <Input
                id="gen-title"
                type="text"
                placeholder="Title"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
            </div>
          )}
        </form.Field>

        <form.Field name="description">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="gen-description">Form description</Label>
              <Textarea
                id="gen-description"
                placeholder="Description"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
            </div>
          )}
        </form.Field>

        <form.Field name="chapterId">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="gen-chapter">
                Chapter <span className="text-red-500">*</span>
              </Label>
              <Select
                value={field.state.value}
                onValueChange={(value) => field.handleChange(value)}
              >
                <SelectTrigger id="gen-chapter">
                  <SelectValue placeholder="Select a chapter" />
                </SelectTrigger>
                <SelectContent>
                  {chapterList?.map((chapter) => (
                    <SelectItem
                      key={chapter.ChapterID}
                      value={chapter.ChapterID.toString()}
                    >
                      {chapter.Name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              {field.state.meta.errors[0] && (
                <p className="text-sm text-destructive">
                  {field.state.meta.errors[0]?.message}
                </p>
              )}
            </div>
          )}
        </form.Field>

        <form.Field name="activismInterests">
          {(field) => (
            <Label className="flex flex-row items-center gap-2">
              <Checkbox
                checked={field.state.value}
                onCheckedChange={(checked) =>
                  field.handleChange(Boolean(checked))
                }
              />
              Show &quot;What are your activism interests, if any?&quot;
            </Label>
          )}
        </form.Field>

        <form.Field name="referralApply">
          {(field) => (
            <Label className="flex flex-row items-center gap-2">
              <Checkbox
                checked={field.state.value}
                onCheckedChange={(checked) =>
                  field.handleChange(Boolean(checked))
                }
              />
              Show &quot;Who encouraged you to sign up?&quot;
            </Label>
          )}
        </form.Field>

        <form.Field name="referralApplyName">
          {(field) => (
            <div className="space-y-1">
              <Label htmlFor="gen-referral-name">
                Pre-fill &quot;Who encouraged you to sign up?&quot;
              </Label>
              <Input
                id="gen-referral-name"
                type="text"
                placeholder="Referral"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
            </div>
          )}
        </form.Field>

        <form.Field name="referralFriends">
          {(field) => (
            <Label className="flex flex-row items-center gap-2">
              <Checkbox
                checked={field.state.value}
                onCheckedChange={(checked) =>
                  field.handleChange(Boolean(checked))
                }
              />
              Show &quot;List any existing DxE activists who you are close
              friends with&quot;
            </Label>
          )}
        </form.Field>

        <form.Field name="referralOutlet">
          {(field) => (
            <Label className="flex flex-row items-center gap-2">
              <Checkbox
                checked={field.state.value}
                onCheckedChange={(checked) =>
                  field.handleChange(Boolean(checked))
                }
              />
              Show &quot;Where did you hear about this opportunity to get
              involved in DxE?&quot;
            </Label>
          )}
        </form.Field>

        <div className="flex items-center gap-2">
          <Button type="submit" className="w-1/2">
            Generate URL
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              navigator.clipboard.writeText(output)
              alert('Copied URL')
            }}
            disabled={!output}
            className="w-1/2"
          >
            Copy
          </Button>
        </div>

        <div>
          <Label htmlFor="generated-url">URL</Label>
          <Textarea id="generated-url" readOnly value={output} />
        </div>
      </form>
    </div>
  )
}
