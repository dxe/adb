import { AuthedPageLayout } from "@/components/AuthedPageLayout";
import { ContentWrapper } from "@/components/ContentWrapper";
import { VueNavbar } from "@/components/VueNavbar";
import { useQuery } from "@tanstack/react-query";
import ky from "ky";
import { useMemo } from "react";
import { z } from "zod";
import { sampleSize } from "lodash-es";
import toast from "react-hot-toast";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import NiceModal, { useModal } from "@ebay/nice-modal-react";

export default function TestPage() {
  return (
    <AuthedPageLayout>
      <VueNavbar pageName="TestPage" />
      <ContentWrapper size="sm" className="flex flex-col gap-6">
        <p>Hello from React!</p>
        <ActivistNames />
        <Button onClick={() => toast.success("Hey!")}>Click me</Button>
        <Button variant="outline" onClick={() => NiceModal.show(ExampleDialog)}>
          Show dialog
        </Button>
      </ContentWrapper>
    </AuthedPageLayout>
  );
}

const ActivistNames = () => {
  const { data } = useQuery({
    queryKey: ["activist_names.get"],
    queryFn: async () => {
      const resp = await ky.get("/activist_names/get").json();
      return z.object({ activist_names: z.array(z.string()) }).parse(resp);
    },
  });

  const sampledActivists = useMemo(() => {
    return sampleSize(data?.activist_names ?? [], 25);
  }, [data?.activist_names]);

  return (
    <div>
      <p className="font-bold">Here are some activists:</p>
      <ul className="list-disc pl-4">
        {sampledActivists.map((name) => (
          <li key={name}>{name}</li>
        ))}
      </ul>
    </div>
  );
};

const ExampleDialog = NiceModal.create(() => {
  const modal = useModal();

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
  );
});
