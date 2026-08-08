import { useState } from 'react'
import { FileType, Pencil, Plus, Trash2 } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { DataTable } from '@/components/data-table/data-table'
import type { ColumnDef } from '@/components/data-table/types'
import { PageHeader } from '@/components/ui/page-header'
import { useTemplatesTable, useTemplateMutations } from '@/features/templates/hooks/use-templates'
import { TemplateFormModal } from '@/features/templates/components/template-form-modal'
import { formatDate } from '@/utils/format'
import type { Template } from '@/types/models'

export default function TemplatesPage() {
  const table = useTemplatesTable()
  const { deleteTemplate } = useTemplateMutations()
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Template | null>(null)
  const [deleting, setDeleting] = useState<Template | null>(null)

  const columns: ColumnDef<Template>[] = [
    {
      key: 'name',
      header: 'Template',
      sortable: true,
      render: (template) => (
        <div className="flex items-center gap-2.5">
          <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-violet-50 text-violet-600">
            <FileType className="h-4 w-4" />
          </span>
          <p className="font-medium text-slate-900">{template.name}</p>
        </div>
      ),
    },
    {
      key: 'category',
      header: 'Category',
      render: (template) => <Badge variant="neutral">{template.category}</Badge>,
    },
    {
      key: 'version',
      header: 'Version',
      hideBelow: 'sm',
      render: (template) => (
        <span className="rounded-md bg-slate-100 px-1.5 py-0.5 font-mono text-xs text-slate-600">
          v{template.version}
        </span>
      ),
    },
    {
      key: 'is_active',
      header: 'Active',
      hideBelow: 'md',
      render: (template) =>
        template.is_active ? <Badge variant="success">Active</Badge> : <Badge variant="neutral">Inactive</Badge>,
    },
    {
      key: 'created_at',
      header: 'Created',
      hideBelow: 'lg',
      sortable: true,
      render: (template) => <span className="text-slate-500">{formatDate(template.created_at)}</span>,
    },
    {
      key: 'actions',
      header: '',
      align: 'right',
      render: (template) => (
        <div className="flex items-center justify-end gap-1" onClick={(event) => event.stopPropagation()}>
          <Button
            variant="ghost"
            size="icon"
            title="Edit template"
            onClick={() => {
              setEditing(template)
              setFormOpen(true)
            }}
          >
            <Pencil className="h-4 w-4" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            title="Delete template"
            className="text-slate-400 hover:bg-rose-50 hover:text-rose-600"
            onClick={() => setDeleting(template)}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <div className="space-y-6">
      <PageHeader
        eyebrow="Workflows"
        title="Templates"
        description="Reusable form schemas for applications."
        actions={
          <Button
            onClick={() => {
              setEditing(null)
              setFormOpen(true)
            }}
          >
            <Plus className="h-4 w-4" />
            <span className="hidden sm:inline">New template</span>
          </Button>
        }
      />

      <DataTable
        table={table}
        columns={columns}
        rowKey={(template) => template.id}
        noun="templates"
        searchPlaceholder="Search templates…"
      />

      <TemplateFormModal open={formOpen} onClose={() => setFormOpen(false)} template={editing} />
      <ConfirmDialog
        open={Boolean(deleting)}
        title="Delete template?"
        description={deleting ? `“${deleting.name}” will be removed from the library.` : undefined}
        confirmLabel="Delete template"
        loading={deleteTemplate.isPending}
        onConfirm={() => {
          if (deleting) deleteTemplate.mutate(deleting.id, { onSettled: () => setDeleting(null) })
        }}
        onCancel={() => setDeleting(null)}
      />
    </div>
  )
}
