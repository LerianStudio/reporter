import { inject, injectable } from 'inversify'
import { type TemplateEntity } from '@/core/domain/entities/template-entity'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import {
  type CreateTemplateDto,
  type TemplateDto
} from '../../dto/template-dto'
import { TemplateMapper } from '../../mappers/template-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'

export type CreateTemplate = {
  execute(templateData: CreateTemplateDto): Promise<TemplateDto>
}

@injectable()
export class CreateTemplateUseCase implements CreateTemplate {
  constructor(
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(templateData: CreateTemplateDto): Promise<TemplateDto> {
    // Create template entity (file will be passed to external API)
    const templateEntity: TemplateEntity = {
      organizationId: templateData.organizationId,
      name: templateData.name,
      outputFormat: templateData.outputFormat,
      templateFile: templateData.templateFile
    }

    // Save template
    const createdTemplate = await this.templateRepository.create(templateEntity)

    // Return mapped DTO
    return TemplateMapper.toResponseDto(createdTemplate)
  }
}
