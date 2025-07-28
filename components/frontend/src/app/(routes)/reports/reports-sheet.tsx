'use client'

import React, { useState } from 'react'
import { useIntl } from 'react-intl'
import { useForm, useFieldArray } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Check, Plus, Trash2 } from 'lucide-react'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle
} from '@/components/ui/sheet'
import { LoadingButton } from '@/components/ui/loading-button'
import { Button } from '@/components/ui/button'
import { Form } from '@/components/ui/form'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { InputField } from '@/components/form/input-field'
import { SelectField } from '@/components/form/select-field'
import { SelectItem } from '@/components/ui/select'
import { report } from '@/schema/report'
import { useCreateReport } from '@/client/reports'
import { useListTemplates } from '@/client/templates'
import { CreateReportDto } from '@/core/application/dto/report-dto'
import { useOrganization } from '@lerianstudio/console-layout'
import { useToast } from '@/hooks/use-toast'
import z from 'zod'

type ReportsSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

const initialValues: ReportFormData = {
  templateId: '',
  fields: [
    {
      database: '',
      table: '',
      field: '',
      values: ''
    }
  ]
}

const reportFormSchema = z.object({
  templateId: report.templateId,
  fields: report.fields
})

type ReportFormData = z.infer<typeof reportFormSchema>

export function ReportsSheet({
  open,
  onOpenChange,
  onSuccess
}: ReportsSheetProps) {
  const intl = useIntl()
  const { toast } = useToast()
  const { currentOrganization } = useOrganization()
  const [activeTab, setActiveTab] = useState('details')

  // Initialize form
  const form = useForm<ReportFormData>({
    resolver: zodResolver(reportFormSchema),
    defaultValues: initialValues
  })

  // Initialize field array for filters
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'fields'
  })

  // Fetch templates for dropdown
  const { data: templatesData } = useListTemplates({
    organizationId: currentOrganization?.id || ''
  })

  // API mutation for creating report
  const createReportMutation = useCreateReport({
    organizationId: currentOrganization?.id!,
    onSuccess: () => {
      form.reset()
      onOpenChange(false)
      onSuccess?.()
      toast({
        title: intl.formatMessage({
          id: 'reports.create.success',
          defaultMessage: 'New Report successfully started'
        }),
        variant: 'success'
      })
    }
  })

  // Handle form submission
  const handleSubmit = async (values: ReportFormData) => {
    // Transform form data to API payload
    const payload: CreateReportDto = {
      templateId: values.templateId,
      organizationId: currentOrganization.id,
      filters: {
        // Include array of filter criteria
        fields: values.fields.map((filter) => ({
          database: filter.database,
          table: filter.table,
          field: filter.field,
          values: filter.values
            .split(',')
            .map((value: string) => value.trim())
            .filter(Boolean)
        }))
      }
    }

    await createReportMutation.mutateAsync(payload)
  }

  // Handle adding new filter section
  const handleAddFilter = () => {
    append({
      database: '',
      table: '',
      field: '',
      values: ''
    })
  }

  // Handle removing filter section
  const handleRemoveFilter = (index: number) => {
    if (fields.length > 1) {
      remove(index)
    }
  }

  const isLoading = createReportMutation.isPending
  const templates = templatesData?.items || []

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="flex w-[594px] flex-col">
        {/* Header */}
        <SheetHeader>
          <SheetTitle>
            {intl.formatMessage({
              id: 'reports.sheet.title',
              defaultMessage: 'New Report'
            })}
          </SheetTitle>
          <SheetDescription>
            {intl.formatMessage({
              id: 'reports.sheet.description',
              defaultMessage:
                'Upload a template file and configure its settings fr report generation.'
            })}
          </SheetDescription>
        </SheetHeader>

        {/* Form */}
        <Form {...form}>
          {/* Tabs */}
          <Tabs
            value={activeTab}
            onValueChange={setActiveTab}
            className="flex flex-1 flex-col"
          >
            <TabsList className="mb-6">
              <TabsTrigger value="details">
                {intl.formatMessage({
                  id: 'reports.sheet.tab.details',
                  defaultMessage: 'Report Details'
                })}
              </TabsTrigger>
              <TabsTrigger value="filters">
                {intl.formatMessage({
                  id: 'reports.sheet.tab.filters',
                  defaultMessage: 'Filters'
                })}
              </TabsTrigger>
            </TabsList>

            {/* Report Details Tab */}
            <TabsContent value="details" className="flex-1 space-y-4 pb-8">
              {/* Template Selection */}
              <SelectField
                name="templateId"
                label={intl.formatMessage({
                  id: 'reports.form.template',
                  defaultMessage: 'Select template'
                })}
                tooltip={intl.formatMessage({
                  id: 'reports.form.template.tooltip',
                  defaultMessage: 'Choose a template for report generation'
                })}
                control={form.control}
                required
                disabled={isLoading}
              >
                {templates.map((template) => (
                  <SelectItem key={template.id} value={template.id!}>
                    {template.fileName}
                  </SelectItem>
                ))}
              </SelectField>

              {/* Mandatory fields note */}
              <p className="text-muted-foreground text-sm">
                {intl.formatMessage({
                  id: 'reports.form.mandatoryFields',
                  defaultMessage: '(*) mandatory fields'
                })}
              </p>
            </TabsContent>

            {/* Filters Tab */}
            <TabsContent value="filters" className="flex-1 space-y-4 pb-8">
              {/* Filters Header */}
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-1">
                    <p className="text-muted-foreground text-sm">
                      {intl.formatMessage({
                        id: 'reports.filters.description',
                        defaultMessage:
                          'Define the filtering criteria applied to the data used in the report'
                      })}
                    </p>
                  </div>
                  <Button
                    size="sm"
                    variant="default"
                    type="button"
                    onClick={handleAddFilter}
                    disabled={isLoading}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>

                {/* Dynamic Filter Sections */}
                <div className="space-y-4">
                  {fields.map((field, index) => (
                    <div
                      key={field.id}
                      className="space-y-4 rounded-lg border p-6"
                    >
                      {/* Filter Header with Remove Button */}
                      {fields.length > 1 && (
                        <div className="-mt-4 -mr-4 flex justify-end">
                          <Button
                            size="sm"
                            variant="ghost"
                            type="button"
                            onClick={() => handleRemoveFilter(index)}
                            disabled={isLoading}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      )}

                      <div className="grid grid-cols-2 gap-4">
                        {/* Database */}
                        <InputField
                          name={`fields.${index}.database`}
                          label={intl.formatMessage({
                            id: 'reports.filters.database',
                            defaultMessage: 'Database'
                          })}
                          tooltip={intl.formatMessage({
                            id: 'reports.filters.database.tooltip',
                            defaultMessage: 'Select the database to filter from'
                          })}
                          control={form.control}
                          required
                          disabled={isLoading}
                        />

                        {/* Table */}
                        <InputField
                          name={`fields.${index}.table`}
                          label={intl.formatMessage({
                            id: 'reports.filters.table',
                            defaultMessage: 'Table'
                          })}
                          tooltip={intl.formatMessage({
                            id: 'reports.filters.table.tooltip',
                            defaultMessage: 'Select the table to filter from'
                          })}
                          control={form.control}
                          required
                          disabled={isLoading}
                        />
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        {/* Field */}
                        <InputField
                          name={`fields.${index}.field`}
                          label={intl.formatMessage({
                            id: 'reports.filters.field',
                            defaultMessage: 'Field'
                          })}
                          tooltip={intl.formatMessage({
                            id: 'reports.filters.field.tooltip',
                            defaultMessage: 'Select the field to filter by'
                          })}
                          control={form.control}
                          required
                          disabled={isLoading}
                        />
                      </div>

                      {/* Values */}
                      <InputField
                        name={`fields.${index}.values`}
                        label={intl.formatMessage({
                          id: 'reports.filters.values',
                          defaultMessage: 'Values'
                        })}
                        tooltip={intl.formatMessage({
                          id: 'reports.filters.values.tooltip',
                          defaultMessage: 'Enter the values to filter by'
                        })}
                        description={intl.formatMessage({
                          id: 'reports.filters.values.description',
                          defaultMessage:
                            'Use comma separation to indicate multiple values'
                        })}
                        control={form.control}
                        textArea
                        disabled={isLoading}
                      />
                    </div>
                  ))}
                </div>
              </div>

              {/* Mandatory fields note */}
              <p className="text-muted-foreground text-sm">
                {intl.formatMessage({
                  id: 'reports.form.mandatoryFields',
                  defaultMessage: '(*) mandatory fields'
                })}
              </p>
            </TabsContent>
          </Tabs>

          {/* Footer */}
          <SheetFooter className="mt-auto">
            <LoadingButton
              type="submit"
              loading={isLoading}
              className="flex w-full items-center gap-2"
              onClick={form.handleSubmit(handleSubmit)}
            >
              {intl.formatMessage({
                id: 'reports.form.generateButton',
                defaultMessage: 'Generate Report'
              })}
              <Check className="h-4 w-4" />
            </LoadingButton>
          </SheetFooter>
        </Form>
      </SheetContent>
    </Sheet>
  )
}
