/**
 * Domain models — field names mirror the backend's JSON responses exactly
 * (see backend/internal/domain/*.go). All IDs are UUIDv7 strings.
 */

export interface EntityBase {
  id: string
  created_at: string
  updated_at: string
  deleted_at?: string | null
}

export type UserStatus = 'active' | 'inactive' | 'locked' | 'pending'

export interface User extends EntityBase {
  email: string
  name: string
  status: UserStatus
  last_login_at?: string | null
  roles?: Role[]
}

export interface Role extends EntityBase {
  name: string
  description: string
  is_default: boolean
  permissions?: Permission[]
}

export interface Permission extends EntityBase {
  name: string
  route: string
  path: string
  method: string
  service: string
}

export interface Application extends EntityBase {
  applicant_id: string
  workflow_id: string
  template_id: string
  title?: string
  status: string
  priority: string
  submitted_at?: string | null
  completed_at?: string | null
  data?: Record<string, unknown>
  /** Included when the backend joins related records. */
  applicant?: User
  workflow?: Workflow
  template?: Template
}

export interface Approval extends EntityBase {
  application_id: string
  workflow_step_id: string
  approver_id: string
  status: string
  decision?: string | null
  comment?: string | null
  decided_at?: string | null
  escalation_level: number
  metadata?: Record<string, unknown>
  application?: Application
  approver?: User
}

export interface Workflow extends EntityBase {
  name: string
  description?: string | null
  category: string
  version: number
  is_active: boolean
  steps?: unknown
}

export interface Template extends EntityBase {
  name: string
  category: string
  version: number
  is_active: boolean
  schema?: Record<string, unknown>
  ui?: Record<string, unknown>
}

export interface Escalation extends EntityBase {
  approval_id: string
  level: number
  escalated_to: string
  reason: string
  escalated_at: string
  resolved_at?: string | null
}

export interface Notification extends EntityBase {
  user_id: string
  type: string
  channel: string
  title: string
  body: string
  data?: Record<string, unknown>
  read_at?: string | null
  sent_at?: string | null
}

export interface LoginLog extends EntityBase {
  user_id?: string | null
  email: string
  status: string
  failure_reason?: string | null
  ip_address: string
  user_agent: string
  request_id?: string | null
  token_id?: string | null
  attempted_at: string
}

/** Shape of /auth/login and /auth/refresh responses. */
export interface AuthResult {
  user: User
  access_token: string
  refresh_token: string
  expires_at: string
}
