import { inject, injectable } from 'inversify'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { TemplateDto } from '../../dto/template-dto'
import { TemplateMapper } from '../../mappers/template-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type GetTemplate = {
  execute(id: string, organizationId: string): Promise<TemplateDto>
}

@injectable()
export class GetTemplateUseCase implements GetTemplate {
  constructor(
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(id: string, organizationId: string): Promise<TemplateDto> {
    // Fetch template by ID with organization validation
    const template = await this.templateRepository.fetchById(id, organizationId)

    // Map to DTO
    return TemplateMapper.toResponseDto(template)
  }
}
