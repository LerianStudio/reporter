import { injectable, inject } from 'inversify'
import {
  TemplateEntity,
  TemplateSearchEntity
} from '@/core/domain/entities/template-entity'
import { PaginationEntity } from '@/core/domain/entities/pagination-entity'
import { TemplateRepository } from '@/core/domain/repositories/template-repository'
import { SmartTemplatesHttpService } from '../services/smart-templates-http-service'
import { SmartTemplateMapper } from '../mappers/smart-template-mapper'
import { SmartTemplateDto } from '../dto/smart-template-dto'
import { createQueryString } from '@/lib/search'
import { SmartPaginationDto } from '../dto/smart-pagination-dto'

@injectable()
export class SmartTemplateRepository implements TemplateRepository {
  constructor(
    @inject(SmartTemplatesHttpService)
    private readonly httpService: SmartTemplatesHttpService
  ) {}

  private baseUrl: string = '/v1/templates'

  async create(template: TemplateEntity): Promise<TemplateEntity> {
    const data = SmartTemplateMapper.toCreateDto(template)

    const response = await this.httpService.postFormData<SmartTemplateDto>(
      this.baseUrl,
      data,
      {
        headers: {
          'X-Organization-Id': template.organizationId
        }
      }
    )

    return {
      ...SmartTemplateMapper.toEntity(response),
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

    // Add filters to query params
    if (query.outputFormat) {
      queryParams.outputFormat = query.outputFormat
    }
    if (query.name) {
      queryParams.description = query.name
    }

    const response = await this.httpService.get<
      SmartPaginationDto<SmartTemplateDto>
    >(`${this.baseUrl}${createQueryString(queryParams)}`, {
      headers: {
        'X-Organization-Id': organizationId
      }
    })

    return SmartTemplateMapper.toPaginationEntity(response)
  }

  async fetchById(id: string, organizationId: string): Promise<TemplateEntity> {
    const response = await this.httpService.get<SmartTemplateDto>(
      `${this.baseUrl}/${id}`,
      {
        headers: {
          'X-Organization-Id': organizationId
        }
      }
    )

    return {
      ...SmartTemplateMapper.toEntity(response),
      organizationId
    }
  }

  async update(
    id: string,
    organizationId: string,
    template: Partial<TemplateEntity>
  ): Promise<TemplateEntity> {
    let response: SmartTemplateDto

    // If templateFile is provided, use the new patchFormData method
    if (template.templateFile) {
      const data = SmartTemplateMapper.toUpdateDto(template)

      response = await this.httpService.patchFormData<SmartTemplateDto>(
        `${this.baseUrl}/${id}`,
        data,
        {
          headers: {
            'X-Organization-Id': organizationId
          }
        }
      )
    } else {
      // Use JSON for updates without file
      const dto = SmartTemplateMapper.toUpdateDto(template)

      response = await this.httpService.patch<SmartTemplateDto>(
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
      ...SmartTemplateMapper.toEntity(response),
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
