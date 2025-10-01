import { ServerFactory } from '@lerianstudio/sindarian-server'
import { AppModule } from './modules/app-module'

export const app = ServerFactory.create(AppModule)
app.setGlobalPrefix('/api')
