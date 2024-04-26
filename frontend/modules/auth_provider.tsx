import { useState, createContext, useEffect } from 'react'
import { useRouter } from 'next/router'

export type UserInfo = {
  nickname: string
  id: string
  refresh_token: string
  access_token: string
}

export const AuthContext = createContext<{
  authenticated: boolean
  setAuthenticated: (auth: boolean) => void
  user: UserInfo
  setUser: (user: UserInfo) => void
}>({
  authenticated: false,
  setAuthenticated: () => {},
  user: { nickname: '', id: '', refresh_token:'', access_token:'' },
  setUser: () => {},
})

const AuthContextProvider = ({ children }: { children: React.ReactNode }) => {
  const [authenticated, setAuthenticated] = useState(false)
  const [user, setUser] = useState<UserInfo>({  nickname: '', id: '', refresh_token:'', access_token:'' })

  const router = useRouter()

  useEffect(() => {
    const userInfo = localStorage.getItem('user_info')

    if (!userInfo) {
      if (window.location.pathname != '/user/register') {
        router.push('/user/login')
        return
      } else {
        router.push('/user/register')
        return
      }
    } else {
      const user: UserInfo = JSON.parse(userInfo)
      if (user) {
        setUser({
          nickname: user.nickname,
          id: user.id,
          access_token: user.access_token,
          refresh_token: user.refresh_token,
        })
      }
      setAuthenticated(true)
    }
  }, [authenticated])

  return (
    <AuthContext.Provider
      value={{
        authenticated: authenticated,
        setAuthenticated: setAuthenticated,
        user: user,
        setUser: setUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export default AuthContextProvider
