import { inject, injectable } from 'inversify'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import type {
  TemplateDto,
  TemplateSearchParamDto
} from '../../dto/template-dto'
import { TemplateMapper } from '../../mappers/template-mapper'
import { LogOperation } from '@/core/infrastructure/logger/decorators/log-operation'
import { PaginationDto } from '../../dto/pagination-dto'

export type ListTemplates = {
  execute(
    organizationId: string,
    query: TemplateSearchParamDto
  ): Promise<PaginationDto<TemplateDto>>
}

@injectable()
export class ListTemplatesUseCase implements ListTemplates {
  constructor(
    @inject(TemplateRepository)
    private readonly templateRepository: TemplateRepository
  ) {}

  @LogOperation({ layer: 'application' })
  async execute(
    organizationId: string,
    query: TemplateSearchParamDto
  ): Promise<PaginationDto<TemplateDto>> {
    const templates = await this.templateRepository.fetchAll(
      organizationId,
      query
    )

    return TemplateMapper.toPaginationResponseDto(templates)
  }
}
