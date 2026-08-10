import { useQuery } from '@tanstack/react-query'
import { dropdownsService } from '@/services/dropdowns.service'
import type { ComboboxOption } from '@/components/ui/combobox'

// Re-export for convenience
export type DropdownOption = ComboboxOption

interface UseDropdownOptions {
  enabled?: boolean
}

/**
 * Hook to fetch dropdown options for users.
 */
export function useUserDropdown(options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'users'],
    queryFn: () => dropdownsService.users(),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000, // 5 minutes
  })
}

/**
 * Hook to fetch dropdown options for workflows.
 */
export function useWorkflowDropdown(includeInactive = false, options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'workflows', includeInactive],
    queryFn: () => dropdownsService.workflows(includeInactive),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * Hook to fetch dropdown options for templates.
 */
export function useTemplateDropdown(options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'templates'],
    queryFn: () => dropdownsService.templates(),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * Hook to fetch dropdown options for roles.
 */
export function useRoleDropdown(options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'roles'],
    queryFn: () => dropdownsService.roles(),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * Hook to fetch dropdown options for applications.
 */
export function useApplicationDropdown(statuses?: string[], options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'applications', statuses],
    queryFn: () => dropdownsService.applications(statuses),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * Hook to fetch dropdown options for approvals.
 */
export function useApprovalDropdown(options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'approvals'],
    queryFn: () => dropdownsService.approvals(),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * Hook to fetch the steps of a single workflow (id/name options). Disabled
 * until a workflow is selected, so the dropdown never shows an empty list.
 */
export function useWorkflowStepDropdown(workflowId: string | undefined, options?: UseDropdownOptions) {
  return useQuery({
    queryKey: ['dropdowns', 'workflow_steps', workflowId],
    queryFn: () => dropdownsService.workflowSteps(workflowId!),
    enabled: (options?.enabled ?? true) && Boolean(workflowId),
    staleTime: 5 * 60 * 1000,
  })
}

/**
 * Hook to fetch multiple dropdown options in a single request.
 */
export function useMultipleDropdowns(
  entities: string[],
  options?: UseDropdownOptions & { include_inactive?: boolean; statuses?: string[] }
) {
  return useQuery({
    queryKey: ['dropdowns', 'multiple', entities, options],
    queryFn: () =>
      dropdownsService.multiple(entities, {
        include_inactive: options?.include_inactive,
        statuses: options?.statuses,
      }),
    enabled: options?.enabled ?? true,
    staleTime: 5 * 60 * 1000,
  })
}
