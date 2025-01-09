import { useQuery } from "@tanstack/react-query";
import ky from "ky";
import { z } from "zod";

const StaticResourcesHashResp = z.object({
  hash: z.string(),
});

// This is only used for Vue components and should eventually be removed.
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
