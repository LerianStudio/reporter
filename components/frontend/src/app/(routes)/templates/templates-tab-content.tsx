'use client'

import React, { useState, useMemo, useEffect } from 'react'
import { useIntl } from 'react-intl'
import { TemplatesDataTable } from '@/app/(routes)/templates/templates-data-table'
import { TemplatesSheet } from './templates-sheet'
import { Button } from '@/components/ui/button'
import { useListTemplates, useDeleteTemplate } from '@/client/templates'
import { TemplateDto } from '@/core/application/dto/template-dto'
import { useQueryParams } from '@/hooks/use-query-params'
import { useOrganization } from '@lerianstudio/console-layout'
import { useCreateUpdateSheet } from '@/components/sheet/use-create-update-sheet'
import { EntityBox } from '@/components/entity-box'
import { useConfirmDialog } from '@/components/confirmation-dialog/use-confirm-dialog'
import ConfirmationDialog from '@/components/confirmation-dialog'
import { InputField, SelectField } from '@/components/form'
import { FormProvider } from 'react-hook-form'
import { FunnelXIcon, Plus } from 'lucide-react'
import { useToast } from '@/hooks/use-toast'
import { PaginationLimitField } from '@/components/form/pagination-limit-field'
import { SelectItem } from '@/components/ui/select'
import { OUTPUT_FORMAT_OPTIONS } from '@/schema/template'

export function TemplatesTabContent() {
  const intl = useIntl()
  const { toast } = useToast()
  const [total, setTotal] = useState(1000000)
  const { currentOrganization } = useOrganization()

  // Sheet state management
  const { handleCreate, handleEdit, sheetProps } =
    useCreateUpdateSheet<TemplateDto>()

  // Confirmation dialog for delete operation
  const {
    handleDialogOpen,
    dialogProps,
    handleDialogClose,
    data: selectedTemplate
  } = useConfirmDialog<TemplateDto>({
    onConfirm: () => {
      if (selectedTemplate?.id && currentOrganization?.id) {
        deleteTemplateMutation.mutate({ id: selectedTemplate.id })
      }
    }
  })

  // Delete template mutation
  const deleteTemplateMutation = useDeleteTemplate({
    organizationId: currentOrganization?.id || '',
    templateId: selectedTemplate?.id || '',
    onSuccess: () => {
      handleDialogClose()
      refetch()
      toast({
        title: intl.formatMessage({
          id: 'templates.delete.success',
          defaultMessage: 'Template deleted successfully'
        }),
        variant: 'success'
      })
    }
  })

  const { form, searchValues, pagination } = useQueryParams({
    total
  })

  // Fetch templates data using searchValues from useQueryParams
  const {
    data: templatesData,
    isLoading,
    refetch
  } = useListTemplates({
    filters: searchValues,
    organizationId: currentOrganization?.id || ''
  } as any)

  // Create table data structure
  const table = useMemo(
    () => ({
      getRowModel: () => ({
        rows:
          templatesData?.items?.map((item) => ({
            id: item.id,
            original: item
          })) || []
      })
    }),
    [templatesData]
  )

  // Event handlers with sheet integration
  const handleCreateTemplate = () => {
    handleCreate()
  }

  const handleEditTemplate = (template: TemplateDto) => {
    handleEdit(template)
  }

  const handleDeleteTemplate = (id: string, template: TemplateDto) => {
    handleDialogOpen(id, template)
  }

  // Success callback for sheet operations
  const handleSheetSuccess = () => {
    refetch()
  }

  // Fallback data structure for empty state
  const templatesWithFallback = templatesData || { items: [] }

  return (
    <>
      {/* Confirmation Dialog for Delete */}
      <ConfirmationDialog
        title={intl.formatMessage({
          id: 'common.confirmDeletion',
          defaultMessage: 'Confirm Deletion'
        })}
        description={intl.formatMessage(
          {
            id: 'templates.delete.description',
            defaultMessage:
              'Are you sure you want to delete the template "{fileName}"? This action cannot be undone.'
          },
          { fileName: selectedTemplate?.fileName || '' }
        )}
        loading={deleteTemplateMutation.isPending}
        {...dialogProps}
      />

      <FormProvider {...form}>
        <EntityBox.Collapsible>
          <EntityBox.Banner>
            <EntityBox.Header
              title={intl.formatMessage({
                id: 'templates.title',
                defaultMessage: 'Templates'
              })}
              subtitle={intl.formatMessage({
                id: 'templates.subtitle',
                defaultMessage:
                  'Manage your report templates and their configurations.'
              })}
              tooltip={intl.formatMessage({
                id: 'templates.tooltip',
                defaultMessage:
                  'Templates are pre-configured report structures that help you standardize and automate how data is presented. You can define layouts, formats, and parameters to streamline report generation and ensure consistency.'
              })}
              tooltipWidth="655px"
            />

            <EntityBox.Actions>
              <EntityBox.CollapsibleTrigger />
              <Button
                icon={<Plus />}
                iconPlacement="end"
                onClick={handleCreateTemplate}
              >
                {intl.formatMessage({
                  id: 'templates.listingTemplate.addButton',
                  defaultMessage: 'New Template'
                })}
              </Button>
            </EntityBox.Actions>
          </EntityBox.Banner>
          <EntityBox.CollapsibleContent>
            <div className="col-span-2 flex grow flex-row gap-2">
              <InputField
                name="name"
                placeholder={intl.formatMessage({
                  id: 'common.searchPlaceholder',
                  defaultMessage: 'Search...'
                })}
                control={form.control}
              />
              <SelectField
                className="min-w-50"
                name="outputFormat"
                placeholder={intl.formatMessage({
                  id: 'common.outputFormatPlaceholder',
                  defaultMessage: 'Output Format...'
                })}
                control={form.control}
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
            </div>
            <div className="col-start-3 flex items-center justify-end gap-2">
              <PaginationLimitField control={form.control} />
              <Button
                variant="secondary"
                className="h-[34px] w-[34px] p-2"
                onClick={() => {
                  form.setValue('outputFormat', '')
                  form.setValue('name', '')
                }}
              >
                <FunnelXIcon size={16} />
              </Button>
            </div>
          </EntityBox.CollapsibleContent>
        </EntityBox.Collapsible>

        <TemplatesDataTable
          templates={templatesWithFallback}
          isLoading={isLoading}
          table={table}
          onDelete={handleDeleteTemplate}
          handleCreate={handleCreateTemplate}
          handleEdit={handleEditTemplate}
          total={total}
          pagination={pagination}
          form={form}
        />
      </FormProvider>

      {/* Templates Sheet for Create/Edit */}
      <TemplatesSheet {...sheetProps} onSuccess={handleSheetSuccess} />
    </>
  )
}
