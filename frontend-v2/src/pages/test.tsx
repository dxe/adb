import { AuthedPageLayout } from "@/components/AuthedPageLayout";
import { ContentWrapper } from "@/components/ContentWrapper";
import { VueNavbar } from "@/components/VueNavbar";

export default function TestPage() {
  return (
    <AuthedPageLayout>
      <VueNavbar pageName="TestPage" />
      <ContentWrapper size="sm">Hello from React!</ContentWrapper>
    </AuthedPageLayout>
  );
}
