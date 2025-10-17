import { injectable, inject } from 'inversify'
import {
  TemplateEntity,
  TemplateSearchEntity
} from '@/core/domain/entities/template-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { ReporterHttpService } from '../services/reporter-http-service'
import { ReporterTemplateMapper } from '../mappers/reporter-template-mapper'
import { ReporterTemplateDto } from '../dto/reporter-template-dto'
import { createQueryString } from '@/lib/search'
import { ReporterPaginationDto } from '../dto/reporter-pagination-dto'

@injectable()
export class ReporterTemplateRepository implements TemplateRepository {
  constructor(
    @inject(ReporterHttpService)
    private readonly httpService: ReporterHttpService
  ) {}

  private baseUrl: string = '/v1/templates'

  async create(template: TemplateEntity): Promise<TemplateEntity> {
    const data = ReporterTemplateMapper.toCreateDto(template)

    const response = await this.httpService.postFormData<ReporterTemplateDto>(
      this.baseUrl,
      data,
      {
        headers: {
          'X-Organization-Id': template.organizationId
        }
      }
    )

    return {
      ...ReporterTemplateMapper.toEntity(response),
      organizationId: template.organizationId
    }
  }

  async fetchAll(
    organizationId: string,
    query: TemplateSearchEntity
  ): Promise<PaginationEntity<TemplateEntity>> {
    const queryParams: Record<string, any> = {
      limit: query.limit,
      page: query.page
    }

    if (query.outputFormat) {
      queryParams.outputFormat = query.outputFormat
    }
    if (query.name) {
      queryParams.description = query.name
    }
    if (query.createdAt) {
      queryParams.createdAt = query.createdAt
    }

    const response = await this.httpService.get<
      ReporterPaginationDto<ReporterTemplateDto>
    >(`${this.baseUrl}${createQueryString(queryParams)}`, {
      headers: {
        'X-Organization-Id': organizationId
      }
    })

    return ReporterTemplateMapper.toPaginationEntity(response)
  }

  async fetchById(id: string, organizationId: string): Promise<TemplateEntity> {
    const response = await this.httpService.get<ReporterTemplateDto>(
      `${this.baseUrl}/${id}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return {
      ...ReporterTemplateMapper.toEntity(response),
      organizationId
    }
  }

  async update(
    id: string,
    organizationId: string,
    template: Partial<TemplateEntity>
  ): Promise<TemplateEntity> {
    let response: ReporterTemplateDto

    if (template.templateFile) {
      const data = ReporterTemplateMapper.toUpdateDto(template)

      response = await this.httpService.patchFormData<ReporterTemplateDto>(
        `${this.baseUrl}/${id}`,
        data,
        {
          headers: {
            'X-Organization-Id': organizationId
          }
        }
      )
    } else {
      const dto = ReporterTemplateMapper.toUpdateDto(template)

      response = await this.httpService.patch<ReporterTemplateDto>(
        `${this.baseUrl}/${id}`,
        {
          body: JSON.stringify(dto),
          headers: {
            'X-Organization-Id': organizationId,
            'Content-Type': 'application/json'
          }
        }
      )
    }

    return {
      ...ReporterTemplateMapper.toEntity(response),
      organizationId
    }
  }

  async delete(id: string, organizationId: string): Promise<void> {
    await this.httpService.delete(`${this.baseUrl}/${id}`, {
      headers: {
        'X-Organization-Id': organizationId
      }
    })
  }
}
