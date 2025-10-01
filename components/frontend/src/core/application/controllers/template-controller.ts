import { z } from 'zod'
import { CreateTemplateUseCase } from '../use-cases/templates/create-template-use-case'
import { ListTemplatesUseCase } from '../use-cases/templates/list-templates-use-case'
import { GetTemplateUseCase } from '../use-cases/templates/get-template-use-case'
import { UpdateTemplateUseCase } from '../use-cases/templates/update-template-use-case'
import { DeleteTemplateUseCase } from '../use-cases/templates/delete-template-use-case'
import { OutputFormat } from '@/core/domain/entities/template-entity'
import type { TemplateSearchParamDto } from '../dto/template-dto'
import {
  Body,
  Controller,
  Delete,
  Get,
  Inject,
  Param,
  Patch,
  Post,
  Query
} from '@lerianstudio/sindarian-server'
import { NextResponse } from 'next/server'

// Create const array from OutputFormat type for Zod validation
const OUTPUT_FORMAT_VALUES: OutputFormat[] = ['csv', 'xml', 'html', 'txt']

const CreateFormDataSchema = z.object({
  name: z.string().max(1000),
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
@Controller('/organizations/:id/templates')
export class TemplateController {
  constructor(
    @Inject(CreateTemplateUseCase)
    private readonly createTemplateUseCase: CreateTemplateUseCase,
    @Inject(ListTemplatesUseCase)
    private readonly listTemplatesUseCase: ListTemplatesUseCase,
    @Inject(GetTemplateUseCase)
    private readonly getTemplateUseCase: GetTemplateUseCase,
    @Inject(UpdateTemplateUseCase)
    private readonly updateTemplateUseCase: UpdateTemplateUseCase,
    @Inject(DeleteTemplateUseCase)
    private readonly deleteTemplateUseCase: DeleteTemplateUseCase
  ) {}

  /**
   * Get a specific template by ID
   * GET /api/organizations/{id}/templates/{templateId}
   */
  @Get('/:templateId')
  async fetchById(
    @Param('id') organizationId: string,
    @Param('templateId') templateId: string
  ) {
    return await this.getTemplateUseCase.execute(templateId!, organizationId)
  }

  /**
   * List templates with pagination and filtering
   * GET /api/organizations/{id}/templates
   */
  @Get('/')
  async fetchAll(
    @Param('id') organizationId: string,
    @Query() query: TemplateSearchParamDto
  ) {
    return await this.listTemplatesUseCase.execute(organizationId, query)
  }

  /**
   * Create a new template
   * POST /api/organizations/{id}/templates
   */
  @Post('/')
  async create(
    @Param('id') organizationId: string,
    @Body() body: CreateFormData
  ) {
    const { templateFile } = body

    // Additional file validation
    if (!templateFile || templateFile.size === 0) {
      return NextResponse.json(
        { message: 'Template file is required' },
        { status: 400 }
      )
    }

    const template = await this.createTemplateUseCase.execute({
      organizationId,
      ...body
    })

    return NextResponse.json(template, { status: 201 })
  }

  /**
   * Update an existing template
   * PATCH /api/organizations/{id}/templates/{templateId}
   */
  @Patch('/:templateId')
  async update(
    @Param('id') organizationId: string,
    @Param('templateId') templateId: string,
    @Body() body: UpdateFormData
  ) {
    const { name, outputFormat, templateFile } = body

    return await this.updateTemplateUseCase.execute(
      templateId!,
      organizationId,
      {
        name,
        outputFormat,
        templateFile: templateFile || undefined
      }
    )
  }

  /**
   * Delete a template (soft delete)
   * DELETE /api/organizations/{id}/templates/{templateId}
   */
  @Delete('/:templateId')
  async delete(
    @Param('id') organizationId: string,
    @Param('templateId') templateId: string
  ) {
    await this.deleteTemplateUseCase.execute(templateId!, organizationId)

    return NextResponse.json({}, { status: 200 })
  }
}
