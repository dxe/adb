import { ContentWrapper } from '@/app/content-wrapper'
import GeneratorForm from './generator-form'

export default function InterestGeneratorPage() {
  return (
    <ContentWrapper size="sm" className="gap-6">
      <h1 className="text-lg">Interest Form Generator</h1>
      <GeneratorForm />
    </ContentWrapper>
  )
}
