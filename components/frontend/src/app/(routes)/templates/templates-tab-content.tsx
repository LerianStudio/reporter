'use client'

import React, { useState, useMemo } from 'react'
import { useIntl } from 'react-intl'
import { TemplatesDataTable } from '@/app/(routes)/templates/templates-data-table'
import { TemplatesSheet } from './templates-sheet'
import { Button } from '@/components/ui/button'
import { useDeleteTemplate } from '@/client/templates'
import { useListTemplates } from '@/hooks/templates/use-list-templates'
import {
  TemplateDto,
  TemplateFiltersDto
} from '@/core/application/dto/template-dto'
import { useQueryParams } from '@/hooks/use-query-params'
import { useOrganization } from '@lerianstudio/console-layout'
import { useCreateUpdateSheet } from '@/components/sheet/use-create-update-sheet'
import { EntityBox } from '@/components/entity-box'
import { useConfirmDialog } from '@/components/confirmation-dialog/use-confirm-dialog'
import ConfirmationDialog from '@/components/confirmation-dialog'
import { InputField, SelectField } from '@/components/form'
import { FormProvider } from 'react-hook-form'
import {
  FunnelXIcon,
  Plus,
  Loader2,
  Calendar as CalendarIcon,
  ChevronDown
} from 'lucide-react'
import { cn, validateDateString, formatDateForDisplay } from '@/lib/utils'
import { useToast } from '@/hooks/use-toast'
import { PaginationLimitField } from '@/components/form/pagination-limit-field'
import { SelectItem } from '@/components/ui/select'
import {
  Popover,
  PopoverContent,
  PopoverTrigger
} from '@/components/ui/popover'
import { Calendar } from '@/components/ui/calendar'
import { OUTPUT_FORMAT_OPTIONS } from '@/schema/template'
import { format } from 'date-fns'

export function TemplatesTabContent() {
  const intl = useIntl()
  const { toast } = useToast()
  const [isCalendarOpen, setIsCalendarOpen] = useState(false)
  const { currentOrganization } = useOrganization()

  const { handleCreate, handleEdit, sheetProps } =
    useCreateUpdateSheet<TemplateDto>()

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

  const { form, searchValues, pagination } = useQueryParams<TemplateFiltersDto>(
    {
      total: 0, // Will be updated after data loads
      initialValues: {
        name: '',
        outputFormat: undefined,
        createdAt: undefined
      }
    }
  )

  const {
    data: templatesData,
    isLoading,
    isFetching,
    refetch
  } = useListTemplates({
    filters: searchValues,
    organizationId: currentOrganization?.id || ''
  })

  const handleDateSelect = (date: Date | undefined) => {
    if (date) {
      form.setValue('createdAt', format(date, 'yyyy-MM-dd'))
      setIsCalendarOpen(false)
    } else {
      form.setValue('createdAt', undefined)
    }
  }

  const clearDateFilter = () => {
    form.setValue('createdAt', undefined)
  }

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
    [templatesData?.items]
  )

  const handleCreateTemplate = () => {
    handleCreate()
  }

  const handleEditTemplate = (template: TemplateDto) => {
    handleEdit(template)
  }

  const handleDeleteTemplate = (id: string, template: TemplateDto) => {
    handleDialogOpen(id, template)
  }

  const handleSheetSuccess = () => {
    refetch()
  }

  const templatesWithFallback = templatesData || { items: [] }

  const selectedDate = useMemo(() => {
    const dateString = form.watch('createdAt')
    const validation = validateDateString(dateString)

    if (!validation.isValid && validation.error) {
      // Only show toast for actual validation errors, not for empty dates
      if (dateString) {
        const errorMessages = {
          format: {
            title: 'templates.filters.invalidDateFormat',
            description: 'templates.filters.invalidDateDescription'
          },
          invalid: {
            title: 'templates.filters.invalidDate',
            description: 'templates.filters.invalidDateValue'
          },
          range: {
            title: 'templates.filters.dateOutOfRange',
            description: 'templates.filters.dateRangeDescription'
          },
          parsing: {
            title: 'templates.filters.dateParsingError',
            description: 'templates.filters.dateParsingErrorDescription'
          }
        }

        const messages = errorMessages[validation.error.type]
        toast({
          title: intl.formatMessage({
            id: messages.title,
            defaultMessage: validation.error.message
          }),
          description: intl.formatMessage({
            id: messages.description,
            defaultMessage: validation.error.message
          }),
          variant: 'destructive'
        })
      }
      return undefined
    }

    return validation.date
  }, [form.watch('createdAt'), intl, toast])

  return (
    <>
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
          { fileName: selectedTemplate?.name || '' }
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
              {isFetching && (
                <div className="text-muted-foreground flex items-center gap-2 text-sm">
                  <Loader2 className="h-4 w-4 animate-spin" />
                  <span>
                    {intl.formatMessage({
                      id: 'templates.loading.updating',
                      defaultMessage: 'Updating...'
                    })}
                  </span>
                </div>
              )}
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
            <div className="col-span-2 flex grow flex-col gap-4 sm:flex-row">
              <InputField
                name="name"
                placeholder={intl.formatMessage({
                  id: 'common.searchPlaceholder',
                  defaultMessage: 'Search...'
                })}
                control={form.control}
              />

              <div className="sm:w-auto sm:flex-none">
                <Popover open={isCalendarOpen} onOpenChange={setIsCalendarOpen}>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={cn(
                        'h-9 w-fit justify-between gap-3 rounded-md border border-[#C7C7C7] bg-white px-3 py-2 text-left text-sm font-normal hover:bg-white focus-visible:ring-2 focus-visible:ring-offset-0',
                        !selectedDate && 'placeholder:text-shadcn-400'
                      )}
                    >
                      <div className="flex items-center">
                        <CalendarIcon className="mr-2 h-4 w-4" />
                        {selectedDate ? (
                          formatDateForDisplay(selectedDate)
                        ) : (
                          <span>
                            {intl.formatMessage({
                              id: 'templates.filters.selectDate',
                              defaultMessage: 'Select date'
                            })}
                          </span>
                        )}
                      </div>
                      <ChevronDown className="h-4 w-4 opacity-50" />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent
                    className="w-auto overflow-hidden rounded-lg p-0"
                    align="start"
                    side="bottom"
                  >
                    <Calendar
                      mode="single"
                      selected={selectedDate}
                      onSelect={handleDateSelect}
                      initialFocus
                      fixedWeeks
                      showOutsideDays={false}
                    />
                  </PopoverContent>
                </Popover>
              </div>

              <SelectField
                className="min-w-50 flex-1 sm:flex-none"
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
                variant="outline"
                className="h-[34px] w-[34px] bg-white p-2 hover:bg-white"
                onClick={() => {
                  form.setValue('outputFormat', undefined)
                  form.setValue('name', '')
                  clearDateFilter()
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
          total={templatesData?.total || templatesData?.items?.length || 0}
          pagination={pagination}
          form={form}
        />
      </FormProvider>

      <TemplatesSheet {...sheetProps} onSuccess={handleSheetSuccess} />
    </>
  )
}
