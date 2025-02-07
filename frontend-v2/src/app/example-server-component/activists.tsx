'use client'

import { useQuery } from '@tanstack/react-query'

export async function getActivists() {
  return Promise.resolve({
    activists: [
      { id: 1, name: 'Alice' },
      { id: 2, name: 'Bob' },
    ],
  })
}

export default function Activists() {
  // This useQuery could just as well happen in some deeper
  // child to <Activists>, data will be available immediately either way
  const { data } = useQuery({
    queryKey: ['activists'],
    queryFn: () => getActivists(),
  })

  // TODO: add example of loading data from child relationships
  //   // This query was not prefetched on the server and will not start
  //   // fetching until on the client, both patterns are fine to mix.
  //   const { data: eventsData } = useQuery({
  //     queryKey: ['activists-events'],
  //     queryFn: getActivistsEvents,
  //   })

  return (
    <>
      <h1>Activists</h1>
      <ul>
        {data?.activists.map((activist) => (
          <li key={activist.id}>{activist.name}</li>
        ))}
      </ul>
    </>
  )
}
