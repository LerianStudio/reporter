import { getController } from '@/lib/http/server'
import { TemplateController } from '@/core/application/controllers/template-controller'

export const GET = getController(TemplateController, (c) => c.fetchById)

export const PATCH = getController(TemplateController, (c) => c.update)

export const DELETE = getController(TemplateController, (c) => c.delete)
