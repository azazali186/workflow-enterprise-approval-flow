import { useEffect } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { AlertCircle, CheckCircle2, Gauge, ShieldCheck, Workflow, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { FormField } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useAppSelector } from '@/store/hooks'
import { useLogin } from '@/features/auth/hooks/use-login'
import { toErrorMessage } from '@/services/api/errors'

const schema = z.object({
  email: z.string().email('Enter a valid email address'),
  password: z.string().min(1, 'Password is required'),
})

type FormValues = z.infer<typeof schema>

const features = [
  { icon: Workflow, text: 'Design approval workflows in minutes' },
  { icon: ShieldCheck, text: 'Role-based access, enforced end to end' },
  { icon: Gauge, text: 'Real-time analytics and escalations' },
]

export default function LoginPage() {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated)
  const login = useLogin()
  const navigate = useNavigate()
  const location = useLocation()

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: '', password: '' },
  })

  useEffect(() => {
    if (login.isSuccess && login.data) {
      const from = (location.state as { from?: string } | null)?.from ?? '/'
      navigate(from, { replace: true })
    }
  }, [login.isSuccess, login.data, location.state, navigate])

  if (isAuthenticated) return <Navigate to="/" replace />

  const onSubmit = (values: FormValues) => login.mutate(values)

  return (
    <div className="flex min-h-dvh">
      {/* Brand panel */}
      <div className="relative hidden w-1/2 overflow-hidden bg-slate-950 lg:block">
        <div
          className="absolute inset-0"
          style={{
            background:
              'radial-gradient(60rem 60rem at 15% -10%, rgb(79 70 229 / 0.35), transparent 60%), radial-gradient(50rem 50rem at 110% 30%, rgb(124 58 237 / 0.28), transparent 55%), radial-gradient(40rem 40rem at 50% 120%, rgb(14 165 233 / 0.18), transparent 60%)',
          }}
        />
        <div className="relative flex h-full flex-col justify-between p-12">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-violet-600 shadow-lg">
              <Zap className="h-5 w-5 text-white" fill="currentColor" />
            </div>
            <div>
              <p className="text-sm font-bold text-white">Approval Flow</p>
              <p className="text-[10px] font-medium uppercase tracking-widest text-slate-400">
                Admin Console
              </p>
            </div>
          </div>

          <div>
            <motion.h1
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.1 }}
              className="max-w-md text-4xl font-bold leading-tight tracking-tight text-white"
            >
              Enterprise workflow,{' '}
              <span className="bg-gradient-to-r from-primary-400 to-violet-400 bg-clip-text text-transparent">
                under control.
              </span>
            </motion.h1>
            <motion.p
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.2 }}
              className="mt-4 max-w-md text-sm leading-relaxed text-slate-400"
            >
              Approve, escalate and track every business process from a single
              console — with granular roles and real-time visibility.
            </motion.p>

            <motion.ul
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.3 }}
              className="mt-8 space-y-3"
            >
              {features.map(({ icon: Icon, text }) => (
                <li key={text} className="flex items-center gap-3 text-sm text-slate-300">
                  <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-white/5 ring-1 ring-white/10">
                    <Icon className="h-4 w-4 text-primary-400" />
                  </span>
                  {text}
                </li>
              ))}
            </motion.ul>
          </div>

          <p className="text-xs text-slate-500">
            © {new Date().getFullYear()} Approval Flow Enterprise
          </p>
        </div>
      </div>

      {/* Form panel */}
      <div className="flex flex-1 items-center justify-center px-4 py-12 sm:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.4 }}
          className="w-full max-w-sm"
        >
          <div className="mb-8 flex items-center gap-2.5 lg:hidden">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-violet-600 shadow-lg">
              <Zap className="h-5 w-5 text-white" fill="currentColor" />
            </div>
            <div>
              <p className="text-sm font-bold text-slate-900">Approval Flow</p>
              <p className="text-[10px] font-medium uppercase tracking-widest text-slate-400">
                Admin Console
              </p>
            </div>
          </div>

          <h2 className="text-2xl font-bold tracking-tight text-slate-900">Sign in</h2>
          <p className="mt-1 text-sm text-slate-500">
            Enter your credentials to access the console.
          </p>

          <form onSubmit={handleSubmit(onSubmit)} className="mt-8 space-y-4" noValidate>
            {login.isError && (
              <div
                className="flex items-start gap-2.5 rounded-lg border border-rose-200 bg-rose-50 px-3.5 py-3 text-sm text-rose-700"
                role="alert"
              >
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
                <span>{toErrorMessage(login.error, 'Sign in failed. Please try again.')}</span>
              </div>
            )}

            <FormField label="Email address" htmlFor="email" required error={errors.email?.message}>
              <Input
                id="email"
                type="email"
                autoComplete="email"
                placeholder="you@company.com"
                invalid={Boolean(errors.email)}
                {...register('email')}
              />
            </FormField>

            <FormField label="Password" htmlFor="password" required error={errors.password?.message}>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                invalid={Boolean(errors.password)}
                {...register('password')}
              />
            </FormField>

            <Button type="submit" size="lg" className="w-full" loading={login.isPending}>
              {login.isPending ? 'Signing in…' : 'Sign in'}
            </Button>
          </form>

          <div className="mt-6 flex items-center justify-center gap-2 text-xs text-slate-400">
            <CheckCircle2 className="h-3.5 w-3.5" />
            Secured with encrypted credentials and session tokens
          </div>
        </motion.div>
      </div>
    </div>
  )
}
