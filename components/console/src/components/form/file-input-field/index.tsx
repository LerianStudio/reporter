import React, { ReactNode, useRef } from 'react'
import { Control } from 'react-hook-form'
import { useIntl } from 'react-intl'
import { Upload, X } from 'lucide-react'
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormTooltip
} from '@/components/ui/form'
import { Button } from '@/components/ui/button'

export type FileInputFieldProps = {
  name: string
  label?: ReactNode
  tooltip?: string
  accept?: string
  description?: ReactNode
  control: Control<any>
  disabled?: boolean
  required?: boolean
  maxSize?: number // in bytes
  onFileChange?: (file: File | null) => void
}

export const FileInputField = ({
  name,
  label,
  tooltip,
  accept = '.tpl',
  description,
  control,
  disabled = false,
  required = false,
  maxSize = 5 * 1024 * 1024, // 5MB default
  onFileChange
}: FileInputFieldProps) => {
  const intl = useIntl()
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = () => {
    fileInputRef.current?.click()
  }

  const handleFileChange = (
    event: React.ChangeEvent<HTMLInputElement>,
    onChange: (file: File | null) => void
  ) => {
    const file = event.target.files?.[0] || null
    onChange(file)
    onFileChange?.(file)
  }

  const handleRemoveFile = (onChange: (file: File | null) => void) => {
    onChange(null)
    onFileChange?.(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes < 1024) {
      return intl.formatMessage(
        { id: 'fileInput.fileSize.bytes', defaultMessage: '{bytes} B' },
        { bytes }
      )
    }
    if (bytes < 1024 * 1024) {
      const kb = (bytes / 1024).toFixed(1)
      return intl.formatMessage(
        { id: 'fileInput.fileSize.kilobytes', defaultMessage: '{kb} KB' },
        { kb }
      )
    }
    const mb = (bytes / (1024 * 1024)).toFixed(1)
    return intl.formatMessage(
      { id: 'fileInput.fileSize.megabytes', defaultMessage: '{mb} MB' },
      { mb }
    )
  }

  return (
    <FormField
      name={name}
      control={control}
      render={({ field: { onChange, value, ...field } }) => (
        <FormItem required={required}>
          {label && (
            <FormLabel
              extra={tooltip ? <FormTooltip>{tooltip}</FormTooltip> : undefined}
            >
              {label}
            </FormLabel>
          )}
          <FormControl>
            <div className="space-y-2">
              {/* Hidden file input */}
              <input
                ref={fileInputRef}
                type="file"
                accept={accept}
                onChange={(e) => handleFileChange(e, onChange)}
                className="hidden"
                disabled={disabled}
              />

              {/* Custom file input area */}
              <div
                className={`relative flex min-h-[44px] w-full items-center rounded-md border border-zinc-300 bg-white px-3 py-2 text-sm ${disabled ? 'cursor-not-allowed opacity-50' : 'cursor-pointer hover:border-zinc-400'} transition-colors duration-200`}
                onClick={!disabled ? handleFileSelect : undefined}
              >
                <div className="flex flex-1 items-center gap-2">
                  {value ? (
                    <>
                      <Upload className="h-4 w-4 text-zinc-600" />
                      <div className="flex flex-1 items-center justify-between">
                        <div className="flex flex-col">
                          <span className="font-medium text-zinc-900">
                            {value.name}
                          </span>
                          <span className="text-xs text-zinc-500">
                            {formatFileSize(value.size)}
                          </span>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="h-6 w-6 p-0 hover:bg-zinc-100"
                          onClick={(e) => {
                            e.stopPropagation()
                            handleRemoveFile(onChange)
                          }}
                          disabled={disabled}
                        >
                          <X className="h-3 w-3" />
                        </Button>
                      </div>
                    </>
                  ) : (
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-zinc-500">
                        {intl.formatMessage({
                          id: 'fileInput.chooseFile',
                          defaultMessage: 'Choose File'
                        })}
                      </span>
                      <span className="text-zinc-400">
                        {intl.formatMessage({
                          id: 'fileInput.noFileChosen',
                          defaultMessage: 'No file chosen'
                        })}
                      </span>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </FormControl>
          <FormMessage />
          {description && (
            <FormDescription className="text-zinc-500">
              {description}
            </FormDescription>
          )}
        </FormItem>
      )}
    />
  )
}
