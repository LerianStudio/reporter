import { inject } from 'inversify'
import { NextResponse } from 'next/server'
import { z } from 'zod'
import { Controller } from '@/lib/http/server/decorators/controller-decorator'
import { LoggerInterceptor } from '@/core/infrastructure/logger/decorators'
import { CreateTemplateUseCase } from '../use-cases/templates/create-template-use-case'
import { ListTemplatesUseCase } from '../use-cases/templates/list-templates-use-case'
import { GetTemplateUseCase } from '../use-cases/templates/get-template-use-case'
import { UpdateTemplateUseCase } from '../use-cases/templates/update-template-use-case'
import { DeleteTemplateUseCase } from '../use-cases/templates/delete-template-use-case'
import { ValidateFormData } from '@/lib/zod/decorators/validate-form-data'
import { OutputFormat } from '@/core/domain/entities/template-entity'
import { Delete, Get, Param, Query } from '@/lib/http/server'
import type { TemplateSearchParamDto } from '../dto/template-dto'
import { BaseController } from '@/lib/http/server/base-controller'

type TemplateParams = {
  id: string // organization ID
  templateId?: string
}

// Create const array from OutputFormat type for Zod validation
const OUTPUT_FORMAT_VALUES: OutputFormat[] = ['csv', 'xml', 'html', 'txt']

const CreateFormDataSchema = z.object({
  organizationId: z.string().min(1),
  name: z.string().max(1000).optional(),
  outputFormat: z.enum(
    OUTPUT_FORMAT_VALUES as [OutputFormat, ...OutputFormat[]]
  ),
  templateFile: z
    .instanceof(File, {
      message: 'Template file is required'
    })
    .refine(
      (file) => file.name.endsWith('.tpl'),
      'File must be a .tpl template file'
    )
    .refine(
      (file) => file.size <= 5 * 1024 * 1024, // 5MB limit
      'File size must be less than 5MB'
    )
    .refine((file) => file.size > 0, 'File cannot be empty')
})

type CreateFormData = z.infer<typeof CreateFormDataSchema>

const UpdateFormDataSchema = z.object({
  name: z.string().max(1000).optional(),
  outputFormat: z
    .enum(OUTPUT_FORMAT_VALUES as [OutputFormat, ...OutputFormat[]])
    .optional(),
  templateFile: z
    .instanceof(File)
    .refine(
      (file) => file.name.endsWith('.tpl'),
      'File must be a .tpl template file'
    )
    .refine(
      (file) => file.size <= 5 * 1024 * 1024, // 5MB limit
      'File size must be less than 5MB'
    )
    .refine((file) => file.size > 0, 'File cannot be empty')
    .optional()
})

type UpdateFormData = z.infer<typeof UpdateFormDataSchema>

/**
 * Template Controller
 *
 * Next.js API route controller for handling template-related HTTP requests.
 * Provides RESTful endpoints that delegate to appropriate use cases.
 * Follows console patterns with proper request/response handling.
 */
@LoggerInterceptor()
@Controller()
export class TemplateController extends BaseController {
  constructor(
    @inject(CreateTemplateUseCase)
    private readonly createTemplateUseCase: CreateTemplateUseCase,
    @inject(ListTemplatesUseCase)
    private readonly listTemplatesUseCase: ListTemplatesUseCase,
    @inject(GetTemplateUseCase)
    private readonly getTemplateUseCase: GetTemplateUseCase,
    @inject(UpdateTemplateUseCase)
    private readonly updateTemplateUseCase: UpdateTemplateUseCase,
    @inject(DeleteTemplateUseCase)
    private readonly deleteTemplateUseCase: DeleteTemplateUseCase
  ) {
    super()
  }

  /**
   * Get a specific template by ID
   * GET /api/organizations/{id}/templates/{templateId}
   */
  @Get()
  async fetchById(
    @Param('id') organizationId: string,
    @Param('templateId') templateId: string
  ) {
    const template = await this.getTemplateUseCase.execute(
      templateId!,
      organizationId
    )

    return NextResponse.json(template)
  }

  /**
   * List templates with pagination and filtering
   * GET /api/organizations/{id}/templates
   */
  @Get()
  async fetchAll(
    @Param('id') organizationId: string,
    @Query() query: TemplateSearchParamDto
  ) {
    const templates = await this.listTemplatesUseCase.execute(
      organizationId,
      query
    )

    return NextResponse.json(templates)
  }

  /**
   * Create a new template
   * POST /api/organizations/{id}/templates
   */
  @ValidateFormData(CreateFormDataSchema)
  async create(request: Request, { params }: { params: TemplateParams }) {
    const formData = await request.formData()
    const organizationId = (await params).id
    const name = (formData.get('name') as string) || ''
    const outputFormat = formData.get('outputFormat') as any
    const templateFile = formData.get('templateFile') as File

    // Additional file validation
    if (!templateFile || templateFile.size === 0) {
      return NextResponse.json(
        { error: 'Template file is required' },
        { status: 400 }
      )
    }

    const template = await this.createTemplateUseCase.execute({
      organizationId,
      name,
      outputFormat,
      templateFile
    })

    return NextResponse.json(template, { status: 201 })
  }

  /**
   * Update an existing template
   * PATCH /api/organizations/{id}/templates/{templateId}
   */
  @ValidateFormData(UpdateFormDataSchema)
  async update(request: Request, { params }: { params: TemplateParams }) {
    const formData = await request.formData()
    const organizationId = (await params).id
    const templateId = (await params).templateId

    const name = (formData.get('name') as string) || undefined
    const outputFormat = formData.get('outputFormat') as any
    const templateFile = formData.get('templateFile') as File | null

    const template = await this.updateTemplateUseCase.execute(
      templateId!,
      organizationId,
      {
        name,
        outputFormat,
        templateFile: templateFile || undefined
      }
    )

    return NextResponse.json(template)
  }

  /**
   * Delete a template (soft delete)
   * DELETE /api/organizations/{id}/templates/{templateId}
   */
  @Delete()
  async delete(
    @Param('id') organizationId: string,
    @Param('templateId') templateId: string
  ) {
    await this.deleteTemplateUseCase.execute(templateId!, organizationId)

    return NextResponse.json({}, { status: 200 })
  }
}
