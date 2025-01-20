import { useQuery } from "@tanstack/react-query";
import ky from "ky";
import { z } from "zod";

const StaticResourcesHashResp = z.object({
  hash: z.string(),
});

// This is only used for Vue components and should eventually be removed.
// It gets the "static resource hash" from the backend, which is a random
// hash generated whenever the server starts. It's a poor man's way
// to ensure that the frontend is always fetching the latest version
// of the Vue.js assets.
export const useStaticResourceHash = () => {
  const { data: staticResourceHash } = useQuery({
    queryKey: ["static_resources_hash"],
    queryFn: async () => {
      const resp = await ky.get("/static_resources_hash").json();
      return StaticResourcesHashResp.parse(resp);
    },
  });

  return staticResourceHash?.hash;
};
