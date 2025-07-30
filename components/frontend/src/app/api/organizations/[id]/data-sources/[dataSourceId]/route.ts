import { getController } from '@/lib/http/server'
import { DataSourceController } from '@/core/application/controllers/data-source-controller'

export const GET = getController(DataSourceController, (c) => c.fetchById) 