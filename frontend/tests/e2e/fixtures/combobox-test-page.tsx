import React from 'react'
import { createRoot } from 'react-dom/client'
import { Combobox } from '@/components/ui/combobox'
import '@/index.css'

const options = [
  { id: '1', name: 'Alice Johnson' },
  { id: '2', name: 'Bob Smith' },
  { id: '3', name: 'Charlie Brown' },
  { id: '4', name: 'Diana Ross' },
  { id: '5', name: 'Edward Norton' },
]

const App = () => {
  const [value, setValue] = React.useState('')
  const [multiValue, setMultiValue] = React.useState('')
  const [formValue, setFormValue] = React.useState('')

  return (
    <div style={{ padding: '20px', maxWidth: '400px' }}>
      <h1>Combobox E2E Test Page</h1>

      <div style={{ marginBottom: '20px' }}>
        <h2>Basic Combobox</h2>
        <Combobox
          id="basic-combobox"
          options={options}
          value={value}
          onChange={setValue}
          placeholder="Select a user..."
          searchPlaceholder="Search users..."
        />
        <p data-testid="selected-value">Selected: {value || 'None'}</p>
      </div>

      <div style={{ marginBottom: '20px' }}>
        <h2>Disabled Combobox</h2>
        <Combobox
          id="disabled-combobox"
          options={options}
          value="2"
          disabled
          placeholder="Disabled select..."
        />
      </div>

      <div style={{ marginBottom: '20px' }}>
        <h2>Error Combobox</h2>
        <Combobox
          id="error-combobox"
          options={options}
          value=""
          invalid
          error="This field is required"
          placeholder="Select with error..."
        />
      </div>

      <div style={{ marginBottom: '20px' }}>
        <h2>Empty Options Combobox</h2>
        <Combobox
          id="empty-combobox"
          options={[]}
          value=""
          placeholder="No options available..."
          emptyText="No items found"
        />
      </div>

      <div style={{ marginBottom: '20px' }}>
        <h2>Many Options Combobox</h2>
        <Combobox
          id="many-options-combobox"
          options={Array.from({ length: 100 }, (_, i) => ({
            id: String(i),
            name: `Option ${i}`,
          }))}
          value={multiValue}
          onChange={setMultiValue}
          placeholder="Select from 100 options..."
        />
      </div>

      <div style={{ marginBottom: '20px' }}>
        <h2>Special Characters Combobox</h2>
        <Combobox
          id="special-chars-combobox"
          options={[
            { id: '1', name: "O'Brien & Sons" },
            { id: '2', name: '<script>alert("xss")</script>' },
            { id: '3', name: 'Price: $100.00' },
            { id: '4', name: '日本語テスト' },
          ]}
          placeholder="Select special chars..."
        />
      </div>

      <div>
        <h2>Form Integration</h2>
        <form id="test-form">
          <Combobox
            id="form-combobox"
            name="user_id"
            options={options}
            value={formValue}
            onChange={setFormValue}
            placeholder="Select user for form..."
          />
          <button type="submit">Submit</button>
        </form>
      </div>
    </div>
  )
}

createRoot(document.getElementById('root')!).render(<App />)
