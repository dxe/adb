'use client'

import { useQuery } from '@tanstack/react-query'
import { Button } from 'components/ui/button'
import { API_PATH, apiClient } from 'lib/api'
import { sampleSize } from 'lodash-es'
import { Loader } from 'lucide-react'
import { useMemo } from 'react'
import toast from 'react-hot-toast'

export function ActivistNames() {
  // This useQuery could just as well happen in some deeper
  // child to <Activists>, data will be available immediately either way
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

/** Example component of how to fetch data using Tanstack Query. */
export default function Activists() {
  // TODO: add example of loading data from child relationships
  //   // This query was not prefetched on the server and will not start
  //   // fetching until on the client, both patterns are fine to mix.
  //   const { data: eventsData } = useQuery({
  //     queryKey: ['activists-events'],
  //     queryFn: getActivistsEvents,
  //   })

  return (
    <>
      <ActivistNames />
      <Button onClick={() => toast.success('Hey!')}>Click me</Button>
    </>
  )
}
