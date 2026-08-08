import { test, expect, type Page } from '@playwright/test'

// Test page URL (will be served by Vite dev server)
const TEST_PAGE = '/tests/e2e/fixtures/combobox-test-page.html'

test.describe('Combobox Component - Comprehensive E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(TEST_PAGE)
    await page.waitForLoadState('networkidle')
  })

  // ==================== Basic Rendering ====================
  
  test.describe('Rendering', () => {
    test('renders with placeholder text', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await expect(combobox).toBeVisible()
      await expect(combobox).toContainText('Select a user...')
    })

    test('renders with custom placeholder', async ({ page }) => {
      const combobox = page.locator('#many-options-combobox button[role="combobox"]')
      await expect(combobox).toContainText('Select from 100 options...')
    })

    test('shows chevron icon', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      const chevron = combobox.locator('svg.lucide-chevron-down')
      await expect(chevron).toBeVisible()
    })

    test('hides clear button when no value selected', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      const clearButton = combobox.locator('svg.lucide-x')
      await expect(clearButton).not.toBeVisible()
    })
  })

  // ==================== Dropdown Behavior ====================
  
  test.describe('Dropdown Behavior', () => {
    test('opens dropdown on click', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).toBeVisible()
    })

    test('shows all options when opened', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const options = page.locator('#basic-combobox [role="option"]')
      await expect(options).toHaveCount(5)
    })

    test('closes dropdown on outside click', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      // Click outside
      await page.click('body', { position: { x: 0, y: 0 } })
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).not.toBeVisible()
    })

    test('closes dropdown after selection', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      await page.click('#basic-combobox [role="option"]:has-text("Alice Johnson")')
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).not.toBeVisible()
    })

    test('toggles dropdown on repeated clicks', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      
      // Open
      await combobox.click()
      await expect(page.locator('#basic-combobox [role="listbox"]')).toBeVisible()
      
      // Close
      await combobox.click()
      await expect(page.locator('#basic-combobox [role="listbox"]')).not.toBeVisible()
      
      // Open again
      await combobox.click()
      await expect(page.locator('#basic-combobox [role="listbox"]')).toBeVisible()
    })
  })

  // ==================== Search/Filter ====================
  
  test.describe('Search/Filter', () => {
    test('shows search input when dropdown opens', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await expect(searchInput).toBeVisible()
      await expect(searchInput).toHaveAttribute('placeholder', 'Search users...')
    })

    test('filters options based on search input', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await searchInput.fill('alice')
      
      const options = page.locator('#basic-combobox [role="option"]')
      await expect(options).toHaveCount(1)
      await expect(options.first()).toContainText('Alice Johnson')
    })

    test('performs case-insensitive search', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await searchInput.fill('ALICE')
      
      const options = page.locator('#basic-combobox [role="option"]')
      await expect(options).toHaveCount(1)
    })

    test('shows empty text when no options match', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await searchInput.fill('xyz123')
      
      const emptyText = page.locator('#basic-combobox:has-text("No results found")')
      await expect(emptyText).toBeVisible()
    })

    test('shows all options when search is cleared', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await searchInput.fill('alice')
      
      // Clear search
      await searchInput.clear()
      
      const options = page.locator('#basic-combobox [role="option"]')
      await expect(options).toHaveCount(5)
    })

    test('searches partial matches', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await searchInput.fill('son')
      
      const options = page.locator('#basic-combobox [role="option"]')
      await expect(options).toHaveCount(1)
      await expect(options.first()).toContainText('Johnson')
    })
  })

  // ==================== Selection ====================
  
  test.describe('Selection', () => {
    test('selects option on click', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      await page.click('#basic-combobox [role="option"]:has-text("Bob Smith")')
      
      await expect(combobox).toContainText('Bob Smith')
      await expect(page.locator('[data-testid="selected-value"]')).toContainText('2')
    })

    test('shows check icon for selected option', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      // Select an option first
      await page.click('#basic-combobox [role="option"]:has-text("Charlie Brown")')
      
      // Open again and verify check icon
      await combobox.click()
      
      const selectedOption = page.locator('#basic-combobox [role="option"]:has-text("Charlie Brown")')
      const checkIcon = selectedOption.locator('svg.lucide-check')
      await expect(checkIcon).toBeVisible()
    })

    test('updates hidden input value on selection', async ({ page }) => {
      const combobox = page.locator('#form-combobox button[role="combobox"]')
      await combobox.click()
      
      await page.click('#form-combobox [role="option"]:has-text("Diana Ross")')
      
      const hiddenInput = page.locator('#form-combobox input[type="hidden"]')
      await expect(hiddenInput).toHaveValue('4')
    })
  })

  // ==================== Clear Functionality ====================
  
  test.describe('Clear Functionality', () => {
    test('shows clear button when value is selected', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      await page.click('#basic-combobox [role="option"]:has-text("Alice Johnson")')
      
      const clearButton = combobox.locator('svg.lucide-x')
      await expect(clearButton).toBeVisible()
    })

    test('clears selection when clear button clicked', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      
      // Select an option
      await combobox.click()
      await page.click('#basic-combobox [role="option"]:has-text("Alice Johnson")')
      await expect(combobox).toContainText('Alice Johnson')
      
      // Clear selection
      const clearButton = combobox.locator('svg.lucide-x')
      await clearButton.click()
      
      await expect(combobox).toContainText('Select a user...')
      await expect(page.locator('[data-testid="selected-value"]')).toContainText('None')
    })
  })

  // ==================== Keyboard Navigation ====================
  
  test.describe('Keyboard Navigation', () => {
    test('opens dropdown with Enter key', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.focus()
      await page.keyboard.press('Enter')
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).toBeVisible()
    })

    test('closes dropdown with Escape key', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      await page.keyboard.press('Escape')
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).not.toBeVisible()
    })

    test('navigates options with ArrowDown', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      // Navigate down twice
      await page.keyboard.press('ArrowDown')
      await page.keyboard.press('ArrowDown')
      
      // Second option should be highlighted
      const secondOption = page.locator('#basic-combobox [role="option"]').nth(1)
      await expect(secondOption).toHaveAttribute('data-highlighted', 'true')
    })

    test('navigates options with ArrowUp', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      // Navigate down first
      await page.keyboard.press('ArrowDown')
      await page.keyboard.press('ArrowDown')
      await page.keyboard.press('ArrowDown')
      
      // Navigate up
      await page.keyboard.press('ArrowUp')
      
      const secondOption = page.locator('#basic-combobox [role="option"]').nth(1)
      await expect(secondOption).toHaveAttribute('data-highlighted', 'true')
    })

    test('selects highlighted option with Enter', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      // Navigate to first option
      await page.keyboard.press('ArrowDown')
      
      // Select with Enter
      await page.keyboard.press('Enter')
      
      await expect(combobox).toContainText('Alice Johnson')
    })

    test('closes dropdown with Tab key', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      await page.keyboard.press('Tab')
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).not.toBeVisible()
    })
  })

  // ==================== Disabled State ====================
  
  test.describe('Disabled State', () => {
    test('does not open dropdown when disabled', async ({ page }) => {
      const combobox = page.locator('#disabled-combobox button[role="combobox"]')
      await combobox.click()
      
      const listbox = page.locator('#disabled-combobox [role="listbox"]')
      await expect(listbox).not.toBeVisible()
    })

    test('shows disabled styling', async ({ page }) => {
      const combobox = page.locator('#disabled-combobox button[role="combobox"]')
      await expect(combobox).toBeDisabled()
    })

    test('shows selected value when disabled', async ({ page }) => {
      const combobox = page.locator('#disabled-combobox button[role="combobox"]')
      await expect(combobox).toContainText('Bob Smith')
    })
  })

  // ==================== Error State ====================
  
  test.describe('Error State', () => {
    test('shows error message', async ({ page }) => {
      const errorMessage = page.locator('#error-combobox:has-text("This field is required")')
      await expect(errorMessage).toBeVisible()
    })

    test('shows error styling on trigger', async ({ page }) => {
      const combobox = page.locator('#error-combobox button[role="combobox"]')
      await expect(combobox).toHaveAttribute('aria-invalid', 'true')
    })

    test('links error message via aria-describedby', async ({ page }) => {
      const combobox = page.locator('#error-combobox button[role="combobox"]')
      await expect(combobox).toHaveAttribute('aria-describedby', 'error-combobox-error')
    })
  })

  // ==================== Empty Options ====================
  
  test.describe('Empty Options', () => {
    test('shows empty text when no options', async ({ page }) => {
      const combobox = page.locator('#empty-combobox button[role="combobox"]')
      await combobox.click()
      
      const emptyText = page.locator('#empty-combobox:has-text("No items found")')
      await expect(emptyText).toBeVisible()
    })

    test('shows custom empty text', async ({ page }) => {
      const combobox = page.locator('#empty-combobox button[role="combobox"]')
      await combobox.click()
      
      await expect(page.locator('text=No items found')).toBeVisible()
    })
  })

  // ==================== Special Characters ====================
  
  test.describe('Special Characters', () => {
    test('renders special characters correctly', async ({ page }) => {
      const combobox = page.locator('#special-chars-combobox button[role="combobox"]')
      await combobox.click()
      
      await expect(page.locator('#special-chars-combobox [role="option"]:has-text("O\'Brien & Sons")')).toBeVisible()
      await expect(page.locator('#special-chars-combobox [role="option"]:has-text("$100.00")')).toBeVisible()
    })

    test('handles XSS attempts safely', async ({ page }) => {
      const combobox = page.locator('#special-chars-combobox button[role="combobox"]')
      await combobox.click()
      
      const xssOption = page.locator('#special-chars-combobox [role="option"]:has-text("script")')
      await expect(xssOption).toBeVisible()
      
      // Click it - should not execute script
      await xssOption.click()
      
      // Page should still be functional
      await expect(page.locator('body')).toBeVisible()
    })

    test('renders unicode characters', async ({ page }) => {
      const combobox = page.locator('#special-chars-combobox button[role="combobox"]')
      await combobox.click()
      
      await expect(page.locator('#special-chars-combobox [role="option"]:has-text("日本語テスト")')).toBeVisible()
    })
  })

  // ==================== Accessibility ====================
  
  test.describe('Accessibility', () => {
    test('has correct ARIA attributes', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      
      await expect(combobox).toHaveAttribute('role', 'combobox')
      await expect(combobox).toHaveAttribute('aria-haspopup', 'listbox')
      await expect(combobox).toHaveAttribute('aria-expanded', 'false')
    })

    test('updates aria-expanded when opened', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      
      await combobox.click()
      await expect(combobox).toHaveAttribute('aria-expanded', 'true')
      
      await page.keyboard.press('Escape')
      await expect(combobox).toHaveAttribute('aria-expanded', 'false')
    })

    test('listbox has correct aria-label', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const listbox = page.locator('#basic-combobox [role="listbox"]')
      await expect(listbox).toHaveAttribute('aria-label', 'Select a user...')
    })

    test('options have correct aria-selected', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      // Select an option
      await page.click('#basic-combobox [role="option"]:has-text("Charlie Brown")')
      
      // Open again
      await combobox.click()
      
      const selectedOption = page.locator('#basic-combobox [role="option"]:has-text("Charlie Brown")')
      await expect(selectedOption).toHaveAttribute('aria-selected', 'true')
      
      const otherOption = page.locator('#basic-combobox [role="option"]:has-text("Alice Johnson")')
      await expect(otherOption).toHaveAttribute('aria-selected', 'false')
    })

    test('search input has accessible label', async ({ page }) => {
      const combobox = page.locator('#basic-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#basic-combobox input[role="searchbox"]')
      await expect(searchInput).toHaveAttribute('aria-label', 'Search users...')
    })
  })

  // ==================== Large Datasets ====================
  
  test.describe('Large Datasets', () => {
    test('renders 100 options', async ({ page }) => {
      const combobox = page.locator('#many-options-combobox button[role="combobox"]')
      await combobox.click()
      
      const options = page.locator('#many-options-combobox [role="option"]')
      await expect(options).toHaveCount(100)
    })

    test('filters 100 options efficiently', async ({ page }) => {
      const combobox = page.locator('#many-options-combobox button[role="combobox"]')
      await combobox.click()
      
      const searchInput = page.locator('#many-options-combobox input[role="searchbox"]')
      await searchInput.fill('50')
      
      const options = page.locator('#many-options-combobox [role="option"]')
      await expect(options.first()).toContainText('Option 50')
    })
  })
})
