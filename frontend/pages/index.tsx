import React, { useState, useEffect, useContext } from 'react';
import axios from 'axios';
import { API_URL, WEBSOCKET_URL } from '../constants';
import { AuthContext, UserInfo } from '../modules/auth_provider';
import { WebsocketContext } from '../modules/websocket_provider';
import { useRouter } from 'next/router';

const index = () => {
  const [rooms, setRooms] = useState<{ id: string; name: string }[]>([])
  const [roomName, setRoomName] = useState('')
  const { user, setUser } = useContext(AuthContext);
  const { setConn } = useContext(WebsocketContext);
  const router = useRouter();

  useEffect(() => {
    const userData = localStorage.getItem('user_info');
    if (userData) {
      const parsedUserData: UserInfo = JSON.parse(userData);
      setUser(parsedUserData);
    } else {
      router.push('/user/login');
    }
  }, []);

  useEffect(() => {
    getRooms();
  }, [user]); // Fetch rooms whenever user changes

  const getRooms = () => {
    console.log("user: ", user);
    axios.get(`${API_URL}/chat/rooms`, {
      headers: { 'Authorization': `Bearer ${user.access_token}` }
    })
        .then(response => {
          setRooms(response.data);
        })
        .catch(async error => {
          if (error.response && error.response.status === 401) {
            try {
              const response = await axios.post(`${API_URL}/user/refresh`, {
                refresh_token: user.refresh_token,
              });
              if (response.data) {
                const userData: UserInfo = {
                  nickname: response.data.nickname,
                  id: response.data.user_id,
                  access_token: response.data.access_token,
                  refresh_token: response.data.refresh_token,
                };
                localStorage.setItem('user_info', JSON.stringify(userData));
                router.push('/');
              } else {
                router.push('/user/login');
              }
            } catch (err) {
              console.log(err);
            }
          } else {
            console.log(error);
          }
        });
  };

  const submitHandler = (e: React.SyntheticEvent) => {
    e.preventDefault();
    axios.post(`${API_URL}/chat/rooms`, {
      name: roomName,
    }, {
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${user.access_token}`,
      },
    })
        .then(response => {
          getRooms()
        })
        .catch(async error => {
          if (error.response && error.response.status === 401) {
            try {
              const response = await axios.post(`${API_URL}/user/refresh`, {
                refresh_token: user.refresh_token,
              });
              if (response.data) {
                const userData: UserInfo = {
                  nickname: response.data.nickname,
                  id: response.data.user_id,
                  access_token: response.data.access_token,
                  refresh_token: response.data.refresh_token,
                };
                localStorage.setItem('user_info', JSON.stringify(userData));
                router.push('/');
              } else {
                router.push('/user/login');
              }
            } catch (err) {
              console.log(err);
            }
          } else {
            console.log(error);
          }
        });
  };

  const authToken = `Bearer ${user.access_token}`;

  const joinRoom = (roomId:string, roomName: string) => {
    const ws = new WebSocket(
        `${WEBSOCKET_URL}/chat/rooms/${roomId}?access_token=${authToken}`
    );
    if (ws.OPEN) {
      setConn(ws);
      router.push({
        pathname: '/app',
        query: { roomName: roomName, roomId: roomId } // Pass roomName as a query parameter
      });
      return
    } else {
      router.push('/user/login');
      return
    }
  }

  return (
      <>
        <div className='my-8 px-4 md:mx-32 w-full h-full'>
          <div className='flex justify-center mt-3 p-5'>
            <input
                type='text'
                className='border border-grey p-2 rounded-md focus:outline-none focus:border-blue'
                placeholder='room name'
                value={roomName}
                onChange={(e) => setRoomName(e.target.value)}
            />
            <button
                className='bg-blue border text-white rounded-md p-2 md:ml-4'
                onClick={submitHandler}
            >
              create room
            </button>
          </div>
          <div className='mt-6'>
            <div className='font-bold'>Available Rooms</div>
            <div className='grid grid-cols-1 md:grid-cols-5 gap-4 mt-6'>
              {rooms && rooms.map((room, index) => (
                  <div
                      key={index}
                      className='border border-blue p-4 flex items-center rounded-md w-full'
                  >
                    <div className='w-full'>
                      <div className='text-sm'>room</div>
                      <div className='text-blue font-bold text-lg'>{room.name}</div>
                    </div>
                    <div className=''>
                      <button
                          className='px-4 text-white bg-blue rounded-md'
                          onClick={() => joinRoom(room.id, room.name)}
                      >
                        join
                      </button>
                    </div>
                  </div>
              ))}
            </div>
          </div>
        </div>
      </>
  );
};

export default index;
