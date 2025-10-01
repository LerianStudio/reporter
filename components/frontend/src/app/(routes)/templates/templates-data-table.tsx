'use client'

import React from 'react'
import { useIntl } from 'react-intl'
import { MoreVertical, Plus, Database } from 'lucide-react'
import {
  Table,
  TableContainer,
  TableHead,
  TableRow,
  TableHeader,
  TableCell,
  TableBody
} from '@/components/ui/table'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Paper } from '@/components/ui/paper'
import { Pagination, PaginationProps } from '@/components/pagination'
import { UseFormReturn } from 'react-hook-form'
import dayjs from 'dayjs'
import { isNil } from 'lodash'
import { TemplateDto } from '@/core/application/dto/template-dto'
import { NameTableCell } from '@/components/table/name-table-cell'

type TemplatesTableProps = {
  templates: { items: TemplateDto[] }
  isLoading: boolean
  table: {
    getRowModel: () => {
      rows: { id: string; original: TemplateDto }[]
    }
  }
  onDelete: (id: string, template: TemplateDto) => void
  handleCreate: () => void
  handleEdit: (template: TemplateDto) => void
  total: number
  pagination: PaginationProps
  form: UseFormReturn<any>
}

type TemplateRowProps = {
  template: { id: string; original: TemplateDto }
  handleEdit: (template: TemplateDto) => void
  onDelete: (id: string, template: TemplateDto) => void
}

// Empty state component
const EmptyTemplates: React.FC<{ onCreateTemplate: () => void }> = ({
  onCreateTemplate
}) => {
  const intl = useIntl()

  return (
    <div
      data-testid="templates-empty-state"
      className="flex flex-col items-center justify-center px-6 py-12 text-center"
    >
      <Database className="mb-4 h-12 w-12 text-gray-400" />
      <h3
        data-testid="empty-state-title"
        className="mb-2 text-lg font-medium text-gray-900"
      >
        {intl.formatMessage({
          id: 'templates.emptyState.title',
          defaultMessage: 'No templates found'
        })}
      </h3>
      <p
        data-testid="empty-state-description"
        className="mb-6 max-w-sm text-gray-500"
      >
        {intl.formatMessage({
          id: 'templates.emptyState.description',
          defaultMessage:
            "You haven't created any templates yet. Create your first template to get started."
        })}
      </p>
      <Button
        data-testid="empty-state-create-button"
        onClick={onCreateTemplate}
        icon={<Plus className="h-4 w-4" />}
      >
        {intl.formatMessage({
          id: 'common.new.template',
          defaultMessage: 'New Template'
        })}
      </Button>
    </div>
  )
}

const TemplateRow: React.FC<TemplateRowProps> = ({
  template,
  handleEdit,
  onDelete
}) => {
  const intl = useIntl()

  return (
    <TableRow key={template.id} data-testid={`template-row-${template.id}`}>
      <NameTableCell
        name={template.original.name || template.original.fileName}
        onClick={() => handleEdit(template.original)}
      />
      <TableCell>
        <Badge variant="secondary">
          {template.original.outputFormat.toLocaleLowerCase()}
        </Badge>
      </TableCell>
      <TableCell>{dayjs(template.original.updatedAt).format('L')}</TableCell>
      <TableCell className="w-0" align="center">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button
              data-testid={`template-actions-button-${template.id}`}
              variant="secondary"
              className="h-auto w-max p-2"
            >
              <MoreVertical size={16} />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent
            data-testid={`template-actions-menu-${template.id}`}
            align="end"
          >
            <DropdownMenuItem
              data-testid={`template-details-${template.id}`}
              onClick={() => handleEdit(template.original)}
            >
              {intl.formatMessage({
                id: `common.details`,
                defaultMessage: 'Details'
              })}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              data-testid={`template-delete-${template.id}`}
              onClick={() => {
                onDelete(template.original.id!, template.original)
              }}
            >
              {intl.formatMessage({
                id: `common.delete`,
                defaultMessage: 'Delete'
              })}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </TableCell>
    </TableRow>
  )
}

export const TemplatesDataTable: React.FC<TemplatesTableProps> = ({
  templates,
  table,
  onDelete,
  handleCreate,
  handleEdit,
  total,
  pagination,
  form
}) => {
  const intl = useIntl()

  return (
    <>
      <Paper data-testid="templates-table-container">
        {isNil(templates?.items) || templates?.items.length === 0 ? (
          <EmptyTemplates onCreateTemplate={handleCreate} />
        ) : (
          <TableContainer>
            <Table data-testid="templates-table">
              <TableHeader>
                <TableRow>
                  <TableHead data-testid="templates-table-header-name">
                    {intl.formatMessage({
                      id: 'common.field.name',
                      defaultMessage: 'Name'
                    })}
                  </TableHead>
                  <TableHead data-testid="templates-table-header-type">
                    {intl.formatMessage({
                      id: 'common.field.type',
                      defaultMessage: 'Type'
                    })}
                  </TableHead>
                  <TableHead data-testid="templates-table-header-modified">
                    {intl.formatMessage({
                      id: 'common.field.lastModified',
                      defaultMessage: 'Last Modified'
                    })}
                  </TableHead>
                  <TableHead
                    data-testid="templates-table-header-actions"
                    className="w-0"
                  >
                    {intl.formatMessage({
                      id: 'common.actions',
                      defaultMessage: 'Actions'
                    })}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody data-testid="templates-table-body">
                {table.getRowModel().rows.map((template) => (
                  <TemplateRow
                    key={template.id}
                    template={template}
                    handleEdit={handleEdit}
                    onDelete={onDelete}
                  />
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        )}

        <div className="flex flex-row items-center justify-between border-t px-6 py-3">
          <p className="text-sm text-gray-500 italic">
            {intl.formatMessage(
              {
                id: 'templates.showing',
                defaultMessage:
                  '{number, plural, =0 {No templates found} one {Showing {count} template} other {Showing {count} templates}}.'
              },
              {
                number: templates?.items.length,
                count: (
                  <span className="font-bold">{templates?.items.length}</span>
                )
              }
            )}
          </p>
          <Pagination total={total} {...pagination} />
        </div>
      </Paper>
    </>
  )
}
