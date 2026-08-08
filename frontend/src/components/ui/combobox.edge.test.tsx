import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Combobox } from './combobox'
import type { ComboboxOption } from './combobox'

describe('Combobox Edge Cases', () => {
  describe('Empty states', () => {
    it('renders with empty options array', () => {
      render(<Combobox options={[]} placeholder="Select..." />)
      expect(screen.getByRole('combobox')).toHaveTextContent('Select...')
    })

    it('shows empty text when no options match', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[]} emptyText="No items available" />)

      await user.click(screen.getByRole('combobox'))
      expect(screen.getByText('No items available')).toBeInTheDocument()
    })

    it('shows default empty text when not provided', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[]} />)

      await user.click(screen.getByRole('combobox'))
      expect(screen.getByText('No results found')).toBeInTheDocument()
    })

    it('does not call onChange when clicking in empty dropdown', async () => {
      const onChange = vi.fn()
      const user = userEvent.setup()
      render(<Combobox options={[]} onChange={onChange} />)

      await user.click(screen.getByRole('combobox'))
      // Click on the empty text area
      await user.click(screen.getByText('No results found'))

      expect(onChange).not.toHaveBeenCalled()
    })
  })

  describe('Special characters', () => {
    const specialOptions: ComboboxOption[] = [
      { id: '1', name: "O'Brien & Sons" },
      { id: '2', name: '<script>alert("xss")</script>' },
      { id: '3', name: 'Price: $100.00' },
      { id: '4', name: 'Email: test@example.com' },
      { id: '5', name: 'Path: /api/v1/users' },
    ]

    it('renders special characters correctly', async () => {
      const user = userEvent.setup()
      render(<Combobox options={specialOptions} />)

      await user.click(screen.getByRole('combobox'))

      expect(screen.getByRole('option', { name: "O'Brien & Sons" })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: '<script>alert("xss")</script>' })).toBeInTheDocument()
      expect(screen.getByRole('option', { name: 'Price: $100.00' })).toBeInTheDocument()
    })

    it('selects option with special characters', async () => {
      const onChange = vi.fn()
      const user = userEvent.setup()
      render(<Combobox options={specialOptions} onChange={onChange} />)

      await user.click(screen.getByRole('combobox'))
      await user.click(screen.getByRole('option', { name: '<script>alert("xss")</script>' }))

      expect(onChange).toHaveBeenCalledWith('2')
    })

    it('searches special characters correctly', async () => {
      const user = userEvent.setup()
      render(<Combobox options={specialOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), '$100')

      expect(screen.getByRole('option', { name: 'Price: $100.00' })).toBeInTheDocument()
    })
  })

  describe('Long option lists', () => {
    const manyOptions: ComboboxOption[] = Array.from({ length: 1000 }, (_, i) => ({
      id: String(i),
      name: `Option ${i}`,
    }))

    it('renders many options without performance issues', async () => {
      const user = userEvent.setup()
      render(<Combobox options={manyOptions} />)

      await user.click(screen.getByRole('combobox'))

      // Should show all options
      expect(screen.getAllByRole('option')).toHaveLength(1000)
    })

    it('filters many options correctly', async () => {
      const user = userEvent.setup()
      render(<Combobox options={manyOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), 'Option 1')

      // Should filter to options containing "Option 1"
      const options = screen.getAllByRole('option')
      expect(options.length).toBeGreaterThan(0)
      expect(options.length).toBeLessThan(1000)
    })

    it('selects from many options', async () => {
      const onChange = vi.fn()
      const user = userEvent.setup()
      render(<Combobox options={manyOptions} onChange={onChange} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), 'Option 999')
      await user.click(screen.getByRole('option', { name: 'Option 999' }))

      expect(onChange).toHaveBeenCalledWith('999')
    })
  })

  describe('Disabled state interactions', () => {
    it('does not open dropdown when disabled', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} disabled />)

      await user.click(screen.getByRole('combobox'))

      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })

    it('does not respond to keyboard when disabled', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} disabled />)

      const trigger = screen.getByRole('combobox')
      trigger.focus()
      await user.keyboard('{Enter}')

      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })

    it('shows selected value when disabled', () => {
      render(
        <Combobox
          options={[{ id: '1', name: 'Selected' }]}
          value="1"
          disabled
        />
      )

      expect(screen.getByRole('combobox')).toHaveTextContent('Selected')
    })

    it('does not show clear button when disabled', () => {
      const { container } = render(
        <Combobox
          options={[{ id: '1', name: 'Selected' }]}
          value="1"
          disabled
        />
      )

      // The combobox should not have the clear button (X icon)
      // When disabled, only the ChevronDown should be present
      const svgs = container.querySelectorAll('svg')
      // Should only have ChevronDown, not the X clear button
      expect(svgs.length).toBe(1)
    })
  })

  describe('Invalid state', () => {
    it('shows error styling on trigger', () => {
      render(<Combobox options={[{ id: '1', name: 'Test' }]} invalid />)

      const trigger = screen.getByRole('combobox')
      expect(trigger).toHaveAttribute('aria-invalid', 'true')
    })

    it('shows error message below component', () => {
      render(
        <Combobox
          options={[{ id: '1', name: 'Test' }]}
          error="This field is required"
          id="test"
        />
      )

      expect(screen.getByText('This field is required')).toBeInTheDocument()
      expect(screen.getByText('This field is required')).toHaveAttribute(
        'id',
        'test-error'
      )
    })

    it('does not show error when error prop is empty', () => {
      const { container } = render(
        <Combobox options={[{ id: '1', name: 'Test' }]} error="" id="test" />
      )

      const errorElement = container.querySelector('#test-error')
      expect(errorElement).not.toBeInTheDocument()
    })
  })

  describe('Value edge cases', () => {
    it('handles undefined value', () => {
      render(<Combobox options={[{ id: '1', name: 'Test' }]} value={undefined} />)

      expect(screen.getByRole('combobox')).toHaveTextContent('Select...')
    })

    it('handles null value', () => {
      render(<Combobox options={[{ id: '1', name: 'Test' }]} value={null as any} />)

      expect(screen.getByRole('combobox')).toHaveTextContent('Select...')
    })

    it('handles value not in options', () => {
      render(
        <Combobox
          options={[{ id: '1', name: 'Test' }]}
          value="nonexistent"
          placeholder="Select..."
        />
      )

      // Should show placeholder since value doesn't match any option
      expect(screen.getByRole('combobox')).toHaveTextContent('Select...')
    })

    it('handles duplicate IDs in options', async () => {
      const onChange = vi.fn()
      const user = userEvent.setup()
      const options = [
        { id: '1', name: 'First' },
        { id: '1', name: 'Duplicate' },
      ]

      render(<Combobox options={options} onChange={onChange} />)

      await user.click(screen.getByRole('combobox'))
      const duplicateOption = screen.getByRole('option', { name: 'Duplicate' })
      await user.click(duplicateOption)

      expect(onChange).toHaveBeenCalledWith('1')
    })

    it('handles empty string value', () => {
      render(
        <Combobox
          options={[{ id: '1', name: 'Test' }]}
          value=""
          placeholder="Select..."
        />
      )

      expect(screen.getByRole('combobox')).toHaveTextContent('Select...')
    })
  })

  describe('Search edge cases', () => {
    const searchOptions: ComboboxOption[] = [
      { id: '1', name: 'abc' },
      { id: '2', name: 'ABC' },
      { id: '3', name: 'aBcDeF' },
      { id: '4', name: '123' },
    ]

    it('performs case-insensitive search', async () => {
      const user = userEvent.setup()
      render(<Combobox options={searchOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), 'abc')

      expect(screen.getAllByRole('option')).toHaveLength(3)
    })

    it('searches partial matches', async () => {
      const user = userEvent.setup()
      render(<Combobox options={searchOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), 'de')

      expect(screen.getAllByRole('option')).toHaveLength(1)
      expect(screen.getByRole('option', { name: 'aBcDeF' })).toBeInTheDocument()
    })

    it('searches numbers', async () => {
      const user = userEvent.setup()
      render(<Combobox options={searchOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), '123')

      expect(screen.getAllByRole('option')).toHaveLength(1)
      expect(screen.getByRole('option', { name: '123' })).toBeInTheDocument()
    })

    it('shows all options when search is cleared', async () => {
      const user = userEvent.setup()
      render(<Combobox options={searchOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), 'abc')
      expect(screen.getAllByRole('option')).toHaveLength(3)

      await user.clear(screen.getByRole('searchbox'))
      expect(screen.getAllByRole('option')).toHaveLength(4)
    })

    it('handles search with special regex characters', async () => {
      const regexOptions: ComboboxOption[] = [
        { id: '1', name: 'Price: $100.00' },
        { id: '2', name: 'Price: $200.00' },
      ]

      const user = userEvent.setup()
      render(<Combobox options={regexOptions} />)

      await user.click(screen.getByRole('combobox'))
      await user.type(screen.getByRole('searchbox'), '$100.00')

      expect(screen.getAllByRole('option')).toHaveLength(1)
    })
  })

  describe('Keyboard edge cases', () => {
    it('does nothing with Tab key when closed', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} />)

      const trigger = screen.getByRole('combobox')
      trigger.focus()
      await user.keyboard('{Tab}')

      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })

    it('closes dropdown on Tab key', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} />)

      await user.click(screen.getByRole('combobox'))
      expect(screen.getByRole('listbox')).toBeInTheDocument()

      await user.keyboard('{Tab}')
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })

    it('ArrowUp does not go below index 0', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} />)

      await user.click(screen.getByRole('combobox'))
      await user.keyboard('{ArrowUp}{ArrowUp}{ArrowUp}')

      const option = screen.getByRole('option')
      expect(option).toHaveAttribute('data-highlighted', 'true')
    })

    it('ArrowDown does not go beyond last option', async () => {
      const options = [
        { id: '1', name: 'One' },
        { id: '2', name: 'Two' },
      ]

      const user = userEvent.setup()
      render(<Combobox options={options} />)

      await user.click(screen.getByRole('combobox'))
      // Press ArrowDown many times
      for (let i = 0; i < 10; i++) {
        await user.keyboard('{ArrowDown}')
      }

      const optionElements = screen.getAllByRole('option')
      expect(optionElements[1]).toHaveAttribute('data-highlighted', 'true')
    })
  })

  describe('Click interactions', () => {
    it('toggles dropdown on repeated clicks', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} />)

      await user.click(screen.getByRole('combobox'))
      expect(screen.getByRole('listbox')).toBeInTheDocument()

      await user.click(screen.getByRole('combobox'))
      expect(screen.queryByRole('listbox')).not.toBeInTheDocument()
    })

    it('does not close dropdown when clicking search input', async () => {
      const user = userEvent.setup()
      render(<Combobox options={[{ id: '1', name: 'Test' }]} />)

      await user.click(screen.getByRole('combobox'))
      await user.click(screen.getByRole('searchbox'))

      expect(screen.getByRole('listbox')).toBeInTheDocument()
    })
  })

  describe('Form integration', () => {
    it('works with react-hook-form', async () => {
      const { useForm, Controller } = await import('react-hook-form')

      function TestForm() {
        const { control, watch } = useForm({
          defaultValues: { field: '' },
        })
        const value = watch('field')

        return (
          <div>
            <Controller
              name="field"
              control={control}
              render={({ field }) => (
                <Combobox
                  options={[{ id: '1', name: 'Test' }]}
                  value={field.value}
                  onChange={field.onChange}
                />
              )}
            />
            <span data-testid="value">{value}</span>
          </div>
        )
      }

      const user = userEvent.setup()
      render(<TestForm />)

      await user.click(screen.getByRole('combobox'))
      await user.click(screen.getByRole('option', { name: 'Test' }))

      expect(screen.getByTestId('value')).toHaveTextContent('1')
    })

    it('submits hidden input value', async () => {
      const onSubmit = vi.fn()

      render(
        <form onSubmit={(e) => { e.preventDefault(); onSubmit() }}>
          <Combobox
            options={[{ id: '1', name: 'Test' }]}
            name="testField"
            value="1"
          />
          <button type="submit">Submit</button>
        </form>
      )

      const user = userEvent.setup()
      await user.click(screen.getByText('Submit'))

      expect(onSubmit).toHaveBeenCalled()
    })
  })
})
