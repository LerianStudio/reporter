'use client'

import React, { useState } from 'react'
import { useIntl } from 'react-intl'
import { useForm, useFieldArray, Control } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Check, Plus } from 'lucide-react'
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
import { SelectField } from '@/components/form/select-field'
import { SelectItem } from '@/components/ui/select'
import { report } from '@/schema/report'
import { useCreateReport } from '@/client/reports'
import { useListTemplates } from '@/client/templates'
import { useListDataSources } from '@/client/data-sources'
import { useOrganization } from '@lerianstudio/console-layout'
import { useToast } from '@/hooks/use-toast'
import { ReportsSheetFilter } from './reports-sheet-filter'
import z from 'zod'

type ReportsSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess?: () => void
}

const initialValues: ReportFormData = {
  templateId: '',
  fields: []
}

const reportFormSchema = z.object({
  templateId: report.templateId,
  fields: report.fields
})

export type ReportFormData = z.infer<typeof reportFormSchema>

export function ReportsSheet({
  open,
  onOpenChange,
  onSuccess
}: ReportsSheetProps) {
  const intl = useIntl()
  const { toast } = useToast()
  const { currentOrganization } = useOrganization()
  const [activeTab, setActiveTab] = useState('details')

  const form = useForm<ReportFormData>({
    resolver: zodResolver(reportFormSchema),
    defaultValues: initialValues
  })

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: 'fields'
  })

  const { data: templates } = useListTemplates({
    organizationId: currentOrganization?.id || '',
    filters: {
      page: 1,
      limit: 100
    }
  })

  const { data: dataSources } = useListDataSources({
    organizationId: currentOrganization?.id || ''
  })

  const createReportMutation = useCreateReport({
    organizationId: currentOrganization?.id || '',
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

  const handleSubmit = async (values: ReportFormData) => {
    const validFields = values.fields.filter((field) => {
      return (
        field.database &&
        field.table &&
        field.field &&
        field.values &&
        (Array.isArray(field.values)
          ? field.values.length > 0 &&
            field.values.some((value) => value?.trim())
          : field.values.trim())
      )
    })

    const payload = {
      templateId: values.templateId,
      ...(validFields.length > 0 && { fields: validFields })
    }

    await createReportMutation.mutateAsync(payload)
  }

  const handleAddFilter = () => {
    append({
      database: '',
      table: '',
      field: '',
      operator: 'eq' as const,
      values: []
    })
  }

  const isLoading = createReportMutation.isPending

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        data-testid="reports-sheet"
        side="right"
        className="flex w-[594px] flex-col"
      >
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

        <Form {...form}>
          <Tabs
            value={activeTab}
            onValueChange={setActiveTab}
            className="flex flex-1 flex-col"
          >
            <TabsList className="mb-6">
              <TabsTrigger data-testid="report-details-tab" value="details">
                {intl.formatMessage({
                  id: 'reports.sheet.tab.details',
                  defaultMessage: 'Report Details'
                })}
              </TabsTrigger>
              <TabsTrigger data-testid="report-filters-tab" value="filters">
                {intl.formatMessage({
                  id: 'reports.sheet.tab.filters',
                  defaultMessage: 'Filters'
                })}
              </TabsTrigger>
            </TabsList>

            <TabsContent value="details" className="flex-1 space-y-4 pb-8">
              <SelectField
                data-testid="report-template-select"
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
                disabled={isLoading}
              >
                {templates?.items?.map((template) => (
                  <SelectItem key={template.id} value={template.id!}>
                    {template.name}
                  </SelectItem>
                ))}
              </SelectField>

              <p className="text-muted-foreground text-sm">
                {intl.formatMessage({
                  id: 'common.form.mandatoryFields',
                  defaultMessage: '(*) mandatory fields'
                })}
              </p>
            </TabsContent>

            <TabsContent value="filters" className="flex-1 space-y-4 pb-8">
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
                    data-testid="report-add-filter-button"
                    size="sm"
                    variant="default"
                    type="button"
                    onClick={handleAddFilter}
                    disabled={isLoading}
                  >
                    <Plus className="h-4 w-4" />
                  </Button>
                </div>

                <div className="space-y-4">
                  {fields.map((field, index) => (
                    <ReportsSheetFilter
                      key={field.id}
                      name={`fields.${index}`}
                      onRemove={() => remove(index)}
                      control={form.control}
                      loading={isLoading}
                      dataSources={dataSources || []}
                    />
                  ))}
                </div>
              </div>

              <p className="text-muted-foreground text-sm">
                {intl.formatMessage({
                  id: 'common.form.mandatoryFields',
                  defaultMessage: '(*) mandatory fields'
                })}
              </p>
            </TabsContent>
          </Tabs>

          <SheetFooter className="mt-auto">
            <LoadingButton
              data-testid="report-generate-button"
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
