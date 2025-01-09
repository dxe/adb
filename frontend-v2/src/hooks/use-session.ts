import { useQuery } from "@tanstack/react-query";
import ky from "ky";
import { z } from "zod";

const Role = z.enum(["admin", "organizer", "attendance", "non-sfbay"]);

const AuthedUserRespSchema = z.object({
  user: z.object({
    Admin: z.boolean(),
    ChapterID: z.number(),
    ChapterName: z.string(),
    Disabled: z.boolean(),
    Email: z.string(),
    ID: z.number(),
    Name: z.string(),
    Roles: z
      .array(z.object({ Role: Role }))
      .transform((roles) => roles.map((it) => it.Role)),
  }),
  mainRole: Role,
});

export const useSession = () => {
  const query = useQuery({
    queryKey: ["user.me"],
    queryFn: async () => {
      try {
        const resp = await ky.get("/user/me").json();
        return AuthedUserRespSchema.parse(resp);
      } catch (err) {
        console.error(`Error fetching authed user: ${err}`);
        return {
          user: null,
          mainRole: null,
        };
      }
    },
  });

  return {
    user: {
      ...query.data?.user,
      role: query.data?.mainRole,
    },
    isLoading: query.isLoading,
  };
};
