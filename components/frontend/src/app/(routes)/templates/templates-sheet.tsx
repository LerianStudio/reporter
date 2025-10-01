'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet'
import { LoadingButton } from '@/components/ui/loading-button'
import { Form } from '@/components/ui/form'
import { InputField } from '@/components/form/input-field'
import { SelectField } from '@/components/form/select-field'
import { FileInputField } from '@/components/form/file-input-field'
import { SelectItem } from '@/components/ui/select'
import {
  createTemplateFormSchema,
  createTemplateUpdateFormSchema,
  TemplateFormData,
  TemplateUpdateFormData,
  OUTPUT_FORMAT_OPTIONS
} from '@/schema/template'
import { useCreateTemplate, useUpdateTemplate } from '@/client/templates'
import {
  TemplateDto,
  CreateTemplateDto
} from '@/core/application/dto/template-dto'
import { useOrganization } from '@lerianstudio/console-layout'
import { getInitialValues } from '@/lib/form'
import { useToast } from '@/hooks/use-toast'

type TemplatesSheetProps = {
  mode: 'create' | 'edit'
  data: TemplateDto | null
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

const initialValues = {
  name: '',
  outputFormat: undefined,
  templateFile: undefined
}

export function TemplatesSheet({
  mode,
  data,
  open,
  onOpenChange,
  onSuccess
}: TemplatesSheetProps) {
  const intl = useIntl()
  const { toast } = useToast()
  const { currentOrganization } = useOrganization()

  const isCreateMode = mode === 'create'
  const schema = isCreateMode
    ? createTemplateFormSchema(intl)
    : createTemplateUpdateFormSchema(intl)

  const form = useForm<TemplateFormData | TemplateUpdateFormData>({
    resolver: zodResolver(schema),
    values: getInitialValues(initialValues, data ?? {}),
    defaultValues: initialValues
  })
  const createTemplateMutation = useCreateTemplate({
    organizationId: currentOrganization?.id!,
    onSuccess: (data: TemplateDto) => {
      const template = data as TemplateDto
      form.reset()
      onOpenChange(false)
      onSuccess?.()
      toast({
        title: intl.formatMessage(
          {
            id: 'templates.create.success',
            defaultMessage: 'New Template {name} successfully created'
          },
          {
            name: template.name
          }
        ),
        variant: 'success'
      })
    }
  })

  const updateTemplateMutation = useUpdateTemplate({
    organizationId: currentOrganization?.id!,
    templateId: data?.id || '',
    onSuccess: (data: unknown) => {
      const template = data as TemplateDto
      form.reset()
      onOpenChange(false)
      onSuccess?.()
      toast({
        title: intl.formatMessage(
          {
            id: 'templates.update.success',
            defaultMessage: 'Template {name} successfully updated'
          },
          {
            name: template.name
          }
        ),
        variant: 'success'
      })
    }
  })

  const handleSubmit = async (
    values: TemplateFormData | TemplateUpdateFormData
  ) => {
    if (isCreateMode) {
      const createData = values as TemplateFormData

      const formData = new FormData()
      formData.append('organizationId', currentOrganization.id)
      formData.append('name', createData.name)
      formData.append('outputFormat', createData.outputFormat)
      formData.append('templateFile', createData.templateFile)

      await createTemplateMutation.mutateAsync(formData)
    } else {
      const updateData = values as TemplateUpdateFormData

      const formData = new FormData()
      if (updateData.name) formData.append('name', updateData.name)
      if (updateData.outputFormat)
        formData.append('outputFormat', updateData.outputFormat)
      if (updateData.templateFile)
        formData.append('templateFile', updateData.templateFile)

      await updateTemplateMutation.mutateAsync(formData)
    }
  }

  const isLoading =
    createTemplateMutation.isPending || updateTemplateMutation.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-[594px] flex-col">
        <SheetHeader className="mb-8">
          <SheetTitle>
            {intl.formatMessage({
              id: 'templates.sheet.title',
              defaultMessage: 'Template Details'
            })}
          </SheetTitle>
          <SheetDescription>
            {intl.formatMessage({
              id: 'templates.sheet.description',
              defaultMessage:
                'Upload a template file and configure its settings for report generation.'
            })}
          </SheetDescription>
        </SheetHeader>

        <Form {...form}>
          <div className="flex-1 space-y-4 pb-8">
            <InputField
              name="name"
              label={intl.formatMessage({
                id: 'templates.form.templateName',
                defaultMessage: 'Template Name'
              })}
              tooltip={intl.formatMessage({
                id: 'templates.form.templateName.tooltip',
                defaultMessage: 'Enter a descriptive name for your template'
              })}
              control={form.control}
              required
              disabled={isLoading}
              data-testid="template-name-input"
            />

            <SelectField
              name="outputFormat"
              label={intl.formatMessage({
                id: 'templates.form.outputFormat',
                defaultMessage: 'Output Formats'
              })}
              tooltip={intl.formatMessage({
                id: 'templates.form.outputFormat.tooltip',
                defaultMessage:
                  'Defines the output format for the generated report, like HTML, TXT, etc.'
              })}
              description={intl.formatMessage({
                id: 'templates.form.outputFormat.description',
                defaultMessage: 'Select the output format.'
              })}
              control={form.control}
              required
              disabled={isLoading}
              data-testid="template-output-format-select"
            >
              {OUTPUT_FORMAT_OPTIONS.map((option) => (
                <SelectItem
                  key={option.value}
                  value={option.value.toLowerCase()}
                >
                  {option.label}
                </SelectItem>
              ))}
            </SelectField>

            <FileInputField
              name="templateFile"
              label={intl.formatMessage({
                id: 'templates.form.templateFile',
                defaultMessage: 'Template File (.tpl)'
              })}
              tooltip={intl.formatMessage({
                id: 'templates.form.templateFile.tooltip',
                defaultMessage:
                  'Upload a template file with your business logic and formatting'
              })}
              description={intl.formatMessage({
                id: 'templates.form.templateFile.description',
                defaultMessage:
                  'Upload a template file with your business logic and formatting.'
              })}
              control={form.control}
              accept=".tpl"
              maxSize={5 * 1024 * 1024} // 5MB
              required
              disabled={isLoading}
              data-testid="template-file-input"
            />

            <p className="text-muted-foreground text-sm">
              {intl.formatMessage({
                id: 'common.form.mandatoryFields',
                defaultMessage: '(*) mandatory fields'
              })}
            </p>
          </div>

          <SheetFooter className="mt-auto">
            <LoadingButton
              type="submit"
              loading={isLoading}
              className="flex w-full items-center gap-2"
              onClick={form.handleSubmit(handleSubmit)}
              data-testid="template-submit-button"
            >
              {intl.formatMessage({
                id: 'templates.form.saveButton',
                defaultMessage: 'Save'
              })}
            </LoadingButton>
          </SheetFooter>
        </Form>
      </SheetContent>
    </Sheet>
  )
}
