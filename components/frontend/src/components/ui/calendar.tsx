'use client'

import * as React from 'react'
import {
  ChevronDownIcon,
  ChevronLeftIcon,
  ChevronRightIcon
} from 'lucide-react'
import { DayButton, DayPicker, getDefaultClassNames } from 'react-day-picker'

import { cn } from '@/lib/utils'
import { Button, buttonVariants } from '@/components/ui/button'

function Calendar({
  className,
  classNames,
  showOutsideDays = true,
  captionLayout = 'label',
  buttonVariant = 'ghost',
  formatters,
  components,
  ...props
}: React.ComponentProps<typeof DayPicker> & {
  buttonVariant?: React.ComponentProps<typeof Button>['variant']
}) {
  const defaultClassNames = getDefaultClassNames()

  return (
    <DayPicker
      showOutsideDays={showOutsideDays}
      className={cn(
        'group/calendar rounded-lg bg-white p-3 shadow-lg [--cell-size:2.5rem] dark:bg-slate-950 [[data-slot=card-content]_&]:bg-transparent [[data-slot=popover-content]_&]:bg-transparent',
        String.raw`rtl:**:[.rdp-button\_next>svg]:rotate-180`,
        String.raw`rtl:**:[.rdp-button\_previous>svg]:rotate-180`,
        className
      )}
      captionLayout={captionLayout}
      formatters={{
        formatMonthDropdown: (date) =>
          date.toLocaleString('default', { month: 'short' }),
        ...formatters
      }}
      classNames={{
        root: cn('w-full min-w-fit', defaultClassNames.root),
        months: cn(
          'flex gap-4 flex-col md:flex-row relative',
          defaultClassNames.months
        ),
        month: cn(
          'flex flex-col w-full gap-4 min-w-fit',
          defaultClassNames.month
        ),
        nav: cn(
          'flex items-center gap-1 w-full absolute top-0 inset-x-0 justify-between',
          defaultClassNames.nav
        ),
        button_previous: cn(
          buttonVariants({ variant: buttonVariant }),
          'size-(--cell-size) aria-disabled:opacity-50 p-0 select-none hover:bg-transparent hover:text-current',
          defaultClassNames.button_previous
        ),
        button_next: cn(
          buttonVariants({ variant: buttonVariant }),
          'size-(--cell-size) aria-disabled:opacity-50 p-0 select-none hover:bg-transparent hover:text-current',
          defaultClassNames.button_next
        ),
        month_caption: cn(
          'flex items-center justify-center h-(--cell-size) w-full px-(--cell-size)',
          defaultClassNames.month_caption
        ),
        dropdowns: cn(
          'w-full flex items-center text-sm font-medium justify-center h-(--cell-size) gap-1.5',
          defaultClassNames.dropdowns
        ),
        dropdown_root: cn(
          'relative has-focus:border-slate-950 border border-slate-200 shadow-xs has-focus:ring-slate-950/50 has-focus:ring-[3px] rounded-md dark:has-focus:border-slate-300 dark:border-slate-800 dark:has-focus:ring-slate-300/50',
          defaultClassNames.dropdown_root
        ),
        dropdown: cn(
          'absolute bg-white inset-0 opacity-0 dark:bg-slate-950',
          defaultClassNames.dropdown
        ),
        caption_label: cn(
          'select-none font-medium',
          captionLayout === 'label'
            ? 'text-sm'
            : 'rounded-md pl-2 pr-1 flex items-center gap-1 text-sm h-8 [&>svg]:text-slate-500 [&>svg]:size-3.5 dark:[&>svg]:text-slate-400',
          defaultClassNames.caption_label
        ),
        table: 'w-full border-collapse',
        weekdays: cn('flex gap-1', defaultClassNames.weekdays),
        weekday: cn(
          'text-slate-500 rounded-md flex-1 font-normal text-[0.8rem] select-none dark:text-slate-400',
          defaultClassNames.weekday
        ),
        week: cn('flex w-full mt-2 gap-1', defaultClassNames.week),
        week_number_header: cn(
          'select-none w-(--cell-size)',
          defaultClassNames.week_number_header
        ),
        week_number: cn(
          'text-[0.8rem] select-none text-slate-500 dark:text-slate-400',
          defaultClassNames.week_number
        ),
        day: cn(
          'relative w-full h-full p-0 text-center [&:first-child[data-selected=true]_button]:rounded-l-md [&:last-child[data-selected=true]_button]:rounded-r-md group/day aspect-square select-none',
          defaultClassNames.day
        ),
        range_start: cn(
          'rounded-l-md bg-accent',
          defaultClassNames.range_start
        ),
        range_middle: cn('rounded-none', defaultClassNames.range_middle),
        range_end: cn('rounded-r-md bg-accent', defaultClassNames.range_end),
        today: cn(
          'bg-accent/20 text-accent-foreground rounded-md data-[selected=true]:rounded-none',
          defaultClassNames.today
        ),
        outside: cn(
          'text-slate-500 aria-selected:text-slate-500 dark:text-slate-400 dark:aria-selected:text-slate-400',
          defaultClassNames.outside
        ),
        disabled: cn(
          'text-slate-500 opacity-50 dark:text-slate-400',
          defaultClassNames.disabled
        ),
        hidden: cn('invisible', defaultClassNames.hidden),
        ...classNames
      }}
      components={{
        Root: ({ className, rootRef, ...props }) => {
          return (
            <div
              data-slot="calendar"
              ref={rootRef}
              className={cn(className)}
              {...props}
            />
          )
        },
        Chevron: ({ className, orientation, ...props }) => {
          if (orientation === 'left') {
            return (
              <ChevronLeftIcon className={cn('size-4', className)} {...props} />
            )
          }

          if (orientation === 'right') {
            return (
              <ChevronRightIcon
                className={cn('size-4', className)}
                {...props}
              />
            )
          }

          return (
            <ChevronDownIcon className={cn('size-4', className)} {...props} />
          )
        },
        DayButton: CalendarDayButton,
        WeekNumber: ({ children, ...props }) => {
          return (
            <td {...props}>
              <div className="flex size-(--cell-size) items-center justify-center text-center">
                {children}
              </div>
            </td>
          )
        },
        ...components
      }}
      {...props}
    />
  )
}

function CalendarDayButton({
  className,
  day,
  modifiers,
  ...props
}: React.ComponentProps<typeof DayButton>) {
  const defaultClassNames = getDefaultClassNames()

  const ref = React.useRef<HTMLButtonElement>(null)
  const mountedRef = React.useRef(true)

  React.useEffect(() => {
    // Cleanup function to track component mount status
    return () => {
      mountedRef.current = false
    }
  }, [])

  React.useEffect(() => {
    if (modifiers.focused && mountedRef.current) {
      // Use requestAnimationFrame to ensure DOM is ready and prevent race conditions
      const focusElement = () => {
        if (mountedRef.current && ref.current) {
          try {
            ref.current.focus({ preventScroll: false })
          } catch (error) {
            // Silently handle focus errors (element might be removed from DOM)
            console.debug('Focus error in calendar day button:', error)
          }
        }
      }

      const animationId = requestAnimationFrame(focusElement)

      // Cleanup function to cancel pending focus operation
      return () => {
        cancelAnimationFrame(animationId)
      }
    }
  }, [modifiers.focused])

  const isSelected =
    modifiers.selected &&
    !modifiers.range_start &&
    !modifiers.range_end &&
    !modifiers.range_middle
  const dayNumber = day.date.getDate()
  const monthName = day.date.toLocaleDateString('en', { month: 'long' })
  const year = day.date.getFullYear()

  return (
    <Button
      ref={ref}
      variant="ghost"
      size="icon"
      data-day={day.date.toLocaleDateString()}
      data-selected-single={isSelected}
      data-range-start={modifiers.range_start}
      data-range-end={modifiers.range_end}
      data-range-middle={modifiers.range_middle}
      // Enhanced accessibility attributes
      aria-label={`${monthName} ${dayNumber}, ${year}`}
      aria-selected={modifiers.selected}
      aria-current={modifiers.today ? 'date' : undefined}
      aria-disabled={modifiers.disabled}
      role="gridcell"
      tabIndex={modifiers.focused ? 0 : -1}
      className={cn(
        'data-[range-end=true]:bg-accent data-[range-end=true]:text-accent-foreground data-[range-start=true]:bg-accent data-[range-start=true]:text-accent-foreground data-[selected-single=true]:bg-accent data-[selected-single=true]:text-accent-foreground data-[range-middle=true]:bg-accent/20 data-[range-middle=true]:text-accent-foreground hover:bg-accent/20 hover:text-accent-foreground focus-visible:ring-ring focus-visible:ring-offset-background flex aspect-square size-auto w-full min-w-(--cell-size) flex-col gap-1 !border-0 leading-none font-normal focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none data-[range-end=true]:rounded-md data-[range-end=true]:rounded-r-md data-[range-middle=true]:rounded-none data-[range-start=true]:rounded-md data-[range-start=true]:rounded-l-md [&>span]:text-xs [&>span]:opacity-70',
        defaultClassNames.day,
        className
      )}
      {...props}
    />
  )
}

export { Calendar, CalendarDayButton }
