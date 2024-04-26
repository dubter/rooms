import React, { useState, useRef, useContext, useEffect } from 'react';
import ChatBody from '../../components/chat_body';
import { WebsocketContext } from '../../modules/websocket_provider';
import { useRouter } from 'next/router';
import { API_URL } from '../../constants';
import autosize from 'autosize';
import { AuthContext } from '../../modules/auth_provider';

export type Message = {
    content: string;
    user_id: string;
    nickname: string;
    room_id: string;
    time_created: string;
    type: 'recv' | 'self';
};

const Index = () => {
    const [messages, setMessage] = useState<Array<Message>>([]);
    const textarea = useRef<HTMLTextAreaElement>(null);
    const bottomRef = useRef<HTMLDivElement>(null); // Ref for the bottom of the page
    const { conn } = useContext(WebsocketContext);
    const [users, setUsers] = useState<Array<{ nickname: string }>>([]);
    const { user } = useContext(AuthContext); // Assuming you have a logout function in your AuthContext
    const router = useRouter();
    const { roomName, roomId } = router.query; // Accessing roomName query parameter

    useEffect(() => {
        if (conn === null) {
            router.push('/');
            return;
        }

        async function getUsers() {
            try {
                const authToken = `Bearer ${user.access_token}`;
                const res = await fetch(`${API_URL}/chat/rooms/${roomId}/clients`, {
                    method: 'GET',
                    headers: { 'Content-Type': 'application/json', Authorization: `${authToken}` },
                });
                const data = await res.json();
                setUsers(data);
            } catch (e) {
                console.error(e);
            }
        }
        getUsers();
    }, []);

    useEffect(() => {
        if (textarea.current) {
            autosize(textarea.current);
        }

        if (conn === null) {
            router.push('/');
            return;
        }

        conn.onmessage = (message) => {
            const m = JSON.parse(message.data);

            if (!Array.isArray(m)) {
                if (m.content === 'joined the room') {
                    setUsers([...users, { nickname: m.nickname }]);
                }

                if (m.content === 'left the room') {
                    const deleteUser = users.filter((user) => user.nickname !== m.nickname);
                    setUsers([...deleteUser]);
                    setMessage([...messages, m]);
                    return;
                }

                user?.nickname === m.nickname ? (m.type = 'self') : (m.type = 'recv');
                setMessage([...messages, m]);
            } else {
                setMessage([...m.reverse()]);
            }
        };

        conn.onclose = () => {};
        conn.onerror = () => {};
        conn.onopen = () => {};

        // Scroll to bottom when messages change
        if (bottomRef.current) {
            bottomRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [textarea, messages, conn, users]);

    const sendMessage = () => {
        if (!textarea.current?.value) return;
        if (conn === null) {
            router.push('/');
            return;
        }

        conn.send(textarea.current.value);
        textarea.current.value = '';
    };

    const handleMenu = () => {
        router.push('/'); // Redirect to the home page
    };

    return (
        <div className="flex flex-col w-full">
            <div className="bg-grey p-4 rounded-md mb-4 sticky top-0 z-10">
                <h2 className="text-lg text-center">{roomName}</h2>
            </div>
            <div className="flex-grow overflow-y-auto p-4 md:mx-6 mb-14">
                <ChatBody data={messages} />
                {/* Ref for scrolling to bottom */}
                <div ref={bottomRef}></div>
            </div>
            <div className="fixed bottom-0 mt-4 w-full">
                <div className="flex md:flex-row px-4 py-2 bg-grey md:mx-4 rounded-md">
                    <div className="flex items-center">
                        <button className="p-2 rounded-md bg-blue text-white" onClick={handleMenu}>
                            Menu
                        </button>
                    </div>
                    <div className="flex w-full mr-4 rounded-md border border-blue" style={{ marginLeft: '8px' }}> {/* Added style for margin top */}
                        <textarea
                            ref={textarea}
                            placeholder="type your message here"
                            className="w-full h-10 p-2 rounded-md focus:outline-none"
                            style={{ resize: 'none' }}
                        />
                    </div>
                    <div className="flex items-center">
                        <button className="p-2 rounded-md bg-blue text-white" onClick={sendMessage}>
                            Send
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Index;