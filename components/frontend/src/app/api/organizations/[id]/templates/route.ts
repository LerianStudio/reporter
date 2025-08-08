import { getController } from '@/lib/http/server'
import { TemplateController } from '@/core/application/controllers/template-controller'

export const GET = getController(TemplateController, (c) => c.fetchAll)

export const POST = getController(TemplateController, (c) => c.create)
