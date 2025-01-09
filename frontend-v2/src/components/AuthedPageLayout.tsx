import { ReactNode, useEffect } from "react";
import { useSession } from "@/hooks/use-session";

export const AuthedPageLayout = (props: { children: ReactNode }) => {
  const session = useSession();

  useEffect(() => {
    if (session.isLoading) {
      return;
    }
    if (!session.user) {
      window.location.pathname = "/login";
    }
  }, [session.isLoading, session.user]);

  return session.isLoading ? null : props.children;
};
