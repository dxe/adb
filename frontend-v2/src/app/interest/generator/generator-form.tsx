'use client'

import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useQuery } from '@tanstack/react-query'
import { API_PATH, apiClient } from '@/lib/api'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@/components/ui/form'
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

export default function GeneratorForm(props: { adbRootUrl: string }) {
  // adbRootUrl may be provided by parent for testing purposes.
  const adbRootUrl =
    props.adbRootUrl ?? `${window.location.protocol}//${window.location.host}`

  const { data: chapterList, isLoading: isChapterListLoading } = useQuery({
    queryKey: [API_PATH.CHAPTER_LIST],
    queryFn: apiClient.getChapterList,
  })

  const form = useForm<FormValues>({
    defaultValues: {
      name: '',
      title: '',
      description: '',
      chapterId: '',
      activismInterests: false,
      referralApplyName: '',
      referralFriends: false,
      referralApply: false,
      referralOutlet: false,
    },
  })

  const [output, setOutput] = useState('')

  const onSubmit = (data: FormValues) => {
    const params = new URLSearchParams()
    params.append('name', data.name)
    params.append('title', data.title)
    params.append('description', data.description)
    params.append('chapterId', data.chapterId)
    params.append('showInterests', data.activismInterests.toString())
    params.append(
      'referralApply',
      data.referralApplyName === '' ? 'null' : data.referralApplyName,
    )
    params.append('showReferralFriends', data.referralFriends.toString())
    params.append('showReferralApply', data.referralApply.toString())
    params.append('showReferralOutlet', data.referralOutlet.toString())

    const url = `${adbRootUrl}/interest?${params.toString()}`
    setOutput(url)
  }

  return (
    <div className="w-full max-w-2xl mx-auto p-8 bg-white rounded shadow">
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <FormField
            control={form.control}
            name="name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Form name (visible in ADB)</FormLabel>
                <FormControl>
                  <Input type="text" placeholder="Name" {...field} />
                </FormControl>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="title"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Form title</FormLabel>
                <FormControl>
                  <Input type="text" placeholder="Title" {...field} />
                </FormControl>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="description"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Form description</FormLabel>
                <FormControl>
                  <Textarea placeholder="Description" {...field} />
                </FormControl>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="chapterId"
            rules={{ required: true }}
            render={({ field }) => (
              <FormItem>
                <FormLabel>
                  Chapter <span className="text-red-500">*</span>
                </FormLabel>

                <Select
                  value={field.value}
                  onValueChange={field.onChange}
                  disabled={isChapterListLoading}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Select a chapter" />
                    </SelectTrigger>
                  </FormControl>
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
                <FormMessage>
                  {form.formState.errors.chapterId && 'Chapter is required.'}
                </FormMessage>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="activismInterests"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center gap-2 space-y-0">
                <FormControl>
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
                <FormLabel>
                  Show "What are your activism interests, if any?"
                </FormLabel>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="referralApply"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center gap-2 space-y-0">
                <FormControl>
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
                <FormLabel>Show "Who encouraged you to sign up?"</FormLabel>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="referralApplyName"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Pre-fill "Who encouraged you to sign up?"</FormLabel>
                <FormControl>
                  <Input type="text" placeholder="Referral" {...field} />
                </FormControl>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="referralFriends"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center gap-2 space-y-0">
                <FormControl>
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
                <FormLabel>
                  Show "List any existing DxE activists who you are close
                  friends with"
                </FormLabel>
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="referralOutlet"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center gap-2 space-y-0">
                <FormControl>
                  <Checkbox
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
                <FormLabel>
                  Show "Where did you hear about this opportunity to get
                  involved in DxE?"
                </FormLabel>
              </FormItem>
            )}
          />

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

          <FormItem>
            <FormLabel>URL</FormLabel>
            <FormControl>
              <Textarea readOnly value={output} />
            </FormControl>
          </FormItem>
        </form>
      </Form>
    </div>
  )
}
