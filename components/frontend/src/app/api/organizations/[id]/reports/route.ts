import { getController } from '@/lib/http/server'
import { ReportController } from '@/core/application/controllers/report-controller'

export const GET = getController(ReportController, (c) => c.fetchAll)

export const POST = getController(ReportController, (c) => c.create)
