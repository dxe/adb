import { useQuery } from "@tanstack/react-query";
import ky from "ky";
import { useMemo } from "react";
import { z } from "zod";

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
      .array(
        z.object({
          Role: z.enum(["admin", "organizer", "attendance", "non-sfbay"]),
        })
      )
      .transform((roles) => roles.map((it) => it.Role)),
  }),
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
        };
      }
    },
  });

  const highestRole = useMemo(() => {
    const roles = query.data?.user?.Roles;
    return roles?.includes("admin")
      ? "admin"
      : roles?.includes("organizer")
        ? "organizer"
        : roles?.includes("attendance")
          ? "attendance"
          : roles?.includes("non-sfbay")
            ? "non-sfbay"
            : undefined;
  }, [query.data?.user?.Roles]);

  return {
    user: {
      ...query.data?.user,
      role: highestRole,
    },
    isLoading: query.isLoading,
  };
};
