import React, { useState } from 'react';
import { API_URL } from '../../../constants';
import { useRouter } from 'next/router';

const Index = () => {
    const [nickname, setNickname] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState<string | null>(null);

    const router = useRouter();

    const submitHandler = async (e: React.SyntheticEvent) => {
        e.preventDefault();

        try {
            const res = await fetch(`${API_URL}/user/register`, {
                method: "POST",
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ nickname, password }),
            });

            if (res.ok) {
                return router.push('/user/login');
            } else {
                const data = await res.json()
                const errorMessage = data.message
                setError(errorMessage);
            }
        } catch (err) {
            console.error(err);
            setError('Network error occurred. Please try again later.');
        }
    };

    const handleBackClick = () => {
        router.push('/user/login');
    };

    return (
        <div className='flex items-center justify-center min-w-full min-h-screen'>
            <form className='flex flex-col md:w-1/5'>
                <div className='text-3xl font-bold text-center'>
                    <span className='text-blue'>Registration</span>
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
                    register
                </button>
                <button
                    className='p-3 mt-2 rounded-md border-2 border-blue font-bold text-blue'
                    type='button'
                    onClick={handleBackClick}
                >
                    Back
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
    );
};

export default Index;
