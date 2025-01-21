import { AuthedPageLayout } from '@/components/AuthedPageLayout'
import { ContentWrapper } from '@/components/ContentWrapper'
import { VueNavbar } from '@/components/VueNavbar'
import { HydrationBoundary, useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { sampleSize } from 'lodash-es'
import toast from 'react-hot-toast'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import NiceModal, { useModal } from '@ebay/nice-modal-react'
import { API_PATH, apiClient } from '@/lib/api'
import { DefaultPageProps, getDefaultServerSideProps } from '@/lib/ssr'
import { Loader } from 'lucide-react'

export default function TestPage({ dehydratedState }: DefaultPageProps) {
  return (
    <HydrationBoundary state={dehydratedState}>
      <AuthedPageLayout>
        <VueNavbar pageName="TestPage" />
        <ContentWrapper size="sm" className="gap-6">
          <p>Hello from React!</p>
          <ActivistNames />
          <Button onClick={() => toast.success('Hey!')}>Click me</Button>
          <Button
            variant="outline"
            onClick={() => NiceModal.show(ExampleDialog)}
          >
            Show dialog
          </Button>
        </ContentWrapper>
      </AuthedPageLayout>
    </HydrationBoundary>
  )
}

export const getServerSideProps = getDefaultServerSideProps

/** Example component of how to fetch data using Tanstack Query. */
const ActivistNames = () => {
  const { data, isLoading } = useQuery({
    queryKey: [API_PATH.ACTIVIST_NAMES_GET],
    queryFn: apiClient.getActivistNames,
  })

  const sampledActivists = useMemo(() => {
    return sampleSize(data?.activist_names ?? [], 25)
  }, [data?.activist_names])

  return (
    <div>
      <p className="font-bold">Here are some activists:</p>
      <ul className="list-disc pl-4">
        {isLoading ? (
          <Loader className="animate-spin" />
        ) : (
          sampledActivists.map((name) => <li key={name}>{name}</li>)
        )}
      </ul>
    </div>
  )
}

/** Example component of how to use a modal. */
const ExampleDialog = NiceModal.create(() => {
  const modal = useModal()

  return (
    <Dialog
      open={modal.visible}
      onOpenChange={(prev) => (!prev ? modal.remove() : modal.show())}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>This is a dialog</DialogTitle>
          <DialogDescription>Hello again.</DialogDescription>
        </DialogHeader>
      </DialogContent>
    </Dialog>
  )
})
