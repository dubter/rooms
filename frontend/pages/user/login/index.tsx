import React, { useState, useContext, useEffect } from 'react'
import { API_URL } from '../../../constants'
import { useRouter } from 'next/router'
import { AuthContext, UserInfo } from '../../../modules/auth_provider'

const index = () => {
  const [nickname, setNickname] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null);

  const router = useRouter()

  const submitHandler = async (e: React.SyntheticEvent) => {
    e.preventDefault()

    try {
      const res = await fetch(`${API_URL}/user/login`, {
        method: "POST",
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ nickname, password }),
      })

      const data = await res.json()
      if (res.ok) {
        const user: UserInfo = {
          nickname: data.nickname,
          id: data.user_id,
          access_token: data.access_token,
          refresh_token: data.refresh_token,
        }

        console.log(user)

        localStorage.setItem('user_info', JSON.stringify(user))
        return router.push('/')
      } else {
      const errorMessage = data.message
      setError(errorMessage);
    }
    } catch (err) {
      console.log(err)
    }
  }

  const handleRegisterClick = () => {
    router.push('/user/register')
  }

  return (
      <div className='flex items-center justify-center min-w-full min-h-screen'>
        <form className='flex flex-col md:w-1/5'>
          <div className='text-3xl font-bold text-center'>
            <span className='text-blue'>Login</span>
          </div>
          <input
              placeholder='nickname'
              className='p-3 mt-8 rounded-md border-2 border-grey focus:outline-none focus:border-blue'
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
          />
          <input
              type='password'
              placeholder='password'
              className='p-3 mt-4 rounded-md border-2 border-grey focus:outline-none focus:border-blue'
              value={password}
              onChange={(e) => setPassword(e.target.value)}
          />
          <button
              className='p-3 mt-6 rounded-md bg-blue font-bold text-white'
              type='submit'
              onClick={submitHandler}
          >
            login
          </button>
          <button
              className='p-3 mt-2 rounded-md border-2 border-blue font-bold text-blue'
              type='button'
              onClick={handleRegisterClick}
          >
            Register
          </button>
          {error && (
              <div className="mt-4 bg-red-200 text-red-700 rounded-md border border-red-500 overflow-hidden">
                <div className="flex items-center px-4 py-2">
                  <span>{error}</span>
                </div>
              </div>
          )}
        </form>
      </div>
  )
}

export default index
