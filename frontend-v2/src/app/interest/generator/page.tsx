import { ContentWrapper } from '@/app/content-wrapper'
import { AuthedPageLayout } from '@/app/authed-page-layout'
import { Navbar } from '@/components/nav'
import GeneratorForm from './generator-form'

export default async function InterestGeneratorPage() {
  return (
    <AuthedPageLayout pageName="InterestFormGenerator">
      <Navbar />
      <ContentWrapper size="sm" className="gap-6">
        <h1 className="text-lg">Interest Form Generator</h1>
        <GeneratorForm />
      </ContentWrapper>
    </AuthedPageLayout>
  )
}
