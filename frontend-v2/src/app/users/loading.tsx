import { Loading } from '@/app/loading'

// app/loading.tsx may not be shown for navigation within the users/ segment, e.g. /users <--> /users/[id].
export default function UsersLoading() {
  return <Loading label="Loading page..." />
}
