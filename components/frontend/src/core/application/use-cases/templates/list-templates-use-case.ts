import { inject, injectable } from 'inversify'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { TemplateDto } from '../../dto/template-dto'
import { TemplateMapper } from '../../mappers/template-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { PaginationDto } from '../../dto/pagination-dto'

export type ListTemplatesParams = {
  organizationId: string
  limit: number
  page: number
  outputFormat?: string
  name?: string
}

export type ListTemplates = {
  execute(params: ListTemplatesParams): Promise<PaginationDto<TemplateDto>>
}

@injectable()
export class ListTemplatesUseCase implements ListTemplates {
  constructor(
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(
    params: ListTemplatesParams
  ): Promise<PaginationDto<TemplateDto>> {
    // Fetch paginated templates
    const templates = await this.templateRepository.fetchAll({
      organizationId: params.organizationId,
      limit: params.limit,
      page: params.page,
      filters: {
        outputFormat: params.outputFormat as any,
        name: params.name
      }
    })

    // Map to DTO
    return TemplateMapper.toPaginationResponseDto(templates)
  }
}
