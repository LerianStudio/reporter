'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { Trash2 } from 'lucide-react'
import { Control, useWatch, useFormContext } from 'react-hook-form'
import { Button } from '@/components/ui/button'
import { InputField } from '@/components/form/input-field'
import { SelectField } from '@/components/form/select-field'
import { SelectItem } from '@/components/ui/select'
import { DataSourceDto } from '@/core/application/dto/data-source-dto'
import { type ReportFormData } from './reports-sheet'
import { filterFieldSchema } from '@/schema/report'
import { z } from 'zod'
import { useGetDataSourceById } from '@/client/data-sources'
import { useOrganization } from '@lerianstudio/console-layout'
import {
  getOperatorsForFieldType,
  operatorRequiresNoValues,
  operatorRequiresMultipleValues
} from '@/utils/filter-operators'

type ReportsSheetFilterProps = {
  name: string
  onRemove?: () => void
  control: Control<ReportFormData>
  loading: boolean
  dataSources: DataSourceDto[]
}

export function ReportsSheetFilter({
  name,
  onRemove,
  control,
  loading,
  dataSources
}: ReportsSheetFilterProps) {
  const intl = useIntl()
  const { currentOrganization } = useOrganization()
  const { setValue } = useFormContext<ReportFormData>()

  const field = useWatch({
    control,
    name: name as `fields.${number}`
  }) as z.infer<typeof filterFieldSchema> | undefined

  const { data: dataSourceDetails } = useGetDataSourceById({
    organizationId: currentOrganization?.id || '',
    dataSourceId: field?.database || ''
  })

  const availableTables = React.useMemo(
    () => dataSourceDetails?.tables || [],
    [dataSourceDetails]
  )

  const availableFields = React.useMemo(() => {
    const table = dataSourceDetails?.tables?.find(
      (t) => t.name === field?.table
    )
    return table?.fields || []
  }, [field?.table, dataSourceDetails])

  const selectedField = React.useMemo(() => {
    if (!field?.field) return null

    const fieldInfo = availableFields.find((f) => f.name === field.field)
    const fieldType = fieldInfo?.type || 'string'

    return { name: field.field, type: fieldType }
  }, [field?.field, availableFields])

  const availableOperators = React.useMemo(() => {
    return getOperatorsForFieldType(selectedField?.type)
  }, [selectedField?.type])

  const valuesRequired = React.useMemo(() => {
    return !operatorRequiresNoValues(field?.operator || '')
  }, [field?.operator])

  const valuesDescription = React.useMemo(() => {
    if (!field?.operator) {
      return intl.formatMessage({
        id: 'reports.filters.values.description',
        defaultMessage: 'Use comma separation to indicate multiple values'
      })
    }

    if (operatorRequiresMultipleValues(field.operator)) {
      if (field.operator === 'between') {
        return intl.formatMessage({
          id: 'reports.filters.values.description.between',
          defaultMessage: 'Enter two values separated by comma (e.g., 10, 20)'
        })
      } else {
        return intl.formatMessage({
          id: 'reports.filters.values.description.multiple',
          defaultMessage: 'Enter multiple values separated by commas'
        })
      }
    }

    return intl.formatMessage({
      id: 'reports.filters.values.description.single',
      defaultMessage: 'Enter a single value'
    })
  }, [field?.operator, intl])

  return (
    <div
      data-testid="report-filter-item"
      className="space-y-4 rounded-lg border p-6"
    >
      <div className="-mt-4 -mr-4 flex justify-end">
        <Button
          data-testid="report-filter-remove-button"
          size="sm"
          variant="ghost"
          type="button"
          onClick={onRemove}
          disabled={loading}
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <SelectField
          data-testid="report-filter-database-select"
          name={`${name}.database`}
          label={intl.formatMessage({
            id: 'reports.filters.database',
            defaultMessage: 'Database'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.database.tooltip',
            defaultMessage: 'Select the database to filter from'
          })}
          onChange={() => {
            setValue(`${name}.table` as any, '')
            setValue(`${name}.field` as any, '')
            setValue(`${name}.operator` as any, '')
            setValue(`${name}.values` as any, [])
          }}
          control={control}
          disabled={loading}
        >
          {dataSources?.map((dataSource) => (
            <SelectItem key={dataSource.id} value={dataSource.id}>
              {dataSource.id}
            </SelectItem>
          ))}
        </SelectField>

        <SelectField
          data-testid="report-filter-table-select"
          name={`${name}.table`}
          label={intl.formatMessage({
            id: 'reports.filters.table',
            defaultMessage: 'Table'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.table.tooltip',
            defaultMessage: 'Select the table to filter from'
          })}
          onChange={() => {
            setValue(`${name}.field` as any, '')
            setValue(`${name}.operator` as any, '')
            setValue(`${name}.values` as any, [])
          }}
          control={control}
          disabled={loading || !field?.database}
        >
          {availableTables.map((table) => (
            <SelectItem key={table.name} value={table.name}>
              {table.name}
            </SelectItem>
          ))}
        </SelectField>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <SelectField
          data-testid="report-filter-field-select"
          name={`${name}.field`}
          label={intl.formatMessage({
            id: 'reports.filters.field',
            defaultMessage: 'Field'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.field.tooltip',
            defaultMessage: 'Select the field to filter by'
          })}
          onChange={() => {
            setValue(`${name}.operator` as any, '')
            setValue(`${name}.values` as any, [])
          }}
          control={control}
          disabled={loading || !field?.table}
        >
          {availableFields.map((fieldInfo) => (
            <SelectItem key={fieldInfo.name} value={fieldInfo.name}>
              {fieldInfo.name}
            </SelectItem>
          ))}
        </SelectField>

        <SelectField
          data-testid="report-filter-operator-select"
          name={`${name}.operator`}
          label={intl.formatMessage({
            id: 'reports.filters.operator',
            defaultMessage: 'Operator'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.operator.tooltip',
            defaultMessage: 'Select the comparison operator for filtering'
          })}
          onChange={() => {
            setValue(`${name}.values` as any, [])
          }}
          control={control}
          disabled={loading || !field?.field}
        >
          {availableOperators.map((operator) => (
            <SelectItem key={operator.value} value={operator.value}>
              {intl.formatMessage({
                id: `reports.filters.operators.${operator.value}`,
                defaultMessage: operator.label
              })}
            </SelectItem>
          ))}
        </SelectField>
      </div>

      {valuesRequired && (
        <InputField
          data-testid="report-filter-values-input"
          name={`${name}.values`}
          label={intl.formatMessage({
            id: 'reports.filters.values',
            defaultMessage: 'Values'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.values.tooltip',
            defaultMessage: 'Enter the values to filter by'
          })}
          description={valuesDescription}
          control={control}
          textArea
          disabled={loading || !field?.operator}
        />
      )}
    </div>
  )
}
