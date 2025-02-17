import { API_PATH, apiClient } from 'lib/api'
import { QueryClient } from '@tanstack/react-query'

// This is only used for Vue components and should eventually be removed.
// It gets the "static resource hash" from the backend, which is a random
// hash generated whenever the server starts. It's a poor man's way
// to ensure that the frontend is always fetching the latest version
// of the Vue.js assets.
export const fetchStaticResourceHash = async () => {
  const queryClient = new QueryClient()
  const staticResourceHash = await queryClient.fetchQuery({
    queryKey: [API_PATH.STATIC_RESOURCE_HASH],
    queryFn: apiClient.getStaticResourceHash,
  })

  return staticResourceHash?.hash
}
