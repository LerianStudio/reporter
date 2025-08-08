import NextAuth from 'next-auth'

declare module 'next-auth' {
  interface Session {
    user: {
      id: string
      username: string
      organizationId: string
      access_token: string
    }
  }

  interface User {
    id: string
    username: string
    organizationId: string
    access_token: string
  }
}

declare module 'next-auth/jwt' {
  interface JWT {
    id: string
    username: string
    organizationId: string
    access_token: string
  }
}
