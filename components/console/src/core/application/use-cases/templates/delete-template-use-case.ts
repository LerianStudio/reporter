import { inject, injectable } from 'inversify'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type DeleteTemplate = {
  execute(id: string, organizationId: string): Promise<void>
}

@injectable()
export class DeleteTemplateUseCase implements DeleteTemplate {
  constructor(
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(id: string, organizationId: string): Promise<void> {
    // Soft delete template
    await this.templateRepository.delete(id, organizationId)
  }
}
