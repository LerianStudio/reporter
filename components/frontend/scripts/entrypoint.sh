#!/bin/sh

# Gera o arquivo public/runtime-env.js com as envs do container
echo "window.RUNTIME_ENV = $(node -p 'JSON.stringify({
  NEXT_PUBLIC_PLUGIN_UI_BASE_PATH: process.env.NEXT_PUBLIC_PLUGIN_UI_BASE_PATH,
  NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL: process.env.NEXT_PUBLIC_MIDAZ_CONSOLE_BASE_URL,
  NEXT_PUBLIC_MIDAZ_AUTH_ENABLED: process.env.NEXT_PUBLIC_MIDAZ_AUTH_ENABLED,
  NEXT_PUBLIC_NEXTAUTH_URL: process.env.NEXT_PUBLIC_NEXTAUTH_URL
})');" > ./public/runtime-env.js

# Executa o comando original (start do Next.js, ou qualquer comando passado)
exec "$@"
