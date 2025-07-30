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
  const { setValue } = useFormContext<ReportFormData>()

  // Watch the specific field values to enable cascading selects
  const field = useWatch({
    control,
    name: name as any
  })

  const currentDatabase = field?.database || ''
  const currentTable = field?.table || ''
  const currentField = field?.field || ''

  const availableTables = React.useMemo(() => {
    const dataSource = dataSources?.find((ds) => ds.id === currentDatabase)
    return dataSource?.tables || []
  }, [dataSources, currentDatabase])

  const availableFields = React.useMemo(() => {
    const dataSource = dataSources?.find((ds) => ds.id === currentDatabase)
    const table = dataSource?.tables?.find((t) => t.name === currentTable)
    return table?.fields || []
  }, [currentDatabase, currentTable, dataSources])

  // Clear dependent fields when database changes
  React.useEffect(() => {
    if (currentDatabase) {
      setValue(`${name}.table` as any, '')
      setValue(`${name}.field` as any, '')
      setValue(`${name}.values` as any, '')
    }
  }, [currentDatabase, setValue, name])

  // Clear dependent fields when table changes
  React.useEffect(() => {
    if (currentTable) {
      setValue(`${name}.field` as any, '')
      setValue(`${name}.values` as any, '')
    }
  }, [currentTable, setValue, name])

  // Clear values when field changes
  React.useEffect(() => {
    if (currentField) {
      setValue(`${name}.values` as any, '')
    }
  }, [currentField, setValue, name])

  return (
    <div className="space-y-4 rounded-lg border p-6">
      {/* Filter Header with Remove Button */}
      <div className="-mt-4 -mr-4 flex justify-end">
        <Button
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
        {/* Database */}
        <SelectField
          name={`${name}.database`}
          label={intl.formatMessage({
            id: 'reports.filters.database',
            defaultMessage: 'Database'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.database.tooltip',
            defaultMessage: 'Select the database to filter from'
          })}
          control={control}
          disabled={loading}
        >
          {dataSources?.map((dataSource) => (
            <SelectItem key={dataSource.id} value={dataSource.id}>
              {dataSource.externalName || dataSource.id}
            </SelectItem>
          ))}
        </SelectField>

        {/* Table */}
        <SelectField
          name={`${name}.table`}
          label={intl.formatMessage({
            id: 'reports.filters.table',
            defaultMessage: 'Table'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.table.tooltip',
            defaultMessage: 'Select the table to filter from'
          })}
          control={control}
          disabled={loading || !currentDatabase}
        >
          {availableTables.map((table) => (
            <SelectItem key={table.name} value={table.name}>
              {table.name}
            </SelectItem>
          ))}
        </SelectField>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {/* Field */}
        <SelectField
          name={`${name}.field`}
          label={intl.formatMessage({
            id: 'reports.filters.field',
            defaultMessage: 'Field'
          })}
          tooltip={intl.formatMessage({
            id: 'reports.filters.field.tooltip',
            defaultMessage: 'Select the field to filter by'
          })}
          control={control}
          disabled={loading || !currentTable}
        >
          {availableFields.map((fieldName) => (
            <SelectItem key={fieldName} value={fieldName}>
              {fieldName}
            </SelectItem>
          ))}
        </SelectField>
      </div>

      {/* Values */}
      <InputField
        name={`${name}.values`}
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
          defaultMessage: 'Use comma separation to indicate multiple values'
        })}
        control={control}
        textArea
        disabled={loading}
      />
    </div>
  )
}
