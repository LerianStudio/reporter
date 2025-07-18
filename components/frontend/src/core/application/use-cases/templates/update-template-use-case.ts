import { inject, injectable } from 'inversify'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import {
  type UpdateTemplateDto,
  type TemplateDto
} from '../../dto/template-dto'
import { TemplateMapper } from '../../mappers/template-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type UpdateTemplate = {
  execute(
    id: string,
    organizationId: string,
    templateData: UpdateTemplateDto
  ): Promise<TemplateDto>
}

@injectable()
export class UpdateTemplateUseCase implements UpdateTemplate {
  constructor(
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(
    id: string,
    organizationId: string,
    templateData: UpdateTemplateDto
  ): Promise<TemplateDto> {
    // Update template
    const updatedTemplate = await this.templateRepository.update(
      id,
      organizationId,
      templateData
    )

    // Map to DTO
    return TemplateMapper.toResponseDto(updatedTemplate)
  }
}
