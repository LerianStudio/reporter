'use client'

import React from 'react'
import { isNil } from 'lodash'
import { useSearchParams } from '@/lib/search/use-search-params'

export type UseTabsProps = {
  initialValue?: string
  onTabChange?: (tab: string) => void
}

/**
 * Hook designed to simplify usage of Tabs together with other components
 * that needs this information. Ex: Breadcrumb
 * @param param0
 * @returns
 */
export const useTabs = ({ initialValue, onTabChange }: UseTabsProps) => {
  const { searchParams, updateSearchParams } = useSearchParams()
  const [activeTab, setActiveTab] = React.useState(initialValue || '')

  /**
   * Update state and route application with the respective tab as a URL search param
   * @param tab
   */
  const handleTabChange = (tab: string) => {
    setActiveTab(tab)
    updateSearchParams({ tab })
    onTabChange?.(tab)
  }

  /**
   * Updates activeTab when changed from URL parameters
   */
  React.useEffect(() => {
    const tab = searchParams?.tab

    // Avoid if no tab params is found
    if (isNil(tab)) {
      return
    }

    setActiveTab(tab)
  }, [searchParams])

  return {
    activeTab,
    handleTabChange
  }
}
