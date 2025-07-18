import dotenv from 'dotenv'

dotenv.config({ path: './.env.playwright' })

export const { BASE_HOST, BASE_PORT } = process.env

export const BASE_URL = `http://${BASE_HOST}:${BASE_PORT}`
