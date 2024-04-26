import React from 'react';
import { Message } from '../pages/app';

const ChatBody = ({ data }: { data: Array<Message> }) => {
    return (
        <>
            {data.map((message: Message, index: number) => {
                const messageDate = new Date(message.time_created).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }); // Convert time_created to time string

                if (message.type === 'self') {
                    return (
                        <div className='flex flex-col mt-2 w-full text-right justify-end' key={index}>
                            <div className='text-sm'>{message.nickname}</div>
                            <div>
                                <div className='bg-blue text-white px-4 py-1 rounded-md inline-block mt-1'>
                                    {message.content}
                                </div>
                                <div className='text-xs text-gray-500 mt-1'>{messageDate}</div> {/* Display message time */}
                            </div>
                        </div>
                    );
                } else {
                    return (
                        <div className='mt-2' key={index}>
                            <div className='text-sm'>{message.nickname}</div>
                            <div>
                                <div className='bg-grey text-dark-secondary px-4 py-1 rounded-md inline-block mt-1'>
                                    {message.content}
                                </div>
                                <div className='text-xs text-gray-500 mt-1'>{messageDate}</div> {/* Display message time */}
                            </div>
                        </div>
                    );
                }
            })}
        </>
    );
};

export default ChatBody;