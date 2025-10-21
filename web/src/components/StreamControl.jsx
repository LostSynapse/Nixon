import React from 'react';
import { Power, Mic, Waves, Users, ArrowUp } from 'lucide-react';

const StreamControl = ({ title, status, onToggle, listeners, peak }) => (
    <div className="bg-gray-800 p-4 rounded-lg shadow-md flex flex-col items-center justify-between">
      <div className="flex items-center justify-between w-full mb-3">
        <div className="flex items-center">
            {title.includes('SRT') ? <Waves className="w-6 h-6 mr-2 text-cyan-400" /> : <Mic className="w-6 h-6 mr-2 text-orange-400" />}
            <h3 className="text-lg font-semibold">{title}</h3>
        </div>
        {title.includes('Icecast') && status && (<div className="flex items-center text-xs text-gray-400"><Users size={14} className="mr-1"/> {listeners}<ArrowUp size={14} className="ml-2 mr-1"/> {peak}</div>)}
      </div>
      <div className={`text-sm font-bold px-2 py-1 rounded-full mb-4 ${status ? 'bg-red-500 text-white' : 'bg-gray-600 text-gray-300'}`}>{status ? 'LIVE' : 'OFFLINE'}</div>
      <button onClick={onToggle} className={`w-full flex items-center justify-center p-3 rounded-lg font-bold text-white transition-all ${status ? 'bg-red-600 hover:bg-red-700' : 'bg-green-600 hover:bg-green-700'}`}><Power className="w-5 h-5 mr-2" />{status ? 'Stop Stream' : 'Go Live'}</button>
    </div>
);

export default StreamControl;

