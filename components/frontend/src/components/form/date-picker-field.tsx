import React, { useState, useMemo, useCallback, useRef } from 'react'
import { useIntl } from 'react-intl'
import { Calendar as CalendarIcon, ChevronDown } from 'lucide-react'
import { format } from 'date-fns'
import { Control } from 'react-hook-form'
import { Button } from '@/components/ui/button'
import { Calendar } from '@/components/ui/calendar'
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
  FormTooltip
} from '@/components/ui/form'
import {
  Popover,
  PopoverContent,
  PopoverTrigger
} from '@/components/ui/popover'
import { cn, validateDateString, formatDateForDisplay } from '@/lib/utils'
import { useToast } from '@/hooks/use-toast'
import { ReactNode } from 'react'

export type DatePickerFieldProps = {
  name: string
  label?: ReactNode
  tooltip?: string
  labelExtra?: ReactNode
  placeholder?: string
  description?: ReactNode
  control: Control<any>
  disabled?: boolean
  required?: boolean
  className?: string
}

export const DatePickerField = ({
  name,
  label,
  tooltip,
  labelExtra,
  placeholder,
  description,
  control,
  disabled = false,
  required = false,
  className
}: DatePickerFieldProps) => {
  const intl = useIntl()
  const { toast } = useToast()
  const [isCalendarOpen, setIsCalendarOpen] = useState(false)
  const previousValidationRef = useRef<{ value?: string; hasError: boolean }>({
    hasError: false
  })

  const showValidationError = useCallback(
    (validation: ReturnType<typeof validateDateString>) => {
      if (!validation.error) return

      const errorMessages = {
        format: {
          title: 'reports.filters.invalidDateFormat',
          description: 'reports.filters.invalidDateDescription'
        },
        invalid: {
          title: 'reports.filters.invalidDate',
          description: 'reports.filters.invalidDateValue'
        },
        range: {
          title: 'reports.filters.dateOutOfRange',
          description: 'reports.filters.dateRangeDescription'
        },
        parsing: {
          title: 'reports.filters.dateParsingError',
          description: 'reports.filters.dateParsingErrorDescription'
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
    },
    [intl, toast]
  )

  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => {
        const handleDateSelect = useCallback(
          (date: Date | undefined) => {
            if (date) {
              field.onChange(format(date, 'yyyy-MM-dd'))
              setIsCalendarOpen(false)
            } else {
              field.onChange(undefined)
            }
          },
          [field, setIsCalendarOpen]
        )

        const selectedDate = useMemo(() => {
          const validation = validateDateString(field.value)

          if (!validation.isValid && validation.error && field.value) {
            const previous = previousValidationRef.current
            const shouldShowToast =
              previous.value !== field.value || !previous.hasError

            if (shouldShowToast) {
              setTimeout(() => showValidationError(validation), 0)
            }

            previousValidationRef.current = {
              value: field.value,
              hasError: true
            }

            return undefined
          }

          previousValidationRef.current = {
            value: field.value,
            hasError: false
          }

          return validation.date
        }, [field.value, showValidationError])

        return (
          <FormItem required={required} className={className}>
            {label && (
              <FormLabel
                extra={
                  tooltip ? <FormTooltip>{tooltip}</FormTooltip> : labelExtra
                }
              >
                {label}
              </FormLabel>
            )}
            <FormControl>
              <Popover open={isCalendarOpen} onOpenChange={setIsCalendarOpen}>
                <PopoverTrigger asChild>
                  <Button
                    variant="outline"
                    disabled={disabled}
                    className={cn(
                      'border-border h-9 w-fit justify-between gap-3 rounded-md border bg-white px-3 py-2 text-left text-sm font-normal hover:bg-white focus-visible:ring-2 focus-visible:ring-offset-0',
                      !selectedDate && 'placeholder:text-shadcn-400'
                    )}
                  >
                    <div className="flex items-center">
                      <CalendarIcon className="mr-2 h-4 w-4" />
                      {selectedDate ? (
                        formatDateForDisplay(selectedDate)
                      ) : (
                        <span>
                          {placeholder ||
                            intl.formatMessage({
                              id: 'reports.filters.selectDate',
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
                    disabled={disabled}
                  />
                </PopoverContent>
              </Popover>
            </FormControl>
            <FormMessage />
            {description && <FormDescription>{description}</FormDescription>}
          </FormItem>
        )
      }}
    />
  )
}
