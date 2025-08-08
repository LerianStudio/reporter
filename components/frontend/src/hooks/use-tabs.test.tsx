import { screen, render, act } from '@testing-library/react'

// Mock the custom useSearchParams hook
const mockSetSearchParams = jest.fn()
const mockUseSearchParams = jest.fn()

jest.mock('../lib/search/use-search-params', () => ({
  useSearchParams: () => mockUseSearchParams()
}))

import { useTabs } from './use-tabs'

function TestComponent() {
  const { activeTab, handleTabChange } = useTabs({
    initialValue: 'tab1'
  })

  return (
    <>
      <p data-testid="activeTab">{activeTab}</p>
      <button data-testid="changeTab" onClick={() => handleTabChange('tab2')} />
    </>
  )
}

function setup(searchParams: any = null) {
  mockUseSearchParams.mockReturnValue({
    searchParams,
    setSearchParams: mockSetSearchParams
  })

  render(<TestComponent />)
  const activeTab = screen.getByTestId('activeTab')
  const button = screen.getByTestId('changeTab')
  return { activeTab, button }
}

describe('useTabs', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  test('should change tabs', async () => {
    const { activeTab, button } = setup()

    expect(activeTab.innerHTML).toEqual('tab1')

    await act(() => {
      button.click()
    })

    expect(activeTab.innerHTML).toEqual('tab2')
    expect(mockSetSearchParams).toHaveBeenCalledWith({ tab: 'tab2' })
  })

  test('should call setSearchParams when tab changes', async () => {
    const { button } = setup()

    await act(() => {
      button.click()
    })

    expect(mockSetSearchParams).toHaveBeenCalledWith({ tab: 'tab2' })
  })

  test('should update activeTab from URL params', async () => {
    const { activeTab } = setup({ tab: 'tab2' })

    // The effect should update the activeTab to match the URL param
    expect(activeTab.innerHTML).toEqual('tab2')
  })
})
